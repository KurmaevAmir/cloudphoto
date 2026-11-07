package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func cmdUpload(ctx context.Context, st *Storage, cmd *Command) error {
	entries, err := os.ReadDir(cmd.Path)
	if err != nil {
		return fmt.Errorf("Каталог %s недоступен: %w", cmd.Path, err)
	}

	var photos []string
	for _, e := range entries {
		name := e.Name()
		if !isPhotoFile(name) {
			continue
		}
		info, err := os.Stat(filepath.Join(cmd.Path, name))
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		if len(photoKey(cmd.Album, name)) > maxKeyLen {
			return fmt.Errorf("Слишком длинное имя файла: %s", name)
		}
		photos = append(photos, name)
	}
	if len(photos) == 0 {
		return fmt.Errorf("В каталоге %s нет фотографий", cmd.Path)
	}

	if err := st.CreateAlbum(ctx, cmd.Album); err != nil {
		return err
	}
	for _, name := range photos {
		if err := st.UploadPhoto(ctx, cmd.Album, name, filepath.Join(cmd.Path, name)); err != nil {
			return err
		}
	}
	return nil
}
