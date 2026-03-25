package service

import (
	"testing"
)

func TestGenerateSessionToken_UniqueTokens(t *testing.T) {
	token1, err := generateSessionToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token2, err := generateSessionToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 == token2 {
		t.Error("expected unique tokens, got duplicates")
	}

	if len(token1) != 64 {
		t.Errorf("expected 64-char hex token, got %d chars", len(token1))
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	token := "test-token-123"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	if hash1 != hash2 {
		t.Error("expected deterministic hash")
	}

	if len(hash1) != 64 {
		t.Errorf("expected 64-char SHA-256 hex hash, got %d chars", len(hash1))
	}
}

func TestHashToken_DifferentInputsDifferentHashes(t *testing.T) {
	hash1 := hashToken("token-a")
	hash2 := hashToken("token-b")

	if hash1 == hash2 {
		t.Error("expected different hashes for different inputs")
	}
}
