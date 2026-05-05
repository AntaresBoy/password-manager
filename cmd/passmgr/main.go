package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"passmgr/internal/clipboard"
	"passmgr/internal/config"
	"passmgr/internal/errno"
	"passmgr/internal/passgen"
	"passmgr/internal/store"
	"passmgr/internal/vault"
)

const clearDelay = 10 * time.Second

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if appErr, ok := err.(*errno.Error); ok {
			os.Exit(appErr.ExitCode())
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return nil
	}

	vaultPath := config.VaultPath()
	if args[0] == "--vault-path" && len(args) >= 3 {
		vaultPath = args[1]
		args = args[2:]
	}
	if strings.HasPrefix(args[0], "--vault-path=") {
		vaultPath = strings.TrimPrefix(args[0], "--vault-path=")
		args = args[1:]
	}
	if len(args) == 0 {
		usage()
		return nil
	}

	switch args[0] {
	case "init":
		return runInit(vaultPath)
	case "add":
		return runAdd(vaultPath, args[1:])
	case "list":
		return runList(vaultPath)
	case "get":
		return runGet(vaultPath, args[1:])
	case "rm":
		return runRemove(vaultPath, args[1:])
	case "gen":
		return runGen(args[1:])
	case "cp":
		return runCopy(vaultPath, args[1:])
	default:
		usage()
		return errno.ErrInvalidInput
	}
}

func runInit(vaultPath string) error {
	password, err := readMasterPassword("Enter master password: ")
	if err != nil {
		return err
	}
	confirm := os.Getenv("PASSMGR_MASTER_PASSWORD_CONFIRM")
	if confirm == "" {
		confirm, err = readLine("Confirm master password: ")
		if err != nil {
			return err
		}
	}
	if password != confirm {
		return errno.ErrPasswordMismatch
	}
	v := vault.New(store.NewFileStore(vaultPath))
	if err := v.Init(password); err != nil {
		return err
	}
	fmt.Printf("Vault created at %s\n", vaultPath)
	return nil
}

func runAdd(vaultPath string, args []string) error {
	if len(args) < 1 {
		return errno.ErrInvalidInput
	}
	name := args[0]
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	username := fs.String("u", "", "username")
	passwordFlag := fs.String("p", "", "password")
	url := fs.String("url", "", "url")
	notes := fs.String("notes", "", "notes")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	entryPassword := *passwordFlag
	if entryPassword == "" {
		generated, err := passgen.Generate(passgen.DefaultOptions())
		if err != nil {
			return err
		}
		entryPassword = generated
	}
	v, master, err := openVault(vaultPath)
	if err != nil {
		return err
	}
	if err := v.AddEntry(vault.Entry{
		Name:     name,
		Username: *username,
		Password: entryPassword,
		URL:      *url,
		Notes:    *notes,
	}); err != nil {
		return err
	}
	if err := v.Save(master); err != nil {
		return err
	}
	fmt.Printf("Added: %s\n", name)
	return nil
}

func runList(vaultPath string) error {
	v, _, err := openVault(vaultPath)
	if err != nil {
		return err
	}
	fmt.Println("Name\tUsername\tURL\tTags")
	for _, entry := range v.Data().Entries {
		fmt.Printf("%s\t%s\t%s\t%s\n", entry.Name, entry.Username, entry.URL, strings.Join(entry.Tags, ","))
	}
	return nil
}

func runGet(vaultPath string, args []string) error {
	if len(args) < 1 {
		return errno.ErrInvalidInput
	}
	name := args[0]
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	showPassword := fs.Bool("show-password", false, "show password")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	v, _, err := openVault(vaultPath)
	if err != nil {
		return err
	}
	entry := v.FindEntry(name)
	if entry == nil {
		return errno.ErrEntryNotFound
	}
	password := "********"
	if *showPassword {
		password = entry.Password
	}
	fmt.Printf("Name: %s\nUsername: %s\nPassword: %s\nURL: %s\nNotes: %s\n", entry.Name, entry.Username, password, entry.URL, entry.Notes)
	return nil
}

func runRemove(vaultPath string, args []string) error {
	if len(args) < 1 {
		return errno.ErrInvalidInput
	}
	name := args[0]
	fs := flag.NewFlagSet("rm", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "confirm deletion")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if !*yes {
		answer, err := readLine("Delete " + name + "? [y/N] ")
		if err != nil {
			return err
		}
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			return nil
		}
	}
	v, master, err := openVault(vaultPath)
	if err != nil {
		return err
	}
	if err := v.RemoveEntry(name); err != nil {
		return err
	}
	if err := v.Save(master); err != nil {
		return err
	}
	fmt.Printf("Removed: %s\n", name)
	return nil
}

func runGen(args []string) error {
	fs := flag.NewFlagSet("gen", flag.ContinueOnError)
	length := fs.Int("length", 16, "password length")
	copyResult := fs.Bool("copy", false, "copy to clipboard")
	noSymbols := fs.Bool("no-symbols", false, "disable symbols")
	if err := fs.Parse(args); err != nil {
		return err
	}
	opts := passgen.DefaultOptions()
	opts.Length = *length
	opts.Symbols = !*noSymbols
	password, err := passgen.Generate(opts)
	if err != nil {
		return err
	}
	fmt.Println(password)
	if *copyResult {
		c := clipboard.NewSystemClipboard()
		if err := c.Copy(password); err != nil {
			return errno.ErrClipboardFail.WithCause(err)
		}
		clipboard.ClearAfter(c, clearDelay)
	}
	return nil
}

func runCopy(vaultPath string, args []string) error {
	if len(args) != 1 {
		return errno.ErrInvalidInput
	}
	v, _, err := openVault(vaultPath)
	if err != nil {
		return err
	}
	entry := v.FindEntry(args[0])
	if entry == nil {
		return errno.ErrEntryNotFound
	}
	c := clipboard.NewSystemClipboard()
	if err := c.Copy(entry.Password); err != nil {
		return errno.ErrClipboardFail.WithCause(err)
	}
	clipboard.ClearAfter(c, clearDelay)
	fmt.Println("Copied, will clear in " + strconv.Itoa(int(clearDelay.Seconds())) + "s")
	return nil
}

func openVault(vaultPath string) (*vault.Vault, string, error) {
	password, err := readMasterPassword("Enter master password: ")
	if err != nil {
		return nil, "", err
	}
	v := vault.New(store.NewFileStore(vaultPath))
	if err := v.Open(password); err != nil {
		return nil, "", err
	}
	return v, password, nil
}

func readMasterPassword(prompt string) (string, error) {
	if password := os.Getenv("PASSMGR_MASTER_PASSWORD"); password != "" {
		return password, nil
	}
	return readLine(prompt)
}

func readLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", errno.ErrInvalidInput.WithCause(err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func usage() {
	fmt.Println(`passmgr - local encrypted password manager

Usage:
  passmgr [--vault-path PATH] init
  passmgr [--vault-path PATH] add <name> -u <username> [-p <password>] [--url <url>] [--notes <notes>]
  passmgr [--vault-path PATH] list
  passmgr [--vault-path PATH] get <name> [--show-password]
  passmgr [--vault-path PATH] rm <name> [--yes]
  passmgr gen [--length N] [--no-symbols] [--copy]
  passmgr [--vault-path PATH] cp <name>

Environment:
  PASSMGR_MASTER_PASSWORD          master password for non-interactive use
  PASSMGR_MASTER_PASSWORD_CONFIRM  init confirmation for non-interactive use`)
}
