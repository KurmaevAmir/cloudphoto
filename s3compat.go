package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func withContentMD5(o *s3.Options) {
	o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
		_, _ = stack.Finalize.Remove("AWSChecksum:ComputeInputPayloadChecksum")
		return stack.Build.Add(middleware.BuildMiddlewareFunc("ContentMD5",
			func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (
				middleware.BuildOutput, middleware.Metadata, error,
			) {
				req, ok := in.Request.(*smithyhttp.Request)
				if !ok {
					return next.HandleBuild(ctx, in)
				}
				var body []byte
				if s := req.GetStream(); s != nil {
					data, err := io.ReadAll(s)
					if err != nil {
						return middleware.BuildOutput{}, middleware.Metadata{}, err
					}
					if req, err = req.SetStream(bytes.NewReader(data)); err != nil {
						return middleware.BuildOutput{}, middleware.Metadata{}, err
					}
					body = data
				}
				sum := md5.Sum(body)
				req.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(sum[:]))
				in.Request = req
				return next.HandleBuild(ctx, in)
			}), middleware.After)
	})
}
