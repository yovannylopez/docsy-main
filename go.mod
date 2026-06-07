module github.com/yovannylopez/docsy-main

go 1.26.2

replace (
	github.com/yovannylopez/docsy-main/pkg/config => ./pkg/config
	github.com/yovannylopez/docsy-main/pkg/constants => ./pkg/constants
	github.com/yovannylopez/docsy-main/pkg/databases => ./pkg/databases
	github.com/yovannylopez/docsy-main/pkg/errors => ./pkg/errors
	github.com/yovannylopez/docsy-main/pkg/http_status => ./pkg/http_status
	github.com/yovannylopez/docsy-main/pkg/logging => ./pkg/logging
	github.com/yovannylopez/docsy-main/pkg/openapi => ./pkg/openapi
	github.com/yovannylopez/docsy-main/pkg/pagination => ./pkg/pagination
	github.com/yovannylopez/docsy-main/pkg/responses => ./pkg/responses
	github.com/yovannylopez/docsy-main/pkg/validators => ./pkg/validators
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/golang-jwt/jwt/v5 v5.2.3
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/labstack/echo/v4 v4.13.4
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.30
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.7.0
	github.com/stretchr/testify v1.11.1
	github.com/yovannylopez/docsy-main/pkg/config v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/constants v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/databases v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/errors v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/http_status v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/logging v1.0.0
	github.com/yovannylopez/docsy-main/pkg/openapi v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/pagination v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/responses v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.49.0
)

require (
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-migrate/migrate/v4 v4.18.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/otp v1.5.0
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
