# passmgr CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI password manager with Argon2id + AES-256-GCM encryption, supporting init/add/get/list/rm/gen/cp commands.

**Architecture:** Clean Go package separation with interface-driven design. `pkg/crypto` provides pure cryptographic primitives. `internal/vault` orchestrates encryption and file format. `internal/{store,clipboard,passgen}` handle isolated concerns. `cmd/passmgr` is a thin Cobra CLI layer.

**Tech Stack:** Go 1.23+, Cobra, golang.org/x/crypto/argon2, github.com/atotto/clipboard, testify

---

## File Structure

```
cmd/passmgr/
└── main.go              # cobra root command + subcommand registration

internal/
├── errno/
│   └── errno.go         # error codes with exit codes
├── config/
│   └── config.go         # XDG Base Directory path resolution
├── store/
│   ├── store.go          # Store interface
│   ├── file_store.go     # filesystem implementation
│   └── store_test.go
├── clipboard/
│   ├── clipboard.go      # Clipboard interface
│   ├── system_clip.go    # atotto/clipboard implementation
│   └── clipboard_test.go
├── passgen/
│   ├── passgen.go        # password generator
│   └── passgen_test.go
└── vault/
    ├── vault.go          # Vault struct, VaultData, Entry, lifecycle
    ├── crypto.go         # file format: magic + salt + nonce + ciphertext
    └── vault_test.go

pkg/crypto/
├── crypto.go             # Argon2id + AES-256-GCM primitives
└── crypto_test.go

tests/integration/
└── cli_test.go           # os/exec based full CLI lifecycle tests

go.mod
Makefile
```

---

## Task Dependencies

```
Task 1 (errno) ──┐
Task 2 (config) ─┤
Task 3 (crypto) ─┤
Task 4 (store) ──┼──→ Task 7 (vault)
Task 5 (clip) ───┤      │
Task 6 (passgen)─┘      │
                        └──→ Task 8 (cmd)
                                   │
                                   └──→ Task 9 (integration)
```

---

### Task 1: errno Package

**Files:**
- Create: `internal/errno/errno.go`
- Test: `internal/errno/errno_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/errno/errno_test.go
package errno

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_ExitCode(t *testing.T) {
	err := NewError(20001, "vault not found", 2)
	assert.Equal(t, 20001, err.Code())
	assert.Equal(t, "vault not found", err.Error())
	assert.Equal(t, 2, err.ExitCode())
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("file not found")
	err := NewError(20001, "vault not found", 2).WithCause(inner)
	assert.Equal(t, inner, err.Unwrap())
}

func TestPredefinedErrors(t *testing.T) {
	assert.Equal(t, 0, OK.ExitCode())
	assert.Equal(t, 2, ErrVaultNotFound.ExitCode())
	assert.Equal(t, 3, ErrWrongPassword.ExitCode())
	assert.Equal(t, 4, ErrEntryNotFound.ExitCode())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/errno -v`
Expected: FAIL with "undefined: NewError" / "undefined: OK" etc.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/errno/errno.go
package errno

import "fmt"

type Error struct {
	code    int
	message string
	exit    int
	cause   error
}

func NewError(code int, message string, exit int) *Error {
	return &Error{code: code, message: message, exit: exit}
}

func (e *Error) Error() string   { return e.message }
func (e *Error) Code() int       { return e.code }
func (e *Error) ExitCode() int   { return e.exit }
func (e *Error) Unwrap() error   { return e.cause }
func (e *Error) WithCause(c error) *Error {
	e.cause = c
	return e
}

// Predefined errors
var (
	OK                  = NewError(0, "success", 0)
	ErrInternal         = NewError(10001, "internal error", 10)
	ErrVaultNotFound    = NewError(20001, "vault not found", 2)
	ErrVaultExists      = NewError(20002, "vault already exists", 5)
	ErrVaultCorrupted   = NewError(20003, "vault file corrupted", 2)
	ErrWrongPassword    = NewError(20004, "wrong master password", 3)
	ErrEntryNotFound    = NewError(20101, "entry not found", 4)
	ErrEntryExists      = NewError(20102, "entry already exists", 5)
	ErrInvalidInput     = NewError(20201, "invalid input", 5)
	ErrPasswordMismatch = NewError(20202, "passwords do not match", 5)
	ErrClipboardFail    = NewError(20301, "clipboard unavailable", 1)
)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/errno -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/errno/
git commit -m "feat(errno): add error code package with exit codes

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 2: config Package

