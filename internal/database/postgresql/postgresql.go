package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/DMarby/picsum-photos/internal/database"

	// Import the pgx stdlib to register the pgx connection type for the sql driver
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/jmoiron/sqlx"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// Register the file driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Provider implements a postgresql based storage
type Provider struct {
	db *sqlx.DB
}

// New establishes the database connection and returns a new Provider instance
func New(address string, maxConns int) (*Provider, error) {
	// Establish the database connection
	db, err := sqlx.Open("pgx", address)
	if err != nil {
		return nil, err
	}

	// Limit the maximum open and idle database connections
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(int(math.Ceil(float64(maxConns) / 2)))

	// Use Unsafe so that the app doesn't fail if we add new columns to the database
	return &Provider{
		db: db.Unsafe(),
	}, nil
}

const waitDelay = time.Second

// Wait blocks until a database connection is ready
// You can use the given context to specify a timeout
func (p *Provider) Wait(ctx context.Context) error {
	lock := make(chan struct{}, 1)
	done := make(chan struct{}, 1)

	ping := func(ctx context.Context) {
		if err := p.db.PingContext(ctx); err == nil {
			done <- struct{}{}
			return
		}

		time.Sleep(waitDelay)
		// Empty the lock channel to allow for another attempt
		<-lock
	}

	for {
		select {
		// Try to send a message to the lock channel
		case lock <- struct{}{}:
			// The lock channel was empty, execute a ping
			go ping(ctx)
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for database connection")
		case <-done:
			return nil
		}
	}
}

// Migrate attempts to migrate the database to the latest migration
func (p *Provider) Migrate(migrationsURL string) error {
	driver, err := postgres.WithInstance(p.db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}

	return err
}

// Get returns the image data for an image id
func (p *Provider) Get(ctx context.Context, id string) (i *database.Image, err error) {
	i = &database.Image{}
	err = p.db.GetContext(ctx, i, "select * from image where id = $1", id)

	if err != nil && err == sql.ErrNoRows {
		return nil, database.ErrNotFound
	}

	return
}

// GetRandom returns a random image ID
func (p *Provider) GetRandom(ctx context.Context) (i *database.Image, err error) {
	i = &database.Image{}
	// This will be slow on large tables
	err = p.db.GetContext(ctx, i, "select * from image order by random() limit 1")
	return
}

// GetRandomWithSeed returns a random image ID based on the given seed
func (p *Provider) GetRandomWithSeed(ctx context.Context, seed int64) (i *database.Image, err error) {
	images, err := p.ListAll(ctx)
	if err != nil {
		return
	}

	source := rand.NewSource(seed)
	random := rand.New(source)

	return &images[random.Intn(len(images))], nil
}

// ListAll returns a list of all the images
func (p *Provider) ListAll(ctx context.Context) ([]database.Image, error) {
	i := []database.Image{}
	err := p.db.SelectContext(ctx, &i, "select * from image order by id")

	if err != nil {
		return nil, err
	}

	return i, nil
}

// List returns a list of all the images with an offset/limit
func (p *Provider) List(ctx context.Context, offset, limit int) ([]database.Image, error) {
	i := []database.Image{}
	err := p.db.SelectContext(ctx, &i, "select * from image order by id OFFSET $1 LIMIT $2", offset, limit)

	if err != nil {
		return nil, err
	}

	return i, nil
}

// Shutdown shuts down the database client
func (p *Provider) Shutdown() {
	p.db.Close()
}
