package main

import (
	"errors"
	"flag"
	"io"
	"strings"
	"unicode/utf8"
)

type Command struct {
	Name     string
	Album    string
	AlbumSet bool
	Path     string
	Photo    string
	PhotoSet bool
}

var errBadArgs = errors.New(msgBadArgs)

func parseArgs(args []string) (*Command, error) {
	if len(args) == 0 {
		return nil, errBadArgs
	}
	cmd := &Command{Name: args[0], Path: "."}

	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	switch cmd.Name {
	case "list":
		fs.StringVar(&cmd.Album, "album", "", "")
	case "upload", "download":
		fs.StringVar(&cmd.Album, "album", "", "")
		fs.StringVar(&cmd.Path, "path", ".", "")
	case "delete":
		fs.StringVar(&cmd.Album, "album", "", "")
		fs.StringVar(&cmd.Photo, "photo", "", "")
	case "mksite":

	default:
		return nil, errBadArgs
	}

	if err := fs.Parse(args[1:]); err != nil {
		return nil, errBadArgs
	}
	if fs.NArg() != 0 {
		return nil, errBadArgs
	}

	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	cmd.AlbumSet = set["album"]
	cmd.PhotoSet = set["photo"]

	if (cmd.Name == "upload" || cmd.Name == "download" || cmd.Name == "delete") && !cmd.AlbumSet {
		return nil, errBadArgs
	}
	if cmd.AlbumSet && !validName(cmd.Album) {
		return nil, errBadArgs
	}
	if cmd.PhotoSet && !validName(cmd.Photo) {
		return nil, errBadArgs
	}
	if set["path"] && cmd.Path == "" {
		return nil, errBadArgs
	}
	if cmd.PhotoSet && len(photoKey(cmd.Album, cmd.Photo)) > maxKeyLen {
		return nil, errBadArgs
	}
	if cmd.AlbumSet && len(albumPrefix(cmd.Album)) >= maxKeyLen {
		return nil, errBadArgs
	}
	return cmd, nil
}

func validName(s string) bool {
	return s != "" && s != "." && s != ".." && utf8.ValidString(s) && !strings.Contains(s, "/")
}
