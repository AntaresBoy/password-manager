package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	keyLength   = 32
	saltLength  = 32
	nonceLength = 12

	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
)

func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, keyLength)
}

func Encrypt(plaintext []byte, password string) (salt, nonce, ciphertext []byte, err error) {
	salt, err = GenerateSalt(saltLength)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate salt: %w", err)
	}

	nonce, err = GenerateNonce()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate nonce: %w", err)
	}

	gcm, err := gcmForPassword(password, salt)
	if err != nil {
		return nil, nil, nil, err
	}

	return salt, nonce, gcm.Seal(nil, nonce, plaintext, nil), nil
}

func Decrypt(salt, nonce, ciphertext []byte, password string) ([]byte, error) {
	if len(salt) != saltLength {
		return nil, errors.New("invalid salt length")
	}
	if len(nonce) != nonceLength {
		return nil, errors.New("invalid nonce length")
	}

	gcm, err := gcmForPassword(password, salt)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed")
	}

	return plaintext, nil
}

func GenerateSalt(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("salt length must be positive")
	}

	salt := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, nonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func gcmForPassword(password string, salt []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(DeriveKey(password, salt))
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	return gcm, nil
}
