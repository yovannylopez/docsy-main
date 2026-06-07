module github.com/yovannylopez/docsy-main/pkg/databases

go 1.26.2

replace (
	github.com/yovannylopez/docsy-main/pkg/errors => ../errors
	github.com/yovannylopez/docsy-main/pkg/logging => ../logging
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.30
	github.com/sony/gobreaker v1.0.0
	github.com/stretchr/testify v1.11.1
	github.com/yovannylopez/docsy-main/pkg/errors v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/logging v1.0.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
