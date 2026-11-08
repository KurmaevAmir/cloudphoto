package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	endpoint = "https://storage.yandexcloud.net"
	region   = "ru-central1"
)

type Storage struct {
	client *s3.Client
	bucket string
}

func newStorage(cfg Config) *Storage {
	return newStorageAt(cfg, endpoint)
}

func newStorageAt(cfg Config, url string) *Storage {
	creds := aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
		return aws.Credentials{AccessKeyID: cfg.AccessKey, SecretAccessKey: cfg.SecretKey}, nil
	})
	client := s3.New(s3.Options{
		Region:                     region,
		BaseEndpoint:               aws.String(url),
		Credentials:                creds,
		UsePathStyle:               true,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
	return &Storage{client: client, bucket: cfg.Bucket}
}

func isNotFound(err error) bool {
	var re *awshttp.ResponseError
	return errors.As(err, &re) && re.HTTPStatusCode() == 404
}

func (s *Storage) ListAlbums(ctx context.Context) ([]string, error) {
	var albums []string
	p := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(photosPrefix),
		Delimiter: aws.String("/"),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить список альбомов: %w", err)
		}
		for _, cp := range page.CommonPrefixes {
			albums = appendAlbum(albums, aws.ToString(cp.Prefix))
		}
		for _, obj := range page.Contents {
			albums = appendAlbum(albums, aws.ToString(obj.Key))
		}
	}
	sort.Strings(albums)
	return slices.Compact(albums), nil
}

func appendAlbum(albums []string, key string) []string {
	name, ok := strings.CutSuffix(strings.TrimPrefix(key, photosPrefix), "/")
	if !ok || name == "" || strings.Contains(name, "/") {
		return albums
	}
	return append(albums, name)
}

func (s *Storage) AlbumExists(ctx context.Context, album string) (bool, error) {
	out, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucket),
		Prefix:  aws.String(albumPrefix(album)),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("не удалось проверить альбом %q: %w", album, err)
	}
	return len(out.Contents) > 0, nil
}

func (s *Storage) CreateAlbum(ctx context.Context, album string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(albumPrefix(album)),
		Body:   strings.NewReader(""),
	})
	if err != nil {
		return fmt.Errorf("не удалось создать альбом %q: %w", album, err)
	}
	return nil
}

func (s *Storage) ListPhotos(ctx context.Context, album string) ([]string, error) {
	prefix := albumPrefix(album)
	var photos []string
	p := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить список фотографий: %w", err)
		}
		for _, obj := range page.Contents {
			name := strings.TrimPrefix(aws.ToString(obj.Key), prefix)
			if name != "" { // пропускаем сам маркер альбома
				photos = append(photos, name)
			}
		}
	}
	sort.Strings(photos)
	return photos, nil
}

func (s *Storage) PhotoExists(ctx context.Context, album, photo string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(photoKey(album, photo)),
	})
	if err == nil {
		return true, nil
	}
	if isNotFound(err) {
		return false, nil
	}
	return false, fmt.Errorf("не удалось проверить фотографию %q: %w", photo, err)
}

func (s *Storage) UploadPhoto(ctx context.Context, album, photo, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл %q: %w", localPath, err)
	}
	defer f.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(photoKey(album, photo)),
		Body:        f,
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("не удалось отправить фотографию %q: %w", photo, err)
	}
	return nil
}

func (s *Storage) DownloadPhoto(ctx context.Context, album, photo string, w io.Writer) error {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(photoKey(album, photo)),
	})
	if err != nil {
		return fmt.Errorf("не удалось загрузить фотографию %q: %w", photo, err)
	}
	defer out.Body.Close()
	if _, err := io.Copy(w, out.Body); err != nil {
		return fmt.Errorf("не удалось загрузить фотографию %q: %w", photo, err)
	}
	return nil
}

func (s *Storage) DeletePrefix(ctx context.Context, prefix string) error {
	p := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("не удалось получить список объектов: %w", err)
		}
		for _, obj := range page.Contents {
			if err := s.DeleteKey(ctx, aws.ToString(obj.Key)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Storage) DeleteKey(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("не удалось удалить объект %q: %w", key, err)
	}
	return nil
}

func (s *Storage) PutHTML(ctx context.Context, key string, html []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(html)),
		ContentType: aws.String("text/html; charset=utf-8"),
	})
	if err != nil {
		return fmt.Errorf("не удалось сохранить страницу %q: %w", key, err)
	}
	return nil
}

func (s *Storage) ListRootKeys(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	p := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить список объектов: %w", err)
		}
		for _, obj := range page.Contents {
			keys = append(keys, aws.ToString(obj.Key))
		}
	}
	return keys, nil
}

func (s *Storage) ConfigureWebsite(ctx context.Context) error {
	_, err := s.client.PutBucketAcl(ctx, &s3.PutBucketAclInput{
		Bucket: aws.String(s.bucket),
		ACL:    types.BucketCannedACLPublicRead,
	}, withContentMD5)
	if err != nil {
		return fmt.Errorf("не удалось открыть публичный доступ к бакету: %w", err)
	}

	_, err = s.client.PutBucketWebsite(ctx, &s3.PutBucketWebsiteInput{
		Bucket: aws.String(s.bucket),
		WebsiteConfiguration: &types.WebsiteConfiguration{
			IndexDocument: &types.IndexDocument{Suffix: aws.String("index.html")},
			ErrorDocument: &types.ErrorDocument{Key: aws.String("error.html")},
		},
	}, withContentMD5)
	if err != nil {
		return fmt.Errorf("не удалось настроить хостинг сайта: %w", err)
	}
	return nil
}
