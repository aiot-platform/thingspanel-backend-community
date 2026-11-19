package utils

import "testing"

// TestCUserTokenRoundTrip 校验 C 端令牌签发与解析的一致性
func TestCUserTokenRoundTrip(t *testing.T) {
	j := NewJWT([]byte("c1hj5b0sp9cend"))

	token, err := j.GenerateCToken(CUserClaims{
		UserID:       "u-1001",
		Phone:        "13800000000",
		SolutionCode: "pet_medical",
	})
	if err != nil {
		t.Fatalf("generate c token failed: %v", err)
	}

	claims, err := j.ParseCToken(token)
	if err != nil {
		t.Fatalf("parse c token failed: %v", err)
	}
	if claims.UserID != "u-1001" || claims.Phone != "13800000000" || claims.SolutionCode != "pet_medical" {
		t.Fatalf("claims mismatch: %+v", claims)
	}
}

// TestCUserTokenIsolatedFromAdminKey 校验 C 端令牌使用独立密钥，无法用后台密钥解析
func TestCUserTokenIsolatedFromAdminKey(t *testing.T) {
	cJWT := NewJWT([]byte("c1hj5b0sp9cend"))
	adminJWT := NewJWT([]byte("1hj5b0sp9"))

	token, err := cJWT.GenerateCToken(CUserClaims{UserID: "u-1001"})
	if err != nil {
		t.Fatalf("generate c token failed: %v", err)
	}

	if _, err := adminJWT.ParseCToken(token); err == nil {
		t.Fatal("expected parse to fail with admin key, but it succeeded")
	}
}
