package container_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	authcontainer "github.com/yovannylopez/docsy-main/internal/auth/infrastructure/container"
	sharedcfg "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/config"
)

func TestNewAuthContainer_WiresDependencies(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")
	c := authcontainer.NewAuthContainer(sqlxDB, "test-jwt-secret", sharedcfg.LDAPConfig{})

	require.NotNil(t, c)
	require.NotNil(t, c.UserRepository)
	require.NotNil(t, c.SessionRepository)
	require.NotNil(t, c.PasswordHasher)
	require.NotNil(t, c.TokenGenerator)
	require.NotNil(t, c.LoginUseCase)
	require.NotNil(t, c.AuthUseCase)
	require.NotNil(t, c.ListAuditLogsUseCase)
	require.NotNil(t, c.AuditService)
	require.NotNil(t, c.AuthHandler)
	require.NotNil(t, c.AuditHandler)
	require.NotNil(t, c.AuthHTTPMiddleware)
}

func TestAuthContainer_Getters(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	c := authcontainer.NewAuthContainer(sqlx.NewDb(db, "postgres"), "secret", sharedcfg.LDAPConfig{})

	require.Equal(t, c.AuthHandler, c.GetAuthHandler())
	require.Equal(t, c.AuditHandler, c.GetAuditHandler())
	require.Equal(t, c.AuditService, c.GetAuditService())
	require.Equal(t, c.PasswordHasher, c.GetPasswordHasher())
}
