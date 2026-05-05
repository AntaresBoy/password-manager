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

func TestVaultPathIgnoresEmptyEnvironmentOverride(t *testing.T) {
	t.Setenv("PASSMGR_VAULT_PATH", "")
	t.Setenv("HOME", filepath.Join("tmp", "home"))
	t.Setenv("XDG_DATA_HOME", filepath.Join("tmp", "xdg-data"))
	t.Setenv("APPDATA", filepath.Join("tmp", "appdata"))
	t.Setenv("USERPROFILE", filepath.Join("tmp", "userprofile"))

	got := VaultPath()
	want := defaultVaultPathForTest()

	if got != want {
		t.Fatalf("VaultPath() with empty override = %q, want default %q", got, want)
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

func TestDefaultVaultPathFallbacks(t *testing.T) {
	tests := []struct {
		name string
		goos string
		env  map[string]string
		want string
	}{
		{
			name: "darwin uses home application support",
			goos: "darwin",
			env:  map[string]string{"HOME": filepath.Join("tmp", "home")},
			want: filepath.Join("tmp", "home", "Library", "Application Support", "passmgr", "vault.dat"),
		},
		{
			name: "windows uses appdata",
			goos: "windows",
			env: map[string]string{
				"APPDATA":     filepath.Join("tmp", "appdata"),
				"USERPROFILE": filepath.Join("tmp", "userprofile"),
			},
			want: filepath.Join("tmp", "appdata", "passmgr", "vault.dat"),
		},
		{
			name: "windows falls back to userprofile roaming",
			goos: "windows",
			env:  map[string]string{"USERPROFILE": filepath.Join("tmp", "userprofile")},
			want: filepath.Join("tmp", "userprofile", "AppData", "Roaming", "passmgr", "vault.dat"),
		},
		{
			name: "linux uses xdg data home",
			goos: "linux",
			env: map[string]string{
				"HOME":          filepath.Join("tmp", "home"),
				"XDG_DATA_HOME": filepath.Join("tmp", "xdg-data"),
			},
			want: filepath.Join("tmp", "xdg-data", "passmgr", "vault.dat"),
		},
		{
			name: "linux falls back to home local share",
			goos: "linux",
			env:  map[string]string{"HOME": filepath.Join("tmp", "home")},
			want: filepath.Join("tmp", "home", ".local", "share", "passmgr", "vault.dat"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultVaultPath(tt.goos, func(key string) string {
				return tt.env[key]
			})

			if got != tt.want {
				t.Fatalf("defaultVaultPath() = %q, want %q", got, tt.want)
			}
		})
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
