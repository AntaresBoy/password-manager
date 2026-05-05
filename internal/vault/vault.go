package vault

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"passmgr/internal/errno"
	"passmgr/internal/store"
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
	v.data = &VaultData{Version: 1, Entries: []Entry{}, ModifiedAt: time.Now()}
	return v.Save(password)
}

func (v *Vault) Open(password string) error {
	if !v.store.Exists() {
		return errno.ErrVaultNotFound
	}
	raw, err := v.store.Read()
	if err != nil {
		return errno.ErrVaultCorrupted.WithCause(err)
	}
	plaintext, err := decryptVault(raw, password)
	if err != nil {
		return errno.ErrWrongPassword.WithCause(err)
	}
	var data VaultData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return errno.ErrVaultCorrupted.WithCause(err)
	}
	v.data = &data
	return nil
}

func (v *Vault) Save(password string) error {
	if v.data == nil {
		return errno.ErrInternal
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

func (v *Vault) AddEntry(entry Entry) error {
	if v.FindEntry(entry.Name) != nil {
		return errno.ErrEntryExists
	}
	now := time.Now()
	if entry.ID == "" {
		entry.ID = newID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	v.data.Entries = append(v.data.Entries, entry)
	return nil
}

func (v *Vault) FindEntry(name string) *Entry {
	if v.data == nil {
		return nil
	}
	for i := range v.data.Entries {
		if v.data.Entries[i].Name == name || v.data.Entries[i].ID == name {
			return &v.data.Entries[i]
		}
	}
	return nil
}

func (v *Vault) RemoveEntry(name string) error {
	if v.data == nil {
		return errno.ErrEntryNotFound
	}
	for i := range v.data.Entries {
		if v.data.Entries[i].Name == name || v.data.Entries[i].ID == name {
			v.data.Entries = append(v.data.Entries[:i], v.data.Entries[i+1:]...)
			return nil
		}
	}
	return errno.ErrEntryNotFound
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b[:])
}
