package vault

import (
	"bytes"
	"errors"
	"fmt"

	pkgcrypto "passmgr/pkg/crypto"
)

const magic = "PMV1"

func encryptVault(plaintext []byte, password string) ([]byte, error) {
	salt, nonce, ciphertext, err := pkgcrypto.Encrypt(plaintext, password)
	if err != nil {
		return nil, fmt.Errorf("encrypt vault: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(magic)
	buf.Write(salt)
	buf.Write(nonce)
	buf.Write(ciphertext)
	return buf.Bytes(), nil
}

func decryptVault(data []byte, password string) ([]byte, error) {
	if len(data) < 4+32+12 {
		return nil, errors.New("vault data too short")
	}
	if string(data[:4]) != magic {
		return nil, errors.New("invalid vault header")
	}

	salt := data[4 : 4+32]
	nonce := data[4+32 : 4+32+12]
	ciphertext := data[4+32+12:]
	return pkgcrypto.Decrypt(salt, nonce, ciphertext, password)
}
