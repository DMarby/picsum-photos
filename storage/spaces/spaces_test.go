// +build integration

package spaces_test

import (
	"io/ioutil"
	"os"
	"reflect"

	"github.com/DMarby/picsum-photos/database"
	"github.com/DMarby/picsum-photos/storage/spaces"

	"testing"
)

func TestSpaces(t *testing.T) {
	provider, err := spaces.New(
		os.Getenv("PICSUM_SPACE"),
		os.Getenv("PICSUM_SPACES_REGION"),
		os.Getenv("PICSUM_SPACES_ACCESS_KEY"),
		os.Getenv("PICSUM_SPACES_SECRET_KEY"),
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Run("Get an image by id", func(t *testing.T) {
		buf, err := provider.Get("1")
		if err != nil {
			t.Fatal(err)
		}

		resultFixture, _ := ioutil.ReadFile("../../test/fixtures/fixture.jpg")
		if !reflect.DeepEqual(buf, resultFixture) {
			t.Error("image data doesn't match")
		}
	})

	t.Run("Returns error on a nonexistant image", func(t *testing.T) {
		_, err := provider.Get("nonexistant")
		if err != database.ErrNotFound {
			t.FailNow()
		}
	})
}

func TestNew(t *testing.T) {
	_, err := spaces.New("", "", "", "")
	if err == nil {
		t.Fatal("no error")
	}
}
