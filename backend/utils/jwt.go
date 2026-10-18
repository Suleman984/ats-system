package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	audienceAdmin      = "admin"
	audienceSuperAdmin = "super_admin"
	// Candidate chat streams: scoped to one application, short-lived because it travels in a URL.
	audienceCandidateStream = "candidate_stream"
	candidateStreamTTL      = time.Hour
	tokenTTL                = 24 * time.Hour
)

type Claims struct {
	CompanyID string `json:"company_id,omitempty"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

// InitJWT must run before any token is issued or verified.
func InitJWT() error {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return errors.New("JWT_SECRET must be set and at least 32 characters long")
	}
	jwtSecret = []byte(secret)
	return nil
}

func GenerateJWT(adminID, companyID uuid.UUID) (string, error) {
	return sign(adminID, audienceAdmin, companyID.String())
}

func GenerateSuperAdminJWT(superAdminID uuid.UUID) (string, error) {
	return sign(superAdminID, audienceSuperAdmin, "")
}

// AdminToken is a verified company-admin token.
type AdminToken struct {
	AdminID   uuid.UUID
	CompanyID uuid.UUID
	IssuedAt  time.Time
}

// VerifyAdminToken also returns when the token was issued, so sessions predating a password change can be rejected.
func VerifyAdminToken(token string) (AdminToken, error) {
	claims, err := parse(token, audienceAdmin)
	if err != nil {
		return AdminToken{}, err
	}
	adminID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return AdminToken{}, errors.New("invalid subject")
	}
	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		return AdminToken{}, errors.New("invalid company")
	}
	if claims.IssuedAt == nil {
		return AdminToken{}, errors.New("missing issued-at")
	}
	return AdminToken{adminID, companyID, claims.IssuedAt.Time}, nil
}

// VerifyJWT returns the admin and company IDs from a company-admin token.
func VerifyJWT(token string) (adminID, companyID uuid.UUID, err error) {
	t, err := VerifyAdminToken(token)
	return t.AdminID, t.CompanyID, err
}

func VerifySuperAdminJWT(token string) (uuid.UUID, error) {
	claims, err := parse(token, audienceSuperAdmin)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("invalid subject")
	}
	return id, nil
}

func GenerateCandidateStreamToken(applicationID, companyID uuid.UUID) (string, error) {
	return signWithTTL(applicationID, audienceCandidateStream, companyID.String(), candidateStreamTTL)
}

// VerifyCandidateStreamToken returns the application and company a stream token was issued for.
func VerifyCandidateStreamToken(token string) (applicationID, companyID uuid.UUID, err error) {
	claims, err := parse(token, audienceCandidateStream)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if applicationID, err = uuid.Parse(claims.Subject); err != nil {
		return uuid.Nil, uuid.Nil, errors.New("invalid subject")
	}
	if companyID, err = uuid.Parse(claims.CompanyID); err != nil {
		return uuid.Nil, uuid.Nil, errors.New("invalid company")
	}
	return applicationID, companyID, nil
}

func sign(subject uuid.UUID, audience, companyID string) (string, error) {
	return signWithTTL(subject, audience, companyID, tokenTTL)
}

func signWithTTL(subject uuid.UUID, audience, companyID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		CompanyID: companyID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject.String(),
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func parse(token, audience string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(*jwt.Token) (any, error) { return jwtSecret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
