// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package s3 provides simplified aws s3 functionality in order to reduce the
// amount of duplicated s3 code we have across all of our Go projects
package s3

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

type client struct {
	svc *s3.S3
	cfg *aws.Config
}

// ===================================================================
// Creating a new API client
// ===================================================================

// New produces a new s3 client.  The basic usage is as follows:
//
// 	s, err := s3.New()
//
// 	s, err := s3.New(s3.WithService(s3crypto.NewDecryptionClient())
//
// 	cfg, err := external.LoadDefaultAWSConfig(
//   	external.WithSharedConfigProfile("exampleProfile"),
// 	)
// 	s, err := s3.New(s3.WithConfig(cfg))
func New(options ...func(*client) error) (*client, error) {
	s := client{}

	for _, option := range options {
		option(&s)
	}

	return &s, nil
}

// Options ===============================================
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis

// WithService allows passing in an already established service object
func WithService(svc *s3.S3) func(*client) error {
	return func(s *client) error {
		s.svc = svc
		return nil
	}
}

// WithConfig allows passing in an already established configuration object
func WithConfig(cfg *aws.Config) func(*client) error {
	return func(s *client) error {
		s.cfg = cfg
		return nil
	}
}

// ===================================================================
// Methods
// ===================================================================

// GetObject will return the contents of the requested item
func (s *client) ListFiles(url string) ([]string, error) {
	svc, err := s.getSvc()
	files := make([]string, 0)
	if err != nil {
		return files, err
	}
	bucket, path, err := s.parseUrl(url, 1)
	if err != nil {
		return files, err
	}
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket), // Required
		Prefix: aws.String(path),
	}
	req, resp := svc.ListObjectsRequest(params)
	if err := req.Send(); err != nil {
		return files, err
	}
	for _, value := range resp.Contents {
		if *value.Size != int64(0) {
			files = append(files, *value.Key)
		}
	}
	return files, nil
}

// GetObject will return the contents of the requested item
func (s *client) GetObject(url string) (string, error) {
	svc, err := s.getSvc()
	if err != nil {
		return "", err
	}

	bucket, path, err := s.parseUrl(url, 2)
	if err != nil {
		return "", err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}

	req, resp := svc.GetObjectRequest(input)
	retry := false
	if err := req.Send(); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			// Some errors are re-tryable
			switch aerr.Code() {
			case "RequestTimeout", "ServiceUnavailable", "SlowDown":
				retry = true
			}
		}

		return "", requestError{
			baseError{cause: errors.Wrap(err, "get_object")},
			retry,
			url,
		}

	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", requestError{
			baseError{cause: errors.Wrap(err, "read_result_body")},
			retry,
			url,
		}
	}

	return string(b), nil
}

func (s *client) PutObject(url string, data string) error {
	return s.PutObjectStream(url, strings.NewReader(data))
}

func (s *client) PutObjectStream(url string, r io.Reader) error {
	svc, err := s.getSvc()
	if err != nil {
		return err
	}

	bucket, path, err := s.parseUrl(url, 2)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(r),
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		//		ServerSideEncryption: s3.ServerSideEncryptionAes256,
		//		Tagging:              aws.String("key1=value1&key2=value2"),
	}

	req, _ := svc.PutObjectRequest(input)
	retry := false
	if err = req.Send(); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			// Some errors are re-tryable
			switch aerr.Code() {
			case "RequestTimeout", "ServiceUnavailable", "SlowDown":
				retry = true
			}
		}

		return requestError{
			baseError{cause: errors.Wrap(err, "put_object")},
			retry,
			url,
		}

	}

	return nil
}

func (s *client) DeleteObject(url string) error {
	svc, err := s.getSvc()
	if err != nil {
		return err
	}

	bucket, path, err := s.parseUrl(url, 2)
	if err != nil {
		return err
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		//		ServerSideEncryption: s3.ServerSideEncryptionAes256,
		//		Tagging:              aws.String("key1=value1&key2=value2"),
	}

	req, _ := svc.DeleteObjectRequest(input)
	retry := false
	if err = req.Send(); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			// Some errors are re-tryable
			switch aerr.Code() {
			case "RequestTimeout", "ServiceUnavailable", "SlowDown":
				retry = true
			}
		}

		return requestError{
			baseError{cause: errors.Wrap(err, "put_object")},
			retry,
			url,
		}

	}

	return nil
}
