package ports

// PasswordHasher defines the operations for password hashing
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) (bool, error)
}
