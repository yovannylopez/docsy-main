module github.com/yovannylopez/docsy-main/pkg/responses

go 1.26.2

replace github.com/yovannylopez/docsy-main/pkg/constants => ../constants

replace github.com/yovannylopez/docsy-main/pkg/http_status => ../http_status

replace github.com/yovannylopez/docsy-main/pkg/pagination => ../pagination

require (
	github.com/labstack/echo/v4 v4.13.4
	github.com/stretchr/testify v1.11.1
	github.com/yovannylopez/docsy-main/pkg/http_status v0.0.0-00010101000000-000000000000
	github.com/yovannylopez/docsy-main/pkg/pagination v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/yovannylopez/docsy-main/pkg/constants v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
