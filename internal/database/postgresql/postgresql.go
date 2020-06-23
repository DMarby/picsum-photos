package postgresql

import (
	"context"
	"database/sql"
	"math/rand"

	"github.com/DMarby/picsum-photos/internal/database"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

// Provider implements a postgresql based storage
type Provider struct {
	db *sqlx.DB
}

// New returns a new Provider instance
func New(address string) (*Provider, error) {
	// Needed to work with pgbouncer
	d := &stdlib.DriverConfig{
		ConnConfig: pgx.ConnConfig{
			PreferSimpleProtocol: true,
			RuntimeParams: map[string]string{
				"client_encoding": "UTF8",
			},
			CustomConnInfo: func(c *pgx.Conn) (*pgtype.ConnInfo, error) {
				info := c.ConnInfo.DeepCopy()
				info.RegisterDataType(pgtype.DataType{
					Value: &pgtype.OIDValue{},
					Name:  "int8OID",
					OID:   pgtype.Int8OID,
				})

				return info, nil
			},
		},
	}

	stdlib.RegisterDriverConfig(d)

	db, err := sqlx.Connect("pgx", d.ConnectionString(address))
	if err != nil {
		return nil, err
	}

	// Use Unsafe so that the app doesn't fail if we add new columns to the database
	return &Provider{
		db: db.Unsafe(),
	}, nil
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