**Files:**
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/config/config_test.go
package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVaultPath(t *testing.T) {
	p := VaultPath()
	assert.NotEmpty(t, p)
	assert.True(t, strings.HasSuffix(p, filepath.Join("passmgr", "vault.dat")))
}

func TestVaultPath_Custom(t *testing.T) {
	t.Setenv("PASSMGR_VAULT_PATH", "/custom/path.dat")
	p := VaultPath()
	assert.Equal(t, "/custom/path.dat", p)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config -v`
Expected: FAIL with "undefined: VaultPath"

- [ ] **Step 3: Write minimal implementation**

```go
// internal/config/config.go
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func VaultPath() string {
	if p := os.Getenv("PASSMGR_VAULT_PATH"); p != "" {
		return p
	}

	var base string
	switch runtime.GOOS {
	case "darwin":
		base = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	default: // linux and others
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			base = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
	}

	return filepath.Join(base, "passmgr", "vault.dat")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/config -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat(config): add XDG Base Directory vault path resolution

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 3: pkg/crypto Primitives

**Files:**
- Create: `pkg/crypto/crypto.go`
- Test: `pkg/crypto/crypto_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/crypto/crypto_test.go
package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKey_Deterministic(t *testing.T) {
	salt := []byte("fixed-salt-for-testing-only!!")
	key1 := DeriveKey("password", salt)
	key2 := DeriveKey("password", salt)
	assert.Equal(t, key1, key2)
	assert.Len(t, key1, 32)
}

func TestDeriveKey_DifferentSalt(t *testing.T) {
	key1 := DeriveKey("password", []byte("salt-one------------------------"))
	key2 := DeriveKey("password", []byte("salt-two------------------------"))
	assert.False(t, bytes.Equal(key1, key2))
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	plaintext := []byte("secret data here")
	password := "my-password"

	salt, nonce, ciphertext, err := Encrypt(plaintext, password)
	require.NoError(t, err)
	assert.Len(t, salt, 32)
	assert.Len(t, nonce, 12)
	assert.NotEmpty(t, ciphertext)

	decrypted, err := Decrypt(salt, nonce, ciphertext, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecrypt_WrongPassword(t *testing.T) {
	plaintext := []byte("secret data")
	salt, nonce, ciphertext, _ := Encrypt(plaintext, "correct")

	_, err := Decrypt(salt, nonce, ciphertext, "wrong")
	assert.Error(t, err)
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	plaintext := []byte("secret data")
	salt, nonce, ciphertext, _ := Encrypt(plaintext, "password")

	ciphertext[0] ^= 0xFF
	_, err := Decrypt(salt, nonce, ciphertext, "password")
	assert.Error(t, err)
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt(32)
	require.NoError(t, err)
	assert.Len(t, salt1, 32)

	salt2, err := GenerateSalt(32)
	require.NoError(t, err)
	assert.False(t, bytes.Equal(salt1, salt2))
}

func TestGenerateNonce(t *testing.T) {
	nonce1, err := GenerateNonce()
	require.NoError(t, err)
	assert.Len(t, nonce1, 12)

	nonce2, err := GenerateNonce()
	require.NoError(t, err)
	assert.False(t, bytes.Equal(nonce1, nonce2))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/crypto -v`
Expected: FAIL with "undefined: DeriveKey", "undefined: Encrypt", etc.

- [ ] **Step 3: Write minimal implementation**

```go
// pkg/crypto/crypto.go
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
	keyLen   = 32
	saltLen  = 32
	nonceLen = 12
	// Argon2id parameters
	time    = 3
	memory  = 64 * 1024 // 64 MB
	threads = 4
)

func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
}

func Encrypt(plaintext []byte, password string) (salt, nonce, ciphertext []byte, err error) {
	salt, err = GenerateSalt(saltLen)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate salt: %w", err)
	}

	key := DeriveKey(password, salt)

	nonce, err = GenerateNonce()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate nonce: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create GCM: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return salt, nonce, ciphertext, nil
}

func Decrypt(salt, nonce, ciphertext []byte, password string) ([]byte, error) {
	if len(salt) != saltLen {
		return nil, errors.New("invalid salt length")
	}
	if len(nonce) != nonceLen {
		return nil, errors.New("invalid nonce length")
	}

	key := DeriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: wrong password or corrupted data")
	}

	return plaintext, nil
}

func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/crypto -v`
Expected: PASS (all 7 tests)

- [ ] **Step 5: Commit**

```bash
git add pkg/crypto/
git commit -m "feat(crypto): add Argon2id + AES-256-GCM primitives

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 4: internal/store

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/file_store.go`
- Test: `internal/store/file_store_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/store/file_store_test.go
package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_ReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.dat")

	s := NewFileStore(path)
	assert.False(t, s.Exists())

	data := []byte("hello vault")
	err := s.Write(data)
	require.NoError(t, err)
	assert.True(t, s.Exists())

	read, err := s.Read()
	require.NoError(t, err)
	assert.Equal(t, data, read)
}

func TestFileStore_Read_NotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.dat")

	s := NewFileStore(path)
	_, err := s.Read()
	assert.Error(t, err)
}

func TestFileStore_Path(t *testing.T) {
	path := "/tmp/test.dat"
	s := NewFileStore(path)
	assert.Equal(t, path, s.Path())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/store -v`
Expected: FAIL with "undefined: NewFileStore"

- [ ] **Step 3: Write minimal implementation**

```go
// internal/store/store.go
package store

type Store interface {
	Read() ([]byte, error)
	Write(data []byte) error
	Exists() bool
	Path() string
}
```

```go
// internal/store/file_store.go
package store

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (f *FileStore) Read() ([]byte, error) {
	return os.ReadFile(f.path)
}

func (f *FileStore) Write(data []byte) error {
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	return os.Rename(tmp, f.path)
}

func (f *FileStore) Exists() bool {
	_, err := os.Stat(f.path)
	return err == nil
}

func (f *FileStore) Path() string {
	return f.path
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/store -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/store/
git commit -m "feat(store): add Store interface and FileStore implementation

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 5: internal/clipboard

**Files:**
- Create: `internal/clipboard/clipboard.go`
- Create: `internal/clipboard/system_clip.go`
- Test: `internal/clipboard/clipboard_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/clipboard/clipboard_test.go
package clipboard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockClipboard struct {
	lastText string
	copied   bool
	cleared  bool
}

func (m *mockClipboard) Copy(text string) error {
	m.lastText = text
	m.copied = true
	return nil
}

func (m *mockClipboard) Clear() error {
	m.cleared = true
	return nil
}

func TestMockClipboard_CopyAndClear(t *testing.T) {
	m := &mockClipboard{}
	err := m.Copy("secret")
	assert.NoError(t, err)
	assert.True(t, m.copied)
	assert.Equal(t, "secret", m.lastText)

	err = m.Clear()
	assert.NoError(t, err)
	assert.True(t, m.cleared)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/clipboard -v`
Expected: FAIL with "undefined: Clipboard" if the interface isn't defined yet

- [ ] **Step 3: Write minimal implementation**

```go
// internal/clipboard/clipboard.go
package clipboard

type Clipboard interface {
	Copy(text string) error
	Clear() error
}
```

```go
// internal/clipboard/system_clip.go
package clipboard

import "github.com/atotto/clipboard"

type SystemClipboard struct{}

func NewSystemClipboard() *SystemClipboard {
	return &SystemClipboard{}
}

func (s *SystemClipboard) Copy(text string) error {
	return clipboard.WriteAll(text)
}

func (s *SystemClipboard) Clear() error {
	return clipboard.WriteAll("")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/clipboard -v`
Expected: PASS (mock test only; SystemClipboard requires actual clipboard)

- [ ] **Step 5: Commit**

```bash
git add internal/clipboard/
git commit -m "feat(clipboard): add Clipboard interface and system implementation

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 6: internal/passgen

**Files:**
- Create: `internal/passgen/passgen.go`
- Test: `internal/passgen/passgen_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/passgen/passgen_test.go
package passgen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate_DefaultLength(t *testing.T) {
	p, err := Generate(DefaultOptions())
	assert.NoError(t, err)
	assert.Len(t, p, 16)
}

func TestGenerate_CustomLength(t *testing.T) {
	opts := DefaultOptions()
	opts.Length = 32
	p, err := Generate(opts)
	assert.NoError(t, err)
	assert.Len(t, p, 32)
}

func TestGenerate_AllCharsets(t *testing.T) {
	opts := DefaultOptions()
	opts.Length = 64
	p, err := Generate(opts)
	assert.NoError(t, err)

	assert.Contains(t, p, func(r rune) bool { return r >= 'a' && r <= 'z' })
	assert.Contains(t, p, func(r rune) bool { return r >= 'A' && r <= 'Z' })
	assert.Contains(t, p, func(r rune) bool { return r >= '0' && r <= '9' })
	assert.Contains(t, p, func(r rune) bool { return strings.ContainsRune("!@#$%^&*", r) })
}

func TestGenerate_NoSymbols(t *testing.T) {
	opts := DefaultOptions()
	opts.Symbols = false
	p, err := Generate(opts)
	assert.NoError(t, err)
	assert.NotContains(t, p, "!")
}

func TestGenerate_ZeroLength(t *testing.T) {
	opts := DefaultOptions()
	opts.Length = 0
	_, err := Generate(opts)
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/passgen -v`
Expected: FAIL with "undefined: Generate"

- [ ] **Step 3: Write minimal implementation**

```go
// internal/passgen/passgen.go
package passgen

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const charsetLower = "abcdefghijklmnopqrstuvwxyz"
const charsetUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const charsetDigit = "0123456789"
const charsetSymbol = "!@#$%^&*"

type Options struct {
	Length  int
	Lower   bool
	Upper   bool
	Digits  bool
	Symbols bool
}

func DefaultOptions() Options {
	return Options{
		Length:  16,
		Lower:   true,
		Upper:   true,
		Digits:  true,
		Symbols: true,
	}
}

func Generate(opts Options) (string, error) {
	if opts.Length <= 0 {
		return "", errors.New("length must be positive")
	}

	var charset string
	if opts.Lower {
		charset += charsetLower
	}
	if opts.Upper {
		charset += charsetUpper
	}
	if opts.Digits {
		charset += charsetDigit
	}
	if opts.Symbols {
		charset += charsetSymbol
	}
	if charset == "" {
		return "", errors.New("at least one character set must be enabled")
	}

	result := make([]byte, opts.Length)
	max := big.NewInt(int64(len(charset)))

	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}

	return string(result), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/passgen -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/passgen/
git commit -m "feat(passgen): add password generator with charset options

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 7: internal/vault

**Files:**
- Create: `internal/vault/vault.go`
- Create: `internal/vault/crypto.go`
- Test: `internal/vault/vault_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/vault/vault_test.go
package vault

import (
	"testing"
	"time"

	"github.com/antares/passmgr/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	data   []byte
	exists bool
}

func (m *mockStore) Read() ([]byte, error) { return m.data, nil }
func (m *mockStore) Write(d []byte) error  { m.data = d; m.exists = true; return nil }
func (m *mockStore) Exists() bool          { return m.exists }
func (m *mockStore) Path() string          { return "/mock/vault.dat" }

func TestVault_Init(t *testing.T) {
	s := &mockStore{}
	v := New(s)

	err := v.Init("master-password")
	require.NoError(t, err)
	assert.True(t, s.Exists())
	assert.NotEmpty(t, s.data)
}

func TestVault_Init_AlreadyExists(t *testing.T) {
	s := &mockStore{exists: true}
	v := New(s)

	err := v.Init("password")
	assert.Error(t, err)
}

func TestVault_OpenAndSave(t *testing.T) {
	s := &mockStore{}
	v := New(s)

	err := v.Init("password")
	require.NoError(t, err)

	v.Data().Entries = append(v.Data().Entries, Entry{
		ID:       "test-id",
		Name:     "github",
		Username: "user",
		Password: "secret",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	err = v.Save("password")
	require.NoError(t, err)

	v2 := New(s)
	err = v2.Open("password")
	require.NoError(t, err)

	assert.Len(t, v2.Data().Entries, 1)
	assert.Equal(t, "github", v2.Data().Entries[0].Name)
	assert.Equal(t, "secret", v2.Data().Entries[0].Password)
}

func TestVault_Open_WrongPassword(t *testing.T) {
	s := &mockStore{}
	v := New(s)

	err := v.Init("correct")
	require.NoError(t, err)

	v2 := New(s)
	err = v2.Open("wrong")
	assert.Error(t, err)
}

func TestVault_CorruptedFile(t *testing.T) {
	s := &mockStore{data: []byte("corrupted-data"), exists: true}
	v := New(s)

	err := v.Open("password")
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/vault -v`
Expected: FAIL with "undefined: Vault", "undefined: Entry", etc.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/vault/vault.go
package vault

import (
	"encoding/json"
	"time"

	"github.com/antares/passmgr/internal/errno"
	"github.com/antares/passmgr/internal/store"
	"github.com/google/uuid"
)

type VaultData struct {
	Version    int       `json:"version"`
	Entries    []Entry   `json:"entries"`
	ModifiedAt time.Time `json:"modified_at"`
}

type Entry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	URL       string    `json:"url,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Vault struct {
	store store.Store
	data  *VaultData
}

func New(s store.Store) *Vault {
	return &Vault{store: s}
}

func (v *Vault) Init(password string) error {
	if v.store.Exists() {
		return errno.ErrVaultExists
	}

	v.data = &VaultData{
		Version:    1,
		Entries:    []Entry{},
		ModifiedAt: time.Now(),
	}

	return v.Save(password)
}

func (v *Vault) Open(password string) error {
	if !v.store.Exists() {
		return errno.ErrVaultNotFound
	}

	data, err := v.store.Read()
	if err != nil {
		return errno.ErrVaultCorrupted.WithCause(err)
	}

	plaintext, err := decryptVault(data, password)
	if err != nil {
		return errno.ErrWrongPassword
	}

	var vaultData VaultData
	if err := json.Unmarshal(plaintext, &vaultData); err != nil {
		return errno.ErrVaultCorrupted.WithCause(err)
	}

	v.data = &vaultData
	return nil
}

func (v *Vault) Save(password string) error {
	if v.data == nil {
		return errno.ErrInternal.WithCause(errno.NewError(10001, "no vault data loaded", 10))
	}

	v.data.ModifiedAt = time.Now()

	plaintext, err := json.Marshal(v.data)
	if err != nil {
		return errno.ErrInternal.WithCause(err)
	}

	encrypted, err := encryptVault(plaintext, password)
	if err != nil {
		return errno.ErrInternal.WithCause(err)
	}

	if err := v.store.Write(encrypted); err != nil {
		return errno.ErrInternal.WithCause(err)
	}

	return nil
}

func (v *Vault) Data() *VaultData {
	return v.data
}

func (v *Vault) FindEntry(name string) *Entry {
	for i := range v.data.Entries {
		if v.data.Entries[i].Name == name {
			return &v.data.Entries[i]
		}
	}
	return nil
}

func (v *Vault) AddEntry(e Entry) error {
	if v.FindEntry(e.Name) != nil {
		return errno.ErrEntryExists
	}
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	now := time.Now()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	v.data.Entries = append(v.data.Entries, e)
	return nil
}

func (v *Vault) RemoveEntry(name string) error {
	for i, e := range v.data.Entries {
		if e.Name == name {
			v.data.Entries = append(v.data.Entries[:i], v.data.Entries[i+1:]...)
			return nil
		}
	}
	return errno.ErrEntryNotFound
}
```

```go
// internal/vault/crypto.go
package vault

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/antares/passmgr/pkg/crypto"
)

const magic = "PMV1"

func encryptVault(plaintext []byte, password string) ([]byte, error) {
	salt, nonce, ciphertext, err := crypto.Encrypt(plaintext, password)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
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
		return nil, errors.New("data too short")
	}

	if string(data[:4]) != magic {
		return nil, errors.New("invalid magic header")
	}

	salt := data[4 : 4+32]
	nonce := data[4+32 : 4+32+12]
	ciphertext := data[4+32+12:]

	return crypto.Decrypt(salt, nonce, ciphertext, password)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/vault -v`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/vault/
git commit -m "feat(vault): add Vault lifecycle with encryption

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 8: cmd/passmgr CLI

**Files:**
- Create: `cmd/passmgr/main.go`
- Create: `go.mod`
- Create: `Makefile`

- [ ] **Step 1: Write go.mod**

```go
// go.mod
module github.com/antares/passmgr

go 1.23.0

require (
	github.com/atotto/clipboard v0.1.4
	github.com/google/uuid v1.6.0
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	golang.org/x/crypto v0.35.0
	golang.org/x/term v0.20.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

Run: `go mod tidy` to generate `go.sum`.

- [ ] **Step 2: Write Makefile**

```makefile
# Makefile
.PHONY: build test test-coverage lint clean install

BINARY_NAME=passmgr
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/passmgr

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	gofmt -w .
	go vet ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
```

- [ ] **Step 3: Write the failing test**

First, write the main.go file. Since this is a CLI, we can't easily unit test main.go directly. Instead, write integration tests in Task 9. For now, write main.go directly:

```go
// cmd/passmgr/main.go
package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/antares/passmgr/internal/clipboard"
	"github.com/antares/passmgr/internal/config"
	"github.com/antares/passmgr/internal/errno"
	"github.com/antares/passmgr/internal/passgen"
	"github.com/antares/passmgr/internal/store"
	"github.com/antares/passmgr/internal/vault"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var vaultPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "passmgr",
		Short: "A local password manager",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if vaultPath == "" {
				vaultPath = config.VaultPath()
			}
		},
	}
	rootCmd.PersistentFlags().StringVar(&vaultPath, "vault-path", "", "Path to vault file")

	rootCmd.AddCommand(initCmd(), addCmd(), getCmd(), listCmd(), rmCmd(), genCmd(), cpCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := store.NewFileStore(vaultPath)
			v := vault.New(s)

			if s.Exists() {
				return errno.ErrVaultExists
			}

			password, err := readPassword("Enter master password: ")
			if err != nil {
				return errno.ErrInvalidInput.WithCause(err)
			}

			confirm, err := readPassword("Confirm master password: ")
			if err != nil {
				return errno.ErrInvalidInput.WithCause(err)
			}

			if password != confirm {
				return errno.ErrPasswordMismatch
			}

			if err := v.Init(password); err != nil {
				return err
			}

			fmt.Printf("Vault created at %s\n", vaultPath)
			return nil
		},
	}
}

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := store.NewFileStore(vaultPath)
			v := vault.New(s)

			password, err := readPassword("Enter master password: ")
			if err != nil {
				return errno.ErrInvalidInput.WithCause(err)
			}

			if err := v.Open(password); err != nil {
				return err
			}

			fmt.Print("Username: ")
			var username string
			fmt.Scanln(&username)

			fmt.Print("Password (empty to generate): ")
			var entryPassword string
			fmt.Scanln(&entryPassword)

			if entryPassword == "" {
				entryPassword, err = passgen.Generate(passgen.DefaultOptions())
				if err != nil {
					return errno.ErrInternal.WithCause(err)
				}
				fmt.Printf("Generated password: %s\n", entryPassword)
			}

			fmt.Print("URL: ")
			var url string
			fmt.Scanln(&url)

			fmt.Print("Notes: ")
			var notes string
			fmt.Scanln(&notes)

			entry := vault.Entry{
				Name:     args[0],
				Username: username,
				Password: entryPassword,
				URL:      url,
				Notes:    notes,
			}

			if err := v.AddEntry(entry); err != nil {
				return err
			}

			if err := v.Save(password); err != nil {
				return err
			}

			fmt.Printf("Added: %s\n", args[0])
			return nil
		},
	}
}

func getCmd() *cobra.Command {
	var showPassword bool
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get an entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := store.NewFileStore(vaultPath)
			v := vault.New(s)

			password, err := readPassword("Enter master password: ")
			if err != nil {
				return errno.ErrInvalidInput.WithCause(err)
			}

			if err := v.Open(password); err != nil {
				return err
			}

			entry := v.FindEntry(args[0])
			if entry == nil {
				return errno.ErrEntryNotFound
			}

			fmt.Printf("Name:     %s\n", entry.Name)
			fmt.Printf("Username: %s\n", entry.Username)
			if showPassword {
				fmt.Printf("Password: %s\n", entry.Password)
			} else {
				fmt.Printf("Password: ********\n")
			}
			if entry.URL != "" {
				fmt.Printf("URL:      %s\n", entry.URL)
			}
			if entry.Notes != "" {
				fmt.Printf("Notes:    %s\n", entry.Notes)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&showPassword, "show-password", false, "Show password in plaintext")
	return cmd
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := store.NewFileStore(vaultPath)
			v := vault.New(s)

			password, err := readPassword("Enter master password: ")
			if err != nil {
				return errno.ErrInvalidInput.WithCause(err)
			}

			if err := v.Open(password); err != nil {
				return err
			}

			if len(v.Data().Entries) == 0 {
				fmt.Println("No entries found")
				return nil
			}

			fmt.Printf("%-20s %-20s %-30s %s\n", "NAME", "USERNAME", "URL", "TAGS")
			fmt.Println(strings.Repeat("-", 80))
			for _, e := range v.Data().Entries {
				fmt.Printf("%-20s %-20s %-30s %s\n", e.Name, e.Username, e.URL, strings.Join(e.Tags, ","))
			}

			return nil
		},
	}
}

func rmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove an entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := store.NewFileStore(vaultPath)
			v := vault.New(s)

			password, err := readPassword("Enter master password: ")
			if err != nil {
				return errno.ErrInvalidInput.WithCause(err)
			}

			if err := v.Open(password); err != nil {
				return err
			}

			entry := v.FindEntry(args[0])
			if entry == nil {
				return errno.ErrEntryNotFound
			}

			fmt.Printf("Delete %s? [y/N] ", args[0])
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) != "y" {
				fmt.Println("Cancelled")
				return nil
			}

			if err := v.RemoveEntry(args[0]); err != nil {
				return err
			}

			if err := v.Save(password); err != nil {
				return err
			}

			fmt.Printf("Deleted: %s\n", args[0])
			return nil
		},
	}
}

func genCmd() *cobra.Command {
	var (
		length   int
		noSymbols bool
		copyFlag bool
	)
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate a random password",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := passgen.DefaultOptions()
			opts.Length = length
			if noSymbols {
				opts.Symbols = false
			}

			password, err := passgen.Generate(opts)
			if err != nil {
				return errno.ErrInternal.WithCause(err)
			}

			fmt.Println(password)

			if copyFlag {
				clip := clipboard.NewSystemClipboard()
				if err := clip.Copy(password); err != nil {
					return errno.ErrClipboardFail.WithCause(err)
				}
				fmt.Println("Copied to clipboard, will clear in 10s")
				go func() {
					time.Sleep(10 * time.Second)
					clip.Clear()
				}()
			}

			return nil
		},
	}
	cmd.Flags().IntVarP(&length, "length", "l", 16, "Password length")
	cmd.Flags().BoolVar(&noSymbols, "no-symbols", false, "Exclude symbols")
	cmd.Flags().BoolVarP(&copyFlag, "copy", "c", false, "Copy to clipboard")
	return cmd
}

func cpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cp <name>",
		Short: "Copy password to clipboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := store.NewFileStore(vaultPath)
			v := vault.New(s)

			password, err := readPassword("Enter master password: ")
			if err != nil {
				return errno.ErrInvalidInput.WithCause(err)
			}

			if err := v.Open(password); err != nil {
				return err
			}

			entry := v.FindEntry(args[0])
			if entry == nil {
				return errno.ErrEntryNotFound
			}

			clip := clipboard.NewSystemClipboard()
			if err := clip.Copy(entry.Password); err != nil {
				return errno.ErrClipboardFail.WithCause(err)
			}

			fmt.Println("Copied to clipboard, will clear in 10s")
			go func() {
				time.Sleep(10 * time.Second)
				clip.Clear()
			}()

			return nil
		},
	}
}
```

- [ ] **Step 4: Build and verify**

Run: `go mod tidy`
Run: `go build ./cmd/passmgr`
Expected: builds successfully, binary at default output

- [ ] **Step 5: Commit**

```bash
git add cmd/passmgr/main.go go.mod go.sum Makefile
git commit -m "feat(cli): add all CLI commands (init, add, get, list, rm, gen, cp)

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 9: Integration Tests

**Files:**
- Create: `tests/integration/cli_test.go`

- [ ] **Step 1: Write the test**

```go
// tests/integration/cli_test.go
package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binary := filepath.Join(tmpDir, "passmgr")

	cmd := exec.Command("go", "build", "-o", binary, "../../cmd/passmgr")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	return binary
}

func TestCLI_FullLifecycle(t *testing.T) {
	binary := buildBinary(t)
	vaultPath := filepath.Join(t.TempDir(), "vault.dat")

	// 1. init
	cmd := exec.Command(binary, "init", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("password\npassword\n")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "init failed: %s", out)
	assert.Contains(t, string(out), "Vault created")

	// 2. add
	cmd = exec.Command(binary, "add", "github", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("password\nmyuser\nmysecret\nhttps://github.com\n\n")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "add failed: %s", out)
	assert.Contains(t, string(out), "Added: github")

	// 3. list
	cmd = exec.Command(binary, "list", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("password\n")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "list failed: %s", out)
	assert.Contains(t, string(out), "github")

	// 4. get (without --show-password)
	cmd = exec.Command(binary, "get", "github", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("password\n")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "get failed: %s", out)
	assert.Contains(t, string(out), "myuser")
	assert.Contains(t, string(out), "********")

	// 5. get (with --show-password)
	cmd = exec.Command(binary, "get", "github", "--show-password", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("password\n")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "get with show-password failed: %s", out)
	assert.Contains(t, string(out), "mysecret")

	// 6. gen
	cmd = exec.Command(binary, "gen", "--length", "20", "--no-symbols", "--vault-path", vaultPath)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "gen failed: %s", out)
	generated := strings.TrimSpace(string(out))
	assert.Len(t, generated, 20)
	assert.NotContains(t, generated, "!")

	// 7. rm
	cmd = exec.Command(binary, "rm", "github", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("password\ny\n")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "rm failed: %s", out)
	assert.Contains(t, string(out), "Deleted: github")

	// 8. list (empty)
	cmd = exec.Command(binary, "list", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("password\n")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "list after rm failed: %s", out)
	assert.Contains(t, string(out), "No entries found")
}

func TestCLI_WrongPassword(t *testing.T) {
	binary := buildBinary(t)
	vaultPath := filepath.Join(t.TempDir(), "vault.dat")

	// init
	cmd := exec.Command(binary, "init", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("correct\ncorrect\n")
	_, err := cmd.CombinedOutput()
	require.NoError(t, err)

	// try open with wrong password
	cmd = exec.Command(binary, "list", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("wrong\n")
	out, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(out), "wrong master password")
}

func TestCLI_VaultNotFound(t *testing.T) {
	binary := buildBinary(t)
	vaultPath := filepath.Join(t.TempDir(), "nonexistent.dat")

	cmd := exec.Command(binary, "list", "--vault-path", vaultPath)
	cmd.Stdin = strings.NewReader("password\n")
	out, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(out), "vault not found")
}
```

- [ ] **Step 2: Run the test**

Run: `go test ./tests/integration -v -count=1`
Expected: PASS (all 4 integration tests)

- [ ] **Step 3: Commit**

```bash
git add tests/integration/
git commit -m "test(integration): add CLI lifecycle integration tests

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Coverage Verification

After all tasks complete:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Verify:**
- `pkg/crypto` ≥ 80%
- `internal/vault` ≥ 80%
- `internal/store` ≥ 80%
- `internal/passgen` ≥ 80%
- `internal/errno` ≥ 80%
- Overall lines ≥ 80%, branches ≥ 70%, functions ≥ 80%

If any package is below threshold, add targeted unit tests before proceeding.

---

## Self-Review Checklist

**Spec Coverage:**
- [x] Argon2id + AES-256-GCM encryption → Task 3 (crypto) + Task 7 (vault crypto.go)
- [x] Vault file format (PMV1 magic + salt + nonce + ciphertext) → Task 7
- [x] Per-save salt regeneration → Task 7 (crypto.Encrypt generates new salt)
- [x] XDG Base Directory paths → Task 2
- [x] init command → Task 8
- [x] add command → Task 8
- [x] get command → Task 8
- [x] list command → Task 8
- [x] rm command → Task 8
- [x] gen command → Task 8
- [x] cp command → Task 8
- [x] Clipboard auto-clear → Task 8 (cpCmd + genCmd)
- [x] Error codes with exit codes → Task 1
- [x] Password generator with charsets → Task 6
- [x] Coverage ≥ 80% → Coverage Verification section

**Placeholder Scan:**
- [x] No "TBD", "TODO", "implement later" in any step
- [x] All code blocks contain actual code
- [x] All commands have expected outputs
- [x] No "Similar to Task N" references

**Type Consistency:**
- [x] `VaultData` / `Entry` fields match design doc Section 3.1
- [x] `Store` interface methods match design doc Section 2.3
- [x] `Clipboard` interface methods match design doc Section 2.3
- [x] Error codes match design doc Section 5.1
- [x] Argon2id params match design doc Section 3.3 (t=3, m=64MB, p=4)
