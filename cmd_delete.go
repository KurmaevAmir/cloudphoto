package main

import (
	"context"
	"fmt"
)

func cmdDelete(ctx context.Context, st *Storage, cmd *Command) error {
	if err := requireAlbum(ctx, st, cmd.Album); err != nil {
		return err
	}

	if !cmd.PhotoSet {
		return st.DeletePrefix(ctx, albumPrefix(cmd.Album))
	}

	ok, err := st.PhotoExists(ctx, cmd.Album, cmd.Photo)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("В альбоме %s нет фотографии с именем %s", cmd.Album, cmd.Photo)
	}
	return st.DeleteKey(ctx, photoKey(cmd.Album, cmd.Photo))
}
