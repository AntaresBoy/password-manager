package errno

import (
	"errors"
	"testing"
)

func TestErrorAccessors(t *testing.T) {
	err := NewError(20001, "vault not found", 2)

	if got := err.Code(); got != 20001 {
		t.Fatalf("Code() = %d, want %d", got, 20001)
	}
	if got := err.Error(); got != "vault not found" {
		t.Fatalf("Error() = %q, want %q", got, "vault not found")
	}
	if got := err.ExitCode(); got != 2 {
		t.Fatalf("ExitCode() = %d, want %d", got, 2)
	}
}

func TestErrorWithCause(t *testing.T) {
	cause := errors.New("file not found")
	err := NewError(20001, "vault not found", 2)

	if got := err.Unwrap(); got != nil {
		t.Fatalf("Unwrap() before WithCause = %v, want nil", got)
	}
	if got := err.WithCause(cause); got != err {
		t.Fatalf("WithCause() returned different error instance")
	}
	if got := err.Unwrap(); got != cause {
		t.Fatalf("Unwrap() after WithCause = %v, want %v", got, cause)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		code     int
		message  string
		exitCode int
	}{
		{name: "OK", err: OK, code: 0, message: "success", exitCode: 0},
		{name: "ErrInternal", err: ErrInternal, code: 10001, message: "internal error", exitCode: 10},
		{name: "ErrVaultNotFound", err: ErrVaultNotFound, code: 20001, message: "vault not found", exitCode: 2},
		{name: "ErrVaultExists", err: ErrVaultExists, code: 20002, message: "vault already exists", exitCode: 5},
		{name: "ErrVaultCorrupted", err: ErrVaultCorrupted, code: 20003, message: "vault file corrupted", exitCode: 2},
		{name: "ErrWrongPassword", err: ErrWrongPassword, code: 20004, message: "wrong master password", exitCode: 3},
		{name: "ErrEntryNotFound", err: ErrEntryNotFound, code: 20101, message: "entry not found", exitCode: 4},
		{name: "ErrEntryExists", err: ErrEntryExists, code: 20102, message: "entry already exists", exitCode: 5},
		{name: "ErrInvalidInput", err: ErrInvalidInput, code: 20201, message: "invalid input", exitCode: 5},
		{name: "ErrPasswordMismatch", err: ErrPasswordMismatch, code: 20202, message: "passwords do not match", exitCode: 5},
		{name: "ErrClipboardFail", err: ErrClipboardFail, code: 20301, message: "clipboard unavailable", exitCode: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Code(); got != tt.code {
				t.Fatalf("Code() = %d, want %d", got, tt.code)
			}
			if got := tt.err.Error(); got != tt.message {
				t.Fatalf("Error() = %q, want %q", got, tt.message)
			}
			if got := tt.err.ExitCode(); got != tt.exitCode {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.exitCode)
			}
		})
	}
}
