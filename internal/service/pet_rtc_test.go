package service

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// TestBuildPetRtcTokenOfficialSample 校验阿里云 ARTC 单参数入会 Token 与官方文档样例一致
func TestBuildPetRtcTokenOfficialSample(t *testing.T) {
	base64Token, err := buildPetRtcToken("abc", "abckey", "abcChannel", "abcUser", "", 1699423634)
	if err != nil {
		t.Fatalf("build token failed: %v", err)
	}

	raw, err := base64.StdEncoding.DecodeString(base64Token)
	if err != nil {
		t.Fatalf("decode base64 failed: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode payload failed: %v", err)
	}

	const wantSign = "3c9ee8d9f8734f0b7560ed8022a0590659113955819724fc9345ab8eedf84f31"
	if payload["token"] != wantSign {
		t.Fatalf("sign mismatch: got %v want %s", payload["token"], wantSign)
	}
	if payload["appid"] != "abc" || payload["channelid"] != "abcChannel" || payload["userid"] != "abcUser" || payload["nonce"] != "" {
		t.Fatalf("payload fields mismatch: %+v", payload)
	}
	if payload["timestamp"] != float64(1699423634) {
		t.Fatalf("timestamp mismatch: %+v", payload["timestamp"])
	}
}
