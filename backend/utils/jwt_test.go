package utils

import (
	"testing"

	"github.com/google/uuid"
)

func setupJWT(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret-that-is-at-least-32-characters")
	if err := InitJWT(); err != nil {
		t.Fatal(err)
	}
}

func TestInitJWTRejectsShortSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "short")
	if err := InitJWT(); err == nil {
		t.Fatal("expected an error for a short secret")
	}
}

func TestAdminTokenRoundTrip(t *testing.T) {
	setupJWT(t)
	admin, company := uuid.New(), uuid.New()

	token, err := GenerateJWT(admin, company)
	if err != nil {
		t.Fatal(err)
	}
	gotAdmin, gotCompany, err := VerifyJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	if gotAdmin != admin || gotCompany != company {
		t.Fatalf("got (%s, %s), want (%s, %s)", gotAdmin, gotCompany, admin, company)
	}
}

func TestTokensAreNotInterchangeable(t *testing.T) {
	setupJWT(t)

	adminToken, _ := GenerateJWT(uuid.New(), uuid.New())
	if _, err := VerifySuperAdminJWT(adminToken); err == nil {
		t.Fatal("a company-admin token must not pass super-admin verification")
	}

	superToken, _ := GenerateSuperAdminJWT(uuid.New())
	if _, _, err := VerifyJWT(superToken); err == nil {
		t.Fatal("a super-admin token must not pass company-admin verification")
	}
}

func TestVerifyRejectsTamperedToken(t *testing.T) {
	setupJWT(t)
	token, _ := GenerateJWT(uuid.New(), uuid.New())
	if _, _, err := VerifyJWT(token + "x"); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}
