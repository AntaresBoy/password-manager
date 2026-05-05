package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKeyDeterministic(t *testing.T) {
	salt := bytes.Repeat([]byte{0x01}, 32)

	key1 := DeriveKey("password", salt)
	key2 := DeriveKey("password", salt)

	if !bytes.Equal(key1, key2) {
		t.Fatal("expected same password and salt to derive the same key")
	}
	if len(key1) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key1))
	}
}

func TestDeriveKeyDifferentSalts(t *testing.T) {
	key1 := DeriveKey("password", bytes.Repeat([]byte{0x01}, 32))
	key2 := DeriveKey("password", bytes.Repeat([]byte{0x02}, 32))

	if bytes.Equal(key1, key2) {
		t.Fatal("expected different salts to derive different keys")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := []byte("secret payload")
	password := "correct horse battery staple"

	salt, nonce, ciphertext, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if len(salt) != 32 {
		t.Fatalf("expected 32-byte salt, got %d", len(salt))
	}
	if len(nonce) != 12 {
		t.Fatalf("expected 12-byte nonce, got %d", len(nonce))
	}
	if len(ciphertext) == 0 {
		t.Fatal("expected non-empty ciphertext")
	}
	if bytes.Contains(ciphertext, plaintext) {
		t.Fatal("ciphertext must not contain plaintext")
	}

	decrypted, err := Decrypt(salt, nonce, ciphertext, password)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecryptFailsWithWrongPassword(t *testing.T) {
	salt, nonce, ciphertext, err := Encrypt([]byte("secret payload"), "correct-password")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	_, err = Decrypt(salt, nonce, ciphertext, "wrong-password")
	if err == nil {
		t.Fatal("expected wrong password to fail")
	}
	if err.Error() != "decryption failed" {
		t.Fatalf("expected generic auth failure, got %q", err.Error())
	}
}

func TestDecryptFailsWithTamperedCiphertext(t *testing.T) {
	salt, nonce, ciphertext, err := Encrypt([]byte("secret payload"), "password")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	ciphertext[len(ciphertext)-1] ^= 0xff

	_, err = Decrypt(salt, nonce, ciphertext, "password")
	if err == nil {
		t.Fatal("expected tampered ciphertext to fail")
	}
	if err.Error() != "decryption failed" {
		t.Fatalf("expected generic auth failure, got %q", err.Error())
	}
}

func TestDecryptValidatesSaltAndNonceLength(t *testing.T) {
	salt, nonce, ciphertext, err := Encrypt([]byte("secret payload"), "password")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := Decrypt(salt[:31], nonce, ciphertext, "password"); err == nil {
		t.Fatal("expected short salt to fail")
	}
	if _, err := Decrypt(salt, nonce[:11], ciphertext, "password"); err == nil {
		t.Fatal("expected short nonce to fail")
	}
}

func TestGenerateSaltLengthAndUniqueness(t *testing.T) {
	salt1, err := GenerateSalt(32)
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	salt2, err := GenerateSalt(32)
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}

	if len(salt1) != 32 {
		t.Fatalf("expected 32-byte salt, got %d", len(salt1))
	}
	if len(salt2) != 32 {
		t.Fatalf("expected 32-byte salt, got %d", len(salt2))
	}
	if bytes.Equal(salt1, salt2) {
		t.Fatal("expected generated salts to be unique")
	}
}

func TestGenerateSaltRejectsInvalidLength(t *testing.T) {
	if _, err := GenerateSalt(0); err == nil {
		t.Fatal("expected zero-length salt to fail")
	}
	if _, err := GenerateSalt(-1); err == nil {
		t.Fatal("expected negative-length salt to fail")
	}
}

func TestGenerateNonceLengthAndUniqueness(t *testing.T) {
	nonce1, err := GenerateNonce()
	if err != nil {
		t.Fatalf("generate nonce: %v", err)
	}
	nonce2, err := GenerateNonce()
	if err != nil {
		t.Fatalf("generate nonce: %v", err)
	}

	if len(nonce1) != 12 {
		t.Fatalf("expected 12-byte nonce, got %d", len(nonce1))
	}
	if len(nonce2) != 12 {
		t.Fatalf("expected 12-byte nonce, got %d", len(nonce2))
	}
	if bytes.Equal(nonce1, nonce2) {
		t.Fatal("expected generated nonces to be unique")
	}
}
