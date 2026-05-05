package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVaultPathUsesEnvironmentOverrideExactly(t *testing.T) {
	custom := "relative/custom-vault.dat"
	t.Setenv("PASSMGR_VAULT_PATH", custom)

	got := VaultPath()

	if got != custom {
		t.Fatalf("VaultPath() = %q, want exact override %q", got, custom)
	}
}

func TestVaultPathDefaultForCurrentPlatform(t *testing.T) {
	os.Unsetenv("PASSMGR_VAULT_PATH")
	t.Cleanup(func() { os.Unsetenv("PASSMGR_VAULT_PATH") })
	t.Setenv("HOME", filepath.Join("tmp", "home"))
	t.Setenv("XDG_DATA_HOME", filepath.Join("tmp", "xdg-data"))
	t.Setenv("APPDATA", filepath.Join("tmp", "appdata"))
	t.Setenv("USERPROFILE", filepath.Join("tmp", "userprofile"))

	want := defaultVaultPathForTest()
	got := VaultPath()

	if got != want {
		t.Fatalf("VaultPath() = %q, want %q", got, want)
	}
}

func defaultVaultPathForTest() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join("tmp", "home", "Library", "Application Support", "passmgr", "vault.dat")
	case "windows":
		return filepath.Join("tmp", "appdata", "passmgr", "vault.dat")
	default:
		return filepath.Join("tmp", "xdg-data", "passmgr", "vault.dat")
	}
}
