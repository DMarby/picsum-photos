package file_test

import (
	"context"
	"io/ioutil"
	"reflect"

	"github.com/DMarby/picsum-photos/internal/storage/file"

	"testing"
)

func TestFile(t *testing.T) {
	provider, err := file.New("../../../test/fixtures/file")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Get an image by id", func(t *testing.T) {
		buf, err := provider.Get(context.Background(), "1")
		if err != nil {
			t.Fatal(err)
		}

		resultFixture, _ := ioutil.ReadFile("../../../test/fixtures/file/1.jpg")
		if !reflect.DeepEqual(buf, resultFixture) {
			t.Error("image data doesn't match")
		}
	})

	t.Run("Returns error on a nonexistant path", func(t *testing.T) {
		_, err := file.New("")
		if err == nil {
			t.FailNow()
		}
	})

	t.Run("Returns error on a nonexistant image", func(t *testing.T) {
		_, err := provider.Get(context.Background(), "nonexistant")
		if err == nil {
			t.FailNow()
		}
	})
}
