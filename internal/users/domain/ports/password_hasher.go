package ports

// PasswordHasher defines the operations for password hashing in the users context
type PasswordHasher interface {
	HashPassword(password string) (string, error)
}
