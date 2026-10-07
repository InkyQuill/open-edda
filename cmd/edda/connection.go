package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/InkyQuill/open-edda/project"

	"golang.org/x/term"
)

type connection struct {
	SessionID        string `json:"sessionID,omitempty"`
	RefreshToken     string `json:"refreshToken,omitempty"`
	RefreshExpiresAt int64  `json:"refreshExpiresAt,omitempty"`
	Server           string `json:"server"`
	Token            string `json:"token,omitempty"`
}

func connectionPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "open-edda", "client.json"), nil
}
func readConnection() (connection, error) {
	var c connection
	p, err := connectionPath()
	if err != nil {
		return c, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(data, &c)
	return c, err
}
func writePrivateJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	limit := 128 << 20
	if filepath.Base(path) == "checkout.json" {
		limit = 64 << 20
	}
	if len(data) > limit {
		return fmt.Errorf("local recovery metadata exceeds %d MiB", limit>>20)
	}
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, ".write-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}
	return syncDirectory(dir)
}
func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func resolveConnection(server string) (connection, error) {
	saved, err := readConnection()
	if err != nil {
		return saved, err
	}
	if server == "" {
		server = os.Getenv("OPEN_EDDA_URL")
	}
	if server == "" {
		server = saved.Server
	}
	server = strings.TrimRight(server, "/")
	token := os.Getenv("OPEN_EDDA_TOKEN")
	if token == "" && server == strings.TrimRight(saved.Server, "/") {
		token = saved.Token
	}
	if _, err := newImportClient(server, "_", token); err != nil {
		return connection{}, fmt.Errorf("connection unavailable; run edda login first: %w", err)
	}
	return connection{Server: server, Token: token}, nil
}
func runLogin(args []string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("login", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.Usage = func() {
		fmt.Fprintln(output, "Usage: edda login [--server URL] [--email EMAIL] [--password-stdin]\n\nWithout --password-stdin, login asks for the server, email and a hidden password.\nFor scripts, provide --email and pipe the password with --password-stdin.")
		flags.PrintDefaults()
	}
	server := flags.String("server", os.Getenv("OPEN_EDDA_URL"), "server URL")
	email := flags.String("email", "", "account email")
	stdin := flags.Bool("password-stdin", false, "read password from standard input")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("login takes no positional arguments; use edda login --help")
	}
	if *server == "" {
		saved, err := readConnection()
		if err != nil {
			return err
		}
		*server = saved.Server
	}
	file, terminal := input.(*os.File)
	terminal = terminal && term.IsTerminal(int(file.Fd()))
	if !*stdin && !terminal {
		return errors.New("interactive login needs a terminal; run edda login in a terminal, or pipe a password with --password-stdin --email EMAIL")
	}
	if *stdin && terminal {
		return errors.New("--password-stdin is for piped input; run edda login without it to enter a hidden password")
	}
	if !*stdin {
		explicitServer := false
		flags.Visit(func(f *flag.Flag) {
			if f.Name == "server" {
				explicitServer = true
			}
		})
		if !explicitServer {
			prompt := "Server URL"
			if *server != "" {
				prompt += " [" + *server + "]"
			}
			value, err := promptLine(input, output, prompt+": ")
			if err != nil {
				return err
			}
			if value != "" {
				*server = value
			}
		}
		if strings.TrimSpace(*email) == "" {
			value, err := promptLine(input, output, "Email: ")
			if err != nil {
				return err
			}
			*email = value
		}
	}
	*server = strings.TrimSpace(*server)
	*email = strings.TrimSpace(*email)
	if *email == "" {
		return errors.New("email is required; use --email EMAIL with --password-stdin")
	}

	client, err := newImportClient(*server, "_", "login")
	if err != nil {
		return err
	}
	var password []byte
	if *stdin {
		password, err = io.ReadAll(io.LimitReader(input, 4097))
		if len(password) > 4096 {
			return errors.New("password input too long")
		}
		password = bytes.TrimSuffix(bytes.TrimSuffix(password, []byte("\n")), []byte("\r"))
	} else {
		password, err = loginPassword(file, output)
	}
	if err != nil {
		return err
	}
	if len(password) == 0 {
		return errors.New("password must not be empty")
	}
	if len(password) > 4096 {
		return errors.New("password input too long")
	}
	body, err := json.Marshal(map[string]string{"email": *email, "password": string(password)})
	if err != nil {
		return err
	}
	client.root = strings.TrimRight(*server, "/") + "/api/"
	client.token = ""
	unlock, err := lockSession()
	if err != nil {
		return err
	}
	defer unlock()
	response, err := client.request(context.Background(), "POST", "auth/login", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var result connection
	if err = json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
		return err
	}
	if result.Token == "" {
		return errors.New("login response has no token")
	}
	path, err := connectionPath()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if err = writePrivateJSON(path, connection{SessionID: rand.Text(), Server: strings.TrimRight(*server, "/"), Token: result.Token, RefreshToken: result.RefreshToken, RefreshExpiresAt: result.RefreshExpiresAt}); err != nil {
		return err
	}
	fmt.Fprintf(output, "Connected to %s as %s.\nRun edda projects to see your projects.\n", strings.TrimRight(*server, "/"), *email)
	return nil
}
func runProjects(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("projects", flag.ContinueOnError)
	flags.SetOutput(output)
	machine := flags.Bool("json", false, "output JSON for scripts")
	flags.Usage = func() {
		fmt.Fprintln(output, "Usage: edda projects [--json]\n\nList your projects. Use --json for machine-readable output.")
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("projects takes no positional arguments; use edda projects --help")
	}

	c, err := resolveConnection("")
	if err != nil {
		return err
	}
	client, err := newImportClient(c.Server, "_", c.Token)
	if err != nil {
		return err
	}
	client.root = c.Server + "/api/"
	response, err := client.request(context.Background(), "GET", "projects", nil, 0)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, (8<<20)+1))
	if err != nil {
		return err
	}
	if len(data) > 8<<20 {
		return errors.New("project list response too large")
	}
	var projects []project.StoryProject
	if err := json.Unmarshal(data, &projects); err != nil {
		return fmt.Errorf("read project list: %w", err)
	}
	if *machine {
		return json.NewEncoder(output).Encode(json.RawMessage(data))
	}
	if len(projects) == 0 {
		_, err := fmt.Fprintln(output, "No projects yet. Create one with edda create --title \"My book\".")
		return err
	}
	table := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	fmt.Fprintln(table, "TITLE\tID")
	for _, p := range projects {
		title := strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return ' '
			}
			return r
		}, p.Title)
		fmt.Fprintf(table, "%s\t%s\n", title, p.ID)
	}
	return table.Flush()
}

