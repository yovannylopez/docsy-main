package migrations

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yovannylopez/docsy-main/pkg/databases/migrate"
)

type testCase struct {
	name          string
	stubErr       error
	wantErr       bool
	wantErrSubstr string
}

func runCommonTests(t *testing.T, cases []testCase, fn func(stubErr error) error) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := fn(tt.stubErr)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrSubstr != "" {
					assert.Contains(t, err.Error(), tt.wantErrSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	cases := []testCase{
		{name: "success", stubErr: nil, wantErr: false},
		{name: "failure wraps error", stubErr: errors.New("boom"), wantErr: true, wantErrSubstr: "error running migrations: boom"},
	}

	runCommonTests(t, cases, func(stubErr error) error {
		return runUp("postgres://user:pass@localhost:5432/db", "migrations/path", func(opts migrate.Options) error { return stubErr })
	})
}

func TestRollback(t *testing.T) {
	cases := []testCase{
		{name: "success", stubErr: nil, wantErr: false},
		{name: "failure wraps error", stubErr: errors.New("oops"), wantErr: true, wantErrSubstr: "error running migration rollback: oops"},
	}

	runCommonTests(t, cases, func(stubErr error) error {
		return runDown("postgres://user:pass@localhost:5432/db", "migrations/path", func(opts migrate.Options) error { return stubErr })
	})
}
