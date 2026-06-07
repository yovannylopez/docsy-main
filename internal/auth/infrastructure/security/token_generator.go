package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

// TokenGenerator is the struct that holds the token generator
type TokenGenerator struct {
	secretKey []byte
}

// NewTokenGenerator creates a new instance of TokenGenerator
func NewTokenGenerator(secretKey string) *TokenGenerator {
	return &TokenGenerator{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken generates access and refresh JWTs. If sessionID is not empty it is included in both tokens.
func (tg *TokenGenerator) GenerateToken(user *entities.User, sessionID string) (*entities.AuthToken, error) {
	var roleName string
	if len(user.Roles) > 0 {
		roleName = user.Roles[0].Name
	}

	accessClaims := jwt.MapClaims{
		constants.JWTClaimUserID:            user.ID,
		constants.JWTClaimEmail:             user.Email,
		constants.JWTClaimRole:              roleName,
		constants.JWTClaimExp:               time.Now().Add(time.Hour * constants.AccessTokenExpirationHours).Unix(),
		constants.JWTClaimIat:               time.Now().Unix(),
		constants.JWTClaimPasswordChangedAt: user.PasswordChangedAt.Unix(),
	}
	if sessionID != "" {
		accessClaims[constants.JWTClaimSessionID] = sessionID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	tokenString, err := token.SignedString(tg.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		constants.JWTClaimUserID: user.ID,
		constants.JWTClaimType:   constants.JWTTokenTypeRefresh,
		constants.JWTClaimExp:    time.Now().Add(time.Hour * 24 * constants.RefreshTokenExpirationDays).Unix(),
		constants.JWTClaimIat:    time.Now().Unix(),
	}
	if sessionID != "" {
		refreshClaims[constants.JWTClaimSessionID] = sessionID
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(tg.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &entities.AuthToken{
		AccessToken:  tokenString,
		TokenType:    constants.BearerTokenType,
		ExpiresAt:    time.Now().Add(time.Hour * constants.AccessTokenExpirationHours),
		RefreshToken: refreshTokenString,
	}, nil
}

// ValidateToken validates a token and returns its claims
func (tg *TokenGenerator) ValidateToken(tokenString string) (map[string]any, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tg.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ParseRefreshToken validates the signature and expiration of the refresh JWT and returns userID and sessionID.
func (tg *TokenGenerator) ParseRefreshToken(refreshToken string) (userID, sessionID string, err error) {
	claims, err := tg.ValidateToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	tokenType, ok := claims[constants.JWTClaimType].(string)
	if !ok || tokenType != constants.JWTTokenTypeRefresh {
		return "", "", fmt.Errorf("token is not a refresh token")
	}

	uid, ok := claims[constants.JWTClaimUserID].(string)
	if !ok || uid == "" {
		return "", "", fmt.Errorf("invalid user_id in refresh token")
	}

	sid, _ := claims[constants.JWTClaimSessionID].(string)

	return uid, sid, nil
}
