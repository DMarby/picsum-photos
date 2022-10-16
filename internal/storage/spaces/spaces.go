package spaces

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/DMarby/picsum-photos/internal/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Provider implements a digitalocean spaces based image storage
type Provider struct {
	spaces *s3.S3
	space  string
}

// New returns a new Provider instance
func New(space, endpoint, accessKey, secretKey string, forcePathStyle bool) (*Provider, error) {
	spacesSession := session.New(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-east-1"), // Needs to be us-east-1 for Spaces, or it'll fail
		S3ForcePathStyle: aws.Bool(forcePathStyle),
	})

	spaces := s3.New(spacesSession)

	object := s3.GetObjectInput{
		Bucket: &space,
		Key:    aws.String("/"),
	}

	result, err := spaces.GetObject(&object)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return &Provider{
		spaces: spaces,
		space:  space,
	}, nil
}

// Get returns the image data for an image id
func (p *Provider) Get(ctx context.Context, id string) ([]byte, error) {
	object := s3.GetObjectInput{
		Bucket: &p.space,
		Key:    aws.String(fmt.Sprintf("%s.jpg", id)),
	}

	output, err := p.spaces.GetObjectWithContext(ctx, &object)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil, storage.ErrNotFound
		}

		return nil, err
	}
	defer output.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, output.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
