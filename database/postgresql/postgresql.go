package postgresql

import (
	"database/sql"

	"github.com/DMarby/picsum-photos/database"

	// Import the postgresql driver
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

// Provider implements a postgresql based storage
type Provider struct {
	db *sqlx.DB
}

// New returns a new Provider instance
func New(address string) (*Provider, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		return nil, err
	}

	// Use Unsafe so that the app doesn't fail if we add new columns to the database
	return &Provider{
		db: db.Unsafe(),
	}, nil
}

// Get returns the image data for an image id
func (p *Provider) Get(id string) (*database.Image, error) {
	i := &database.Image{}
	err := p.db.Get(i, "select * from image where id = $1", id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, database.ErrNotFound
		}

		return nil, err
	}

	return i, nil
}

// GetRandom returns a random image ID
func (p *Provider) GetRandom() (id string, err error) {
	// This will be slow on large tables
	err = p.db.Get(&id, "select id from image order by random() limit 1")
	return
}

// List returns a list of all the images
func (p *Provider) List() ([]database.Image, error) {
	i := []database.Image{}
	err := p.db.Select(&i, "select * from image")

	if err != nil {
		return nil, err
	}

	return i, nil
}

// Shutdown shuts down the database client
func (p *Provider) Shutdown() {
	p.db.Close()
}
