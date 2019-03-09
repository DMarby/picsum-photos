package file_test

import (
	"reflect"

	"github.com/DMarby/picsum-photos/database"
	"github.com/DMarby/picsum-photos/database/file"

	"testing"
)

var image = database.Image{
	ID:     "1",
	Author: "John Doe",
	URL:    "https://picsum.photos",
	Width:  300,
	Height: 400,
}

func TestFile(t *testing.T) {
	provider, err := file.New("../../test/fixtures/file/metadata.json")
	if err != nil {
		t.Fatal(err)
	}

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

func TestMissingMetadata(t *testing.T) {
	_, err := file.New("")
	if err == nil {
		t.FailNow()
	}
}

func TestInvalidJson(t *testing.T) {
	_, err := file.New("../../test/fixtures/file/invalid_metadata.json")
	if err == nil {
		t.FailNow()
	}
}
