// +build integration

package spaces_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/DMarby/picsum-photos/database"
	"github.com/DMarby/picsum-photos/storage/spaces"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"testing"
)

func TestSpaces(t *testing.T) {
	var (
		space     = os.Getenv("PICSUM_SPACE")
		region    = os.Getenv("PICSUM_SPACES_REGION")
		accessKey = os.Getenv("PICSUM_SPACES_ACCESS_KEY")
		secretKey = os.Getenv("PICSUM_SPACES_SECRET_KEY")
	)

	provider, err := spaces.New(
		space,
		region,
		accessKey,
		secretKey,
	)

	if err != nil {
		t.Fatal(err)
	}

	fixture, _ := ioutil.ReadFile("../../test/fixtures/fixture.jpg")

	// Upload a fixture to the bucket
	spacesSession := session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(fmt.Sprintf("https://%s.digitaloceanspaces.com", region)),
		Region:      aws.String("us-east-1"), // Needs to be us-east-1 for Spaces, or it'll fail
	})
	spaces := s3.New(spacesSession)
	object := s3.PutObjectInput{
		Bucket: &space,
		Key:    aws.String("/1.jpg"),
		Body:   bytes.NewReader(fixture),
	}
	_, err = spaces.PutObject(&object)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Get an image by id", func(t *testing.T) {
		buf, err := provider.Get("1")
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(buf, fixture) {
			t.Error("image data doesn't match")
		}
	})

	t.Run("Returns error on a nonexistant image", func(t *testing.T) {
		_, err := provider.Get("nonexistant")
		if err != database.ErrNotFound {
			t.FailNow()
		}
	})

	// Cleanup
	delObject := s3.DeleteObjectInput{
		Bucket: &space,
		Key:    aws.String("/1.jpg"),
	}
	_, err = spaces.DeleteObject(&delObject)
}

func TestNew(t *testing.T) {
	_, err := spaces.New("", "", "", "")
	if err == nil {
		t.Fatal("no error")
	}
}
