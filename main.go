package main

import (
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
	_ = cfg

	cmd, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, msgBadArgs)
		return 1
	}
	fmt.Fprintf(os.Stderr, "Команда не реализована", cmd.Name)
	return 0
}
