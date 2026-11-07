package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func cmdDownload(ctx context.Context, st *Storage, cmd *Command) error {
	if err := checkWritableDir(cmd.Path); err != nil {
		return err
	}
	if err := requireAlbum(ctx, st, cmd.Album); err != nil {
		return err
	}
	photos, err := st.ListPhotos(ctx, cmd.Album)
	if err != nil {
		return err
	}
	for _, name := range photos {
		if !validName(name) {
			return fmt.Errorf("Недопустимое имя фотографии в хранилище: %q", name)
		}
		if err := downloadTo(ctx, st, cmd.Album, name, cmd.Path); err != nil {
			return err
		}
	}
	return nil
}

func checkWritableDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("Каталог %s недоступен", dir)
	}
	f, err := os.CreateTemp(dir, ".cloudphoto-check-*")
	if err != nil {
		return fmt.Errorf("Каталог %s недоступен для записи", dir)
	}
	f.Close()
	os.Remove(f.Name())
	return nil
}

func downloadTo(ctx context.Context, st *Storage, album, name, dir string) error {
	tmp, err := os.CreateTemp(dir, ".cloudphoto-*")
	if err != nil {
		return fmt.Errorf("Не удалось создать файл в %s: %w", dir, err)
	}
	defer os.Remove(tmp.Name())

	if err := st.DownloadPhoto(ctx, album, name, tmp); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("Не удалось сохранить %s: %w", name, err)
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return fmt.Errorf("Не удалось сохранить %s: %w", name, err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, name)); err != nil {
		return fmt.Errorf("Не удалось сохранить %s: %w", name, err)
	}
	return nil
}
