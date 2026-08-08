// Package constants centralizes strings, logical response codes, HTTP headers,
// shared numeric limits and dynamic message functions (messages.go).
// The success/error messages in Spanish are aligned with the descriptions
// of pkg/http_status when applicable (e.g. SuccessMessage ↔ http_status.OK.Description).
// It does not define HTTP numeric codes: use net/http or pkg/http_status for that.
package constants

// Standard response codes
const (
	// Success codes
	SuccessCode = "SUCCESS"

	// General error codes
	ErrorCode         = "ERROR"
	ValidationError   = "VALIDATION_ERROR"
	NotFoundError     = "NOT_FOUND"
	ConflictError     = "CONFLICT"
	UnauthorizedError = "UNAUTHORIZED"
	ForbiddenError    = "FORBIDDEN"
	InternalError     = "INTERNAL_ERROR"
)

// Standard response messages
const (
	// Success messages
	SuccessMessage = "Operation successful"
	CreatedMessage = "Resource created successfully"
	UpdatedMessage = "Resource updated successfully"
	DeletedMessage = "Resource deleted successfully"

	// General error messages
	GenericErrorMessage    = "An unexpected error occurred"
	ValidationErrorMessage = "The provided data is not valid"
	NotFoundMessage        = "The requested resource was not found"
	ConflictMessage        = "The resource already exists"
	UnauthorizedMessage    = "You are not authorized to access this resource"
	ForbiddenMessage       = "Access denied"
	InternalErrorMessage   = "Internal server error"

	// Authentication messages
	InvalidCredentialsMessage         = "Invalid credentials"
	LoginEmailPasswordRequiredMessage = "Email and password are required"
	LoginSuccessMessage               = "Login successful"
	TokenExpiredMessage               = "Token expired"
	TokenInvalidMessage               = "Token invalid"
	LoginRequiredMessage              = "Authentication required"

	// MFA messages
	MFASetupInitiatedMessage         = "MFA setup initiated"
	MFAEnabledMessage                = "MFA enabled successfully"
	MFADisabledMessage               = "MFA disabled successfully"
	MFAChallengeRequiredMessage      = "MFA challenge required"
	MFAVerifiedMessage               = "MFA verified successfully"
	MFAInvalidCodeMessage            = "Invalid or expired MFA code"
	MFANotEnabledMessage             = "MFA is not enabled for this account"
	MFAAlreadyEnabledMessage         = "MFA is already enabled for this account"
	MFASetupTokenRequiredMessage     = "setup_token and totp_code are required"
	MFAChallengeTokenRequiredMessage = "challenge_token and totp_code are required"
	MFATOTPCodeRequiredMessage       = "totp_code is required"

	// Password change messages
	ChangePasswordSuccessMessage = "Password changed successfully"
	// ChangePasswordFailedMessage is generic on purpose: must not reveal whether the
	// current password, the account state, or any other sensitive detail caused the failure.
	ChangePasswordFailedMessage    = "Could not change password"
	ChangePasswordSameAsCurrentMsg = "New password must be different from the current password"
	// PasswordInHistoryMessage is shown when the new password matches a recently used one.
	PasswordInHistoryMessage = "The new password cannot match a recently used password"

	// Validation messages
	RequiredFieldMessage = "The field is required"
	InvalidFormatMessage = "Invalid format"
	InvalidLengthMessage = "Invalid length"

	// Functionality messages
	NotImplementedMessage = "Functionality not implemented yet"

	// Health check messages
	ServiceHealthyMessage = "Service is healthy"
	ServiceReadyMessage   = "Service is ready"
	ServiceAliveMessage   = "Service is alive"
	// Readiness (probes / load balancers): text stable used by HealthHandler.
	ServiceReadyToReceiveTrafficMessage    = "Service is ready to receive traffic"
	ServiceNotReadyToReceiveTrafficMessage = "Service is not ready to receive traffic"
)

// Status slugs for JSON health checks (aligned with regular probes).
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
)

