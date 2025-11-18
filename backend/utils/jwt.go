package utils

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AdminID   string `json:"admin_id"`
	CompanyID string `json:"company_id"`
	jwt.RegisteredClaims
}

type SuperAdminClaims struct {
	SuperAdminID string `json:"super_admin_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(adminID, companyID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	claims := Claims{
		AdminID:   adminID,
		CompanyID: companyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateSuperAdminJWT(superAdminID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	claims := SuperAdminClaims{
		SuperAdminID: superAdminID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyJWT(tokenString string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Debug: Log what we extracted
		log.Printf("VerifyJWT: Successfully parsed claims - AdminID: '%s', CompanyID: '%s'", claims.AdminID, claims.CompanyID)
		
		// Debug: Check raw claims map
		if claimsMap, ok := token.Claims.(jwt.MapClaims); ok {
			claimsJSON, _ := json.Marshal(claimsMap)
			log.Printf("VerifyJWT: Raw claims map: %s", string(claimsJSON))
		}
		
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func VerifySuperAdminJWT(tokenString string) (*SuperAdminClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	token, err := jwt.ParseWithClaims(tokenString, &SuperAdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*SuperAdminClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
