package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"git.inkyquill.net/inky/writer/fileproject"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stderr)
		return fmt.Errorf("command is required")
	}

	switch args[0] {
	case "status":
		return runStatus(args[1:], stdout)
	case "init":
		return runInit(args[1:], stdout)
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runStatus(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root, flagArgs := splitOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}

	layout, err := fileproject.Scan(root)
	if err != nil {
		return err
	}

	if layout.Metadata != nil {
		fmt.Fprintf(stdout, "Project: %s (%s)\n", layout.Metadata.Title, layout.Metadata.ID)
	} else {
		fmt.Fprintln(stdout, "Project: uninitialized Edda folder")
	}
	fmt.Fprintf(stdout, "Root: %s\n", layout.Root)

	counts := fileproject.CountByKind(layout.Files)
	kinds := make([]fileproject.LayoutKind, 0, len(counts))
	for kind := range counts {
		kinds = append(kinds, kind)
	}
	sort.Slice(kinds, func(i, j int) bool { return kinds[i] < kinds[j] })
	fmt.Fprintln(stdout, "Files:")
	for _, kind := range kinds {
		fmt.Fprintf(stdout, "  %s: %d\n", kind, counts[kind])
	}
	if len(kinds) == 0 {
		fmt.Fprintln(stdout, "  none")
	}

	if len(layout.Warnings) > 0 {
		fmt.Fprintln(stdout, "Warnings:")
		for _, warning := range layout.Warnings {
			if warning.Path != "" {
				fmt.Fprintf(stdout, "  %s: %s (%s)\n", warning.Code, warning.Message, warning.Path)
			} else {
				fmt.Fprintf(stdout, "  %s: %s\n", warning.Code, warning.Message)
			}
		}
	}

	return nil
}

func runInit(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	title := flags.String("title", "", "project title")
	id := flags.String("id", "", "project id")
	serverURL := flags.String("server-url", "", "server URL")
	root, flagArgs := splitOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}

	metadata, err := fileproject.InitMetadata(root, fileproject.InitMetadataInput{
		ID:        *id,
		Title:     *title,
		ServerURL: *serverURL,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Initialized Edda project: %s (%s)\n", metadata.Title, metadata.ID)
	return nil
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  edda status [path]")
	fmt.Fprintln(output, "  edda init [path] --title \"Title\" [--id project-id] [--server-url URL]")
}

func splitOptionalPath(args []string) (string, []string) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return ".", args
	}
	return args[0], args[1:]
}
