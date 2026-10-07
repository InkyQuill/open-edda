package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/InkyQuill/open-edda/backup"
	"io"
)

func runBackup(command string, args []string, output io.Writer) error {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	db := flags.String("db", "", "live SQLite path")
	data := flags.String("data", "", "live object data root")
	source := flags.String("source", "", "backup directory")
	dest := flags.String("output", "", "new output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected argument")
	}
	var err error
	switch command {
	case "backup":
		if err := askValue(db, "SQLite database file (--db)", "", output); err != nil {
			return err
		}
		if err := askValue(data, "Object data directory (--data)", "", output); err != nil {
			return err
		}
		if err := askValue(dest, "New backup directory (--output)", "", output); err != nil {
			return err
		}
		if *db == "" || *data == "" || *dest == "" {
			return errors.New("usage: edda backup --db FILE --data ROOT --output NEW_DIRECTORY")
		}
		err = backup.Create(context.Background(), *db, *data, *dest)
	case "verify-backup":
		if err := askValue(source, "Backup directory (--source)", "", output); err != nil {
			return err
		}
		if *source == "" {
			return errors.New("usage: edda verify-backup --source DIRECTORY")
		}
		err = backup.Verify(context.Background(), *source)
	case "restore-backup":
		if err := askValue(source, "Backup directory (--source)", "", output); err != nil {
			return err
		}
		if err := askValue(dest, "New restored data directory (--output)", "", output); err != nil {
			return err
		}
		if *source == "" || *dest == "" {
			return errors.New("usage: edda restore-backup --source DIRECTORY --output NEW_DATA_ROOT")
		}
		err = backup.Restore(context.Background(), *source, *dest)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(output, "Backup operation completed and verified.")
	return nil
}