func runLogout(output io.Writer) error {
	unlock, err := lockSession()
	if err != nil {
		return err
	}
	defer unlock()
	c, err := readConnection()
	if err != nil {
		return err
	}
	if c.RefreshToken != "" {
		client, err := newImportClient(c.Server, "_", c.Token)
		if err != nil {
			return err
		}
		client.root = c.Server + "/api/"
		body, err := json.Marshal(map[string]string{"refreshToken": c.RefreshToken})
		if err != nil {
			return err
		}
		response, err := client.request(context.Background(), "POST", "auth/logout", bytes.NewReader(body), int64(len(body)))
		if err != nil {
			return err
		}
		response.Body.Close()
	}
	c.SessionID = ""
	c.Token = ""
	c.RefreshToken = ""
	c.RefreshExpiresAt = 0
	path, err := connectionPath()
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintln(output, "No saved login.")
		return nil
	} else if err != nil {
		return err
	}
	if err = writePrivateJSON(path, c); err != nil {
		return err
	}
	fmt.Fprintln(output, "Saved login removed. Environment tokens are unchanged.")
	return nil
}

// Read one line without buffering ahead into the subsequent hidden password.
func promptLine(input io.Reader, output io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(output, prompt); err != nil {
		return "", err
	}
	var line []byte
	var b [1]byte
	for len(line) <= 4096 {
		n, err := input.Read(b[:])
		if n > 0 {
			if b[0] == '\n' {
				return strings.TrimSpace(string(line)), nil
			}
			line = append(line, b[0])
		}
		if err != nil {
			return "", fmt.Errorf("input cancelled: %w", err)
		}
	}
	return "", errors.New("input too long")
}

func loginPassword(input *os.File, output io.Writer) ([]byte, error) {
	fd := int(input.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, state)
	terminal := term.NewTerminal(struct {
		io.Reader
		io.Writer
	}{input, output}, "")
	password, err := terminal.ReadPassword("Password: ")
	if err != nil {
		return nil, fmt.Errorf("password input cancelled: %w", err)
	}
	return []byte(password), nil
}
