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

var secondImage = database.Image{
	ID:     "2",
	Author: "John Doe",
	URL:    "https://picsum.photos",
	Width:  300,
	Height: 400,
}

func TestFile(t *testing.T) {
	provider, err := file.New("../../test/fixtures/file/metadata_multiple.json")
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
		_, err := provider.Get("nonexistant")
		if err == nil || err.Error() != database.ErrNotFound.Error() {
			t.FailNow()
		}
	})

	t.Run("Returns a random image", func(t *testing.T) {
		image, err := provider.GetRandom()
		if err != nil {
			t.Fatal(err)
		}

		if image != "1" && image != "2" {
			t.Error("wrong image")
		}
	})

	t.Run("Returns a random based on the seed", func(t *testing.T) {
		image, err := provider.GetRandomWithSeed(0)
		if err != nil {
			t.Fatal(err)
		}

		if image != "1" {
			t.Error("wrong image")
		}
	})

	t.Run("Returns a list of all the images", func(t *testing.T) {
		images, err := provider.ListAll()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(images, []database.Image{image, secondImage}) {
			t.Error("image data doesn't match")
		}
	})

	t.Run("Returns a list of images", func(t *testing.T) {
		images, err := provider.List(1, 1)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(images, []database.Image{secondImage}) {
			t.Error("image data doesn't match")
		}
	})

	t.Run("Handles offset and limit larger then db", func(t *testing.T) {
		_, err := provider.List(10, 30)
		if err != nil {
			t.Fatal(err)
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
