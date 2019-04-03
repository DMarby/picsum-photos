Lorem Picsum
===========

Lorem Ipsum... but for photos.

## Database migrations
We use [migrate](https://github.com/golang-migrate/migrate) to handle database migrations.

### Applying the database migrations
```
migrate -path migrations -database 'postgresql://postgres@localhost/postgres?sslmode=disable' up
```

### Checking currently applied migrations
```
migrate -path migrations -database 'postgresql://postgres@localhost/postgres?sslmode=disable' version
```

### Creating a new migration
```
migrate create -ext sql -dir migrations my_new_migration
```