// Domain specific messages
const (
	// Users
	UserNotFoundMessage          = "User not found"
	UserAlreadyExistsMessage     = "User already exists"
	UserCreatedMessage           = "User created successfully"
	UserUpdatedMessage           = "User updated successfully"
	UserDeletedMessage           = "User deleted successfully"
	InvalidEmailMessage          = "Invalid email format"
	InvalidPasswordMessage       = "Password does not meet requirements"
	EmailAlreadyExistsMessage    = "Email already registered"
	UsernameAlreadyExistsMessage = "Username already taken"

	// Documents
	DocumentNotFoundMessage      = "Document not found"
	DocumentAlreadyExistsMessage = "Document already exists"
	DocumentCreatedMessage       = "Document created successfully"
	DocumentUpdatedMessage       = "Document updated successfully"
	DocumentDeletedMessage       = "Document deleted successfully"
	InvalidFileFormatMessage     = "Invalid file format"
	FileTooLargeMessage          = "File is too large"

	// Roles and permissions
	RoleNotFoundMessage            = "Role not found"
	RoleAlreadyExistsMessage       = "Role already exists"
	PermissionDeniedMessage        = "Permission denied"
	InsufficientPermissionsMessage = "Insufficient permissions"
)

// Standard headers
const (
	RequestIDHeader     = "X-Request-ID"
	AuthorizationHeader = "Authorization"
	ContentTypeHeader   = "Content-Type"
	UserAgentHeader     = "User-Agent"
)

// Header values
const (
	ContentTypeJSON = "application/json"
	// BearerTokenType is the value of the token_type field in OAuth/JWT responses (without space).
	BearerTokenType = "Bearer"
	BearerPrefix    = "Bearer "
)

// Limits and configurations
const (
	MaxRequestSizeMB = 50
	DefaultPageSize  = 20
	MaxPageSize      = 100
)

// Time configurations (in hours/days)
const (
	// Token expiration times
	TokenExpirationHours       = 24
	AccessTokenExpirationHours = 24
	RefreshTokenExpirationDays = 7
	SessionExpirationDays      = 7

	// MFA token TTLs (in minutes)
	MFASetupTokenTTLMinutes     = 10
	MFAChallengeTokenTTLMinutes = 5

	// Timeout times for health checks
	HealthCheckTimeoutSeconds    = 5
	ReadinessCheckTimeoutSeconds = 3

	// Password configurations
	MinPasswordLength = 8

	// File configurations
	DefaultMaxFileSizeMB = 10
	BytesPerKB           = 1024
	BytesPerMB           = 1024 * 1024
	BytesPerGB           = 1024 * 1024 * 1024
	// DefaultStorageQuotaBytes is the soft display quota for the sidebar (10 GiB).
	DefaultStorageQuotaBytes = 10 * BytesPerGB

	// Specific configurations for multiple files
	MaxFilesPerUpload = 10  // Maximum 10 files per upload
	MaxFileSizeMB     = 10  // Maximum 10MB per file
	MaxTotalSizeMB    = 50  // Maximum 50MB total per upload
	MaxFileNameLength = 255 // Maximum 255 characters for file name

	// Rate limiting for uploads
	MaxUploadsPerMinute = 5  // Maximum 5 uploads per minute per user
	MaxUploadsPerHour   = 20 // Maximum 20 uploads per hour per user

	// Allowed MIME types for PDF
	AllowedPDFMimeTypes  = "application/pdf" // Only PDF
	AllowedPDFExtensions = ".pdf"            // Only .pdf extension

	// Pagination configurations
	MaxUsersLimit = 100

	// Batch creation configurations
	MaxUsersBatchSize = 50
)

// Date formats
const (
	DateFormat     = "2006-01-02"
	DateTimeFormat = "2006-01-02T15:04:05Z07:00"
	TimeFormat     = "15:04:05"
)

// Statuses
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusPending  = "pending"
	StatusDeleted  = "deleted"
)

// Content types
const (
	ContentTypeDocument = "document"
	ContentTypeImage    = "image"
	ContentTypeVideo    = "video"
	ContentTypeAudio    = "audio"
)

// System roles
const (
	RoleAdmin   = "admin"
	RoleUser    = "user"
	RoleManager = "manager"
	RoleViewer  = "viewer"
)

// System permissions
const (
	PermissionRead   = "read"
	PermissionWrite  = "write"
	PermissionDelete = "delete"
	PermissionAdmin  = "admin"
)
