package auth

import (
	"context"
	"errors"
	"time"

	"github.com/go-thor/thor/pkg"
	"github.com/golang-jwt/jwt/v5"
)

// JWTOption is a JWT middleware option
type JWTOption func(*jwtOptions)

type jwtOptions struct {
	secret        []byte
	authKey       string
	signingMethod jwt.SigningMethod
	validator     func(claims jwt.Claims) bool
}

// DefaultValidator is the default validator
func DefaultValidator(claims jwt.Claims) bool {
	// Check if the token is expired
	expTime, err := claims.GetExpirationTime()
	if err != nil {
		return false
	}
	if expTime == nil {
		return true // No expiration time set
	}
	return !expTime.Before(time.Now())
}

// WithSecret sets the secret for the middleware
func WithSecret(secret []byte) JWTOption {
	return func(o *jwtOptions) {
		o.secret = secret
	}
}

// WithAuthKey sets the auth key for the middleware
func WithAuthKey(authKey string) JWTOption {
	return func(o *jwtOptions) {
		o.authKey = authKey
	}
}

// WithSigningMethod sets the signing method for the middleware
func WithSigningMethod(method jwt.SigningMethod) JWTOption {
	return func(o *jwtOptions) {
		o.signingMethod = method
	}
}

// WithValidator sets the validator for the middleware
func WithValidator(validator func(claims jwt.Claims) bool) JWTOption {
	return func(o *jwtOptions) {
		o.validator = validator
	}
}

// NewJWT creates a new JWT authentication middleware
func NewJWT(opts ...JWTOption) pkg.Middleware {
	options := &jwtOptions{
		secret:        []byte("thor"),
		authKey:       "authorization",
		signingMethod: jwt.SigningMethodHS256,
		validator:     DefaultValidator,
	}

	for _, opt := range opts {
		opt(options)
	}

	return func(next pkg.Handler) pkg.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Get the token from the context
			tokenStr, ok := ctx.Value(options.authKey).(string)
			if !ok {
				return nil, errors.New("missing token")
			}

			// Parse the token
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// Validate the signing method
				if token.Method != options.signingMethod {
					return nil, errors.New("unexpected signing method")
				}
				return options.secret, nil
			})

			if err != nil {
				return nil, err
			}

			// Validate the token
			if !token.Valid {
				return nil, errors.New("invalid token")
			}

			// Validate the claims
			if options.validator != nil && !options.validator(token.Claims) {
				return nil, errors.New("token validation failed")
			}

			// Call the next handler
			return next(ctx, req)
		}
	}
}

// GenerateToken generates a JWT token
func GenerateToken(secret []byte, claims jwt.Claims, method jwt.SigningMethod) (string, error) {
	// Create a new token
	token := jwt.NewWithClaims(method, claims)

	// Sign the token with the secret
	return token.SignedString(secret)
}
