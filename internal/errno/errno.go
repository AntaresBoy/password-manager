package errno

type Error struct {
	code    int
	message string
	exit    int
	cause   error
}

func NewError(code int, message string, exit int) *Error {
	return &Error{code: code, message: message, exit: exit}
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) ExitCode() int {
	return e.exit
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) WithCause(cause error) *Error {
	e.cause = cause
	return e
}

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
