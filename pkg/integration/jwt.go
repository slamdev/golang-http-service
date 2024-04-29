package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"gopkg.in/go-jose/go-jose.v2/jwt"
)

type JWTCustomClaims struct {
	Roles []string `json:"roles"`
}

func (c *JWTCustomClaims) Validate(_ context.Context) error { return nil }

func validateToken(ctx context.Context, tokenString string, issuers []string, keyFunc func(context.Context) (interface{}, error), customClaimsFunc func() validator.CustomClaims) (interface{}, error) {
	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return nil, fmt.Errorf("could not parse the token: %w", err)
	}

	if string(validator.RS256) != token.Headers[0].Algorithm {
		return nil, fmt.Errorf("expected %q signing algorithm but token specified %q", validator.RS256, token.Headers[0].Algorithm)
	}

	registeredClaims, customClaims, err := deserializeClaims(ctx, keyFunc, customClaimsFunc, token)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize token claims: %w", err)
	}

	if err = validateClaimsWithLeeway(registeredClaims, issuers, 0); err != nil {
		return nil, fmt.Errorf("expected claims not validated: %w", err)
	}

	if err = customClaims.Validate(ctx); err != nil {
		return nil, fmt.Errorf("custom claims not validated: %w", err)
	}

	validatedClaims := &validator.ValidatedClaims{
		RegisteredClaims: validator.RegisteredClaims{
			Issuer:    registeredClaims.Issuer,
			Subject:   registeredClaims.Subject,
			Audience:  registeredClaims.Audience,
			ID:        registeredClaims.ID,
			Expiry:    numericDateToUnixTime(registeredClaims.Expiry),
			NotBefore: numericDateToUnixTime(registeredClaims.NotBefore),
			IssuedAt:  numericDateToUnixTime(registeredClaims.IssuedAt),
		},
		CustomClaims: customClaims,
	}

	return validatedClaims, nil
}

func deserializeClaims(ctx context.Context, keyFunc func(context.Context) (interface{}, error), customClaimsFunc func() validator.CustomClaims, token *jwt.JSONWebToken) (jwt.Claims, validator.CustomClaims, error) {
	key, err := keyFunc(ctx)
	if err != nil {
		return jwt.Claims{}, nil, fmt.Errorf("error getting the keys from the key func: %w", err)
	}

	claims := []interface{}{&jwt.Claims{}}
	claims = append(claims, customClaimsFunc())

	if err = token.Claims(key, claims...); err != nil {
		return jwt.Claims{}, nil, fmt.Errorf("could not get token claims: %w", err)
	}

	registeredClaims := *claims[0].(*jwt.Claims)

	var customClaims validator.CustomClaims
	if len(claims) > 1 {
		customClaims = claims[1].(validator.CustomClaims)
	}

	return registeredClaims, customClaims, nil
}

func validateClaimsWithLeeway(actualClaims jwt.Claims, issuers []string, leeway time.Duration) error {
	now := time.Now()

	foundIssuer := false
	for _, value := range issuers {
		if actualClaims.Issuer == value {
			foundIssuer = true
			break
		}
	}
	if !foundIssuer {
		return jwt.ErrInvalidIssuer
	}

	if actualClaims.NotBefore != nil && now.Add(leeway).Before(actualClaims.NotBefore.Time()) {
		return jwt.ErrNotValidYet
	}

	if actualClaims.Expiry != nil && now.Add(-leeway).After(actualClaims.Expiry.Time()) {
		return jwt.ErrExpired
	}

	if actualClaims.IssuedAt != nil && now.Add(leeway).Before(actualClaims.IssuedAt.Time()) {
		return jwt.ErrIssuedInTheFuture
	}

	return nil
}

func numericDateToUnixTime(date *jwt.NumericDate) int64 {
	if date != nil {
		return date.Time().Unix()
	}
	return 0
}
