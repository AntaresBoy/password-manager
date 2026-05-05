package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// VaultPath returns the configured vault path or the platform default.
func VaultPath() string {
	if path := os.Getenv("PASSMGR_VAULT_PATH"); path != "" {
		return path
	}

	return defaultVaultPath(runtime.GOOS, os.Getenv)
}

func defaultVaultPath(goos string, getenv func(string) string) string {
	var base string
	switch goos {
	case "darwin":
		base = filepath.Join(getenv("HOME"), "Library", "Application Support")
	case "windows":
		base = getenv("APPDATA")
		if base == "" {
			base = filepath.Join(getenv("USERPROFILE"), "AppData", "Roaming")
		}
	default:
		base = getenv("XDG_DATA_HOME")
		if base == "" {
			base = filepath.Join(getenv("HOME"), ".local", "share")
		}
	}

	return filepath.Join(base, "passmgr", "vault.dat")
}
