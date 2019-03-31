// +build integration

package postgresql_test

import (
	"reflect"

	"github.com/DMarby/picsum-photos/database"
	"github.com/DMarby/picsum-photos/database/postgresql"

	"testing"
)

var image = database.Image{
	ID:     "1",
	Author: "John Doe",
	URL:    "https://picsum.photos",
	Width:  300,
	Height: 400,
}

func TestPostgresql(t *testing.T) {
	provider, err := postgresql.New("postgresql://postgres@localhost/postgres")
	if err != nil {
		t.Fatal(err)
	}

	defer provider.Shutdown()

	t.Run("Get an image by id", func(t *testing.T) {
		buf, err := provider.Get("1")
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(buf, &image) {
			t.Error("image data doesn't match")
		}
	})

	t.Run("Returns error on a nonexistant image", func(t *testing.T) {
		_, err := provider.Get("2")
		if err == nil || err.Error() != database.ErrNotFound.Error() {
			t.FailNow()
		}
	})

	t.Run("Returns a random image", func(t *testing.T) {
		image, err := provider.GetRandom()
		if err != nil {
			t.Fatal(err)
		}

		if image != "1" {
			t.Error("wrong image")
		}
	})

	t.Run("Returns a list of images", func(t *testing.T) {
		images, err := provider.List()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(images, []database.Image{image}) {
			t.Error("image data doesn't match")
		}
	})
}

func TestNew(t *testing.T) {
	_, err := postgresql.New("")
	if err == nil {
		t.FailNow()
	}
}
