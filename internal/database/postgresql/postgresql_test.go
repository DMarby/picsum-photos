//go:build integration
// +build integration

package postgresql_test

import (
	"context"
	"reflect"
	"time"

	"github.com/DMarby/picsum-photos/internal/database"
	"github.com/DMarby/picsum-photos/internal/database/postgresql"
	"github.com/jmoiron/sqlx"

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

const address = "postgresql://postgres:postgres@localhost:5433/postgres"

func TestPostgresql(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	provider := mustInitialize(t, "file://../../../migrations")
	defer provider.Shutdown()

	defer clean()

	db := sqlx.MustConnect("pgx", address)
	defer db.Close()

	// Add some test data to the database
	db.MustExec(`
		insert into image(id, author, url, width, height) VALUES
		(1, 'John Doe', 'https://picsum.photos', 300, 400),
		(2, 'John Doe', 'https://picsum.photos', 300, 400)
	`)

	t.Run("Get an image by id", func(t *testing.T) {
		buf, err := provider.Get(ctx, "1")
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(buf, &image) {
			t.Error("image data doesn't match")
		}
	})

	t.Run("Returns error on a nonexistant image", func(t *testing.T) {
		_, err := provider.Get(ctx, "nonexistant")
		if err == nil || err.Error() != database.ErrNotFound.Error() {
			t.FailNow()
		}
	})

	t.Run("Returns a random image", func(t *testing.T) {
		image, err := provider.GetRandom(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if image.ID != "1" && image.ID != "2" && image.ID != "3" {
			t.Error("wrong image")
		}
	})

	t.Run("Returns a random based on the seed", func(t *testing.T) {
		image, err := provider.GetRandomWithSeed(ctx, 0)
		if err != nil {
			t.Fatal(err)
		}

		if image.ID != "1" {
			t.Error("wrong image")
		}
	})

	t.Run("Returns a list of all the images", func(t *testing.T) {
		images, err := provider.ListAll(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(images, []database.Image{image, secondImage}) {
			t.Error("image data doesn't match")
		}
	})

	t.Run("Returns a list of images", func(t *testing.T) {
		images, err := provider.List(ctx, 1, 1)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(images, []database.Image{secondImage}) {
			t.Error("image data doesn't match")
		}
	})
}

// mustInitialize connects to and migrates the database
func mustInitialize(t *testing.T, migrationsURL string) *postgresql.Provider {
	db, err := postgresql.New(address, 0)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = db.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Migrate(migrationsURL)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

// clean cleans up the database after testing
func clean() {
	db := sqlx.MustConnect("pgx", address)
	defer db.Close()

	db.MustExec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
}
