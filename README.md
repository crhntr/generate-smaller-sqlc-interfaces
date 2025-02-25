# Generate Smaller SQLC interfaces

> The bigger the interface the weeker the abstraction. - Rob Pike

SQLC may generates an interface called Querier.
Sometimes you may want to organize separate queries into differenet interfaces (maybe to generate mocks).

This tool helps by findinga a new comment in your sql files and copying from your method signature.

```sql

-- name GetUserByID :one
-- interface: Selectors UserUpdater
SELECT * FROM users WHERE id = $1;

-- name GetTaskByID :on
-- interface: Selectors
SELECT * FROM tasks WHERE id = $1;

-- name SetUserName :exec
-- interface: Selectors, UserUpdater
UPDATE users SET name = $1 WHERE id = $2;

```

```go
/* ... the muxt generated contenets in querier.go ... */

type Selectors interface {
  GetTaskByID(ctx context.Context, id int) (Task, error)
  GetUserByID(ctx context.Context, id int) (User, error)
}

type UserUpdater interface {
  GetUserByID(ctx context.Context, id int) (User, error)
  SetUserName(ctx context.Context, name string, id int) error
}

```

Run it after `sqlc generate` and make sure you are generating Querier.
Pass as arguments the names (or globs) of sql files.
