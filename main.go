package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

const (
	msgNoEnv   = "Не заданы переменные окружения"
	msgBadArgs = "Ошибка в аргументах"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, msgNoEnv)
		return 1
	}

	cmd, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, msgBadArgs)
		return 1
	}

	ctx := context.Background()
	st := newStorage(cfg)
	if err := execute(ctx, st, cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func execute(ctx context.Context, st *Storage, cmd *Command) error {
	switch cmd.Name {
	case "list":
		return cmdList(ctx, st, cmd)
	case "upload":
		return cmdUpload(ctx, st, cmd)
	case "download":
		return cmdDownload(ctx, st, cmd)
	case "delete":
		return cmdDelete(ctx, st, cmd)
	case "mksite":
		return cmdMksite(ctx, st)
	}
	return errors.New(msgBadArgs)
}
