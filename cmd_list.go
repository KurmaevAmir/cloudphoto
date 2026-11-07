package main

import (
	"context"
	"fmt"
)

func errNoAlbum(album string) error {
	return fmt.Errorf("В фотоархиве нет альбома с именем %s", album)
}

func requireAlbum(ctx context.Context, st *Storage, album string) error {
	ok, err := st.AlbumExists(ctx, album)
	if err != nil {
		return err
	}
	if !ok {
		return errNoAlbum(album)
	}
	return nil
}

func cmdList(ctx context.Context, st *Storage, cmd *Command) error {
	if !cmd.AlbumSet {
		albums, err := st.ListAlbums(ctx)
		if err != nil {
			return err
		}
		for _, a := range albums {
			fmt.Println(a)
		}
		return nil
	}

	if err := requireAlbum(ctx, st, cmd.Album); err != nil {
		return err
	}
	photos, err := st.ListPhotos(ctx, cmd.Album)
	if err != nil {
		return err
	}
	for _, p := range photos {
		fmt.Println(p)
	}
	return nil
}
