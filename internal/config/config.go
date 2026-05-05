package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// VaultPath returns the configured vault path or the platform default.
func VaultPath() string {
	if path, ok := os.LookupEnv("PASSMGR_VAULT_PATH"); ok {
		return path
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
	default:
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			base = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
	}

	return filepath.Join(base, "passmgr", "vault.dat")
}
