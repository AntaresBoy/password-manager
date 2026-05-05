package store

type Store interface {
	Read() ([]byte, error)
	Write([]byte) error
	Exists() bool
	Path() string
}
