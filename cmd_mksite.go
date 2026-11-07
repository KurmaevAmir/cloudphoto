package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

var pages = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

type albumLink struct{ Name, Page string }
type photoItem struct{ Name, URL string }

var albumPageRe = regexp.MustCompile(`^album([1-9][0-9]*)\.html$`)

func cmdMksite(ctx context.Context, st *Storage) error {
	if err := st.ConfigureWebsite(ctx); err != nil {
		return err
	}
	albums, err := st.ListAlbums(ctx)
	if err != nil {
		return err
	}

	links := make([]albumLink, 0, len(albums))
	for i, album := range albums {
		page := fmt.Sprintf("album%d.html", i+1)
		names, err := st.ListPhotos(ctx, album)
		if err != nil {
			return err
		}
		photos := make([]photoItem, 0, len(names))
		for _, n := range names {
			photos = append(photos, photoItem{Name: n, URL: photoURL(album, n)})
		}
		html, err := render("album.html", map[string]any{"Name": album, "Photos": photos})
		if err != nil {
			return err
		}
		if err := st.PutHTML(ctx, page, html); err != nil {
			return err
		}
		links = append(links, albumLink{Name: album, Page: page})
	}

	html, err := render("index.html", map[string]any{"Albums": links})
	if err != nil {
		return err
	}
	if err := st.PutHTML(ctx, "index.html", html); err != nil {
		return err
	}
	html, err = render("error.html", nil)
	if err != nil {
		return err
	}
	if err := st.PutHTML(ctx, "error.html", html); err != nil {
		return err
	}

	keys, err := st.ListRootKeys(ctx, "album")
	if err != nil {
		return err
	}
	for _, key := range keys {
		m := albumPageRe.FindStringSubmatch(key)
		if m == nil {
			continue
		}
		if n, err := strconv.Atoi(m[1]); err == nil && n > len(albums) {
			if err := st.DeleteKey(ctx, key); err != nil {
				return err
			}
		}
	}
	return nil
}

func photoURL(album, photo string) string {
	return photosPrefix + escapeSegment(album) + "/" + escapeSegment(photo)
}

func escapeSegment(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), "+", "%2B")
}

func render(name string, data any) ([]byte, error) {
	var buf bytes.Buffer
	if err := pages.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, fmt.Errorf("ошибка шаблона %s: %w", name, err)
	}
	return buf.Bytes(), nil
}
