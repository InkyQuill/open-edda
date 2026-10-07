package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/InkyQuill/open-edda/project"
	"io"
	"net/url"
	"text/tabwriter"
)

func runNetworkHistory(args []string, output io.Writer) error {
	root, rest := splitOptionalPath(args)
	flags := flag.NewFlagSet("history", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	machine := flags.Bool("json", false, "output JSON for scripts")
	cursor := flags.String("cursor", "", "older-page cursor")
	if err := flags.Parse(rest); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("history accepts one checkout")
	}
	state, err := readCheckout(root)
	if err != nil {
		return err
	}
	c, err := resolveConnection(state.Server)
	if err != nil {
		return err
	}
	client, err := newImportClient(c.Server, state.Base.ProjectID, c.Token)
	if err != nil {
		return err
	}
	response, err := client.request(context.Background(), "GET", "versions?cursor="+url.QueryEscape(*cursor), nil, 0)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var page project.VersionPage
	if err := json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&page); err != nil {
		return err
	}
	if *machine {
		return json.NewEncoder(output).Encode(page)
	}
	if len(page.Versions) == 0 {
		fmt.Fprintln(output, "No saved versions.")
		return nil
	}
	table := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	fmt.Fprintln(table, "VERSION\tSAVED AT\tNOTE")
	for _, v := range page.Versions {
		fmt.Fprintf(table, "%s\t%s\t%s\n", terminalText(v.ID), terminalText(v.CreatedAt), terminalText(v.Message))
	}
	if err := table.Flush(); err != nil {
		return err
	}
	if page.Next != "" {
		fmt.Fprintf(output, "More versions: edda history %q --cursor %s\n", root, page.Next)
	}
	return nil
}

// Restoring is a server publication. Local files/base stay unchanged until take
// performs its ordinary three-way reconciliation and creates recovery snapshots.
func runNetworkRestore(args []string, output io.Writer) error {
	root, rest := splitOptionalPath(args)
	flags := flag.NewFlagSet("restore", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	target := flags.String("version", "", "version ID to restore")
	if err := flags.Parse(rest); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: edda restore CHECKOUT --version ID")
	}
	if *target == "" && interactiveInput() {
		if err := runNetworkHistory([]string{root}, output); err != nil {
			return err
		}
	}
	if err := askValue(target, "Version ID (--version)", "", output); err != nil {
		return err
	}
	if _, err := readCheckout(root); err != nil {
		return err
	}
	unlock, err := lockCheckout(root)
	if err != nil {
		return err
	}
	defer unlock()
	state, err := readCheckout(root)
	if err != nil {
		return err
	}
	if state.Pending != nil || state.Update != "" || state.Move != nil {
		return errors.New("finish pending synchronization before restoring")
	}
	c, err := resolveConnection(state.Server)
	if err != nil {
		return err
	}
	client, err := newImportClient(c.Server, state.Base.ProjectID, c.Token)
	if err != nil {
		return err
	}
	ctx := context.Background()
	v, err := client.version(ctx, "versions/"+url.PathEscape(*target))
	if err != nil {
		return err
	}
	if err := validateCheckoutTree(v.Entries); err != nil {
		return err
	}
	// A deterministic key binds this restoration to its base and target. A retry
	// receives the same receipt even after the remote head advances.
	key := sha256.Sum256([]byte(state.Base.ID + "\x00" + v.ID))
	body, err := json.Marshal(map[string]string{"expectedVersion": state.Base.ID, "versionId": v.ID, "operationId": "restore_" + hex.EncodeToString(key[:])})
	if err != nil {
		return err
	}
	response, err := client.request(ctx, "POST", "restore", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var receipt project.ProjectVersion
	if err := json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&receipt); err != nil {
		return err
	}
	if receipt.ProjectID != state.Base.ProjectID || receipt.ParentID != state.Base.ID || !sameFiles(receipt.Entries, v.Entries) {
		return errors.New("restore receipt mismatch")
	}
	fmt.Fprintf(output, "Restored on server as %s. Local work is unchanged; run take to reconcile it.\n", receipt.ID)
	return nil
}
