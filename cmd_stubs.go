package main

import (
	"context"
	"errors"
)

var errNotImplemented = errors.New("Команда не реализована")

func cmdDelete(ctx context.Context, st *Storage, cmd *Command) error { return errNotImplemented }
func cmdMksite(ctx context.Context, st *Storage) error               { return errNotImplemented }
