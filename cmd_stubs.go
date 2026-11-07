package main

import (
	"context"
	"errors"
)

var errNotImplemented = errors.New("Команда не реализована")

func cmdMksite(ctx context.Context, st *Storage) error { return errNotImplemented }
