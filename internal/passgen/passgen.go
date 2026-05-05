package passgen

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars  = "0123456789"
	symbolChars = "!@#$%^&*"
)

type Options struct {
	Length  int
	Lower   bool
	Upper   bool
	Digits  bool
	Symbols bool
}

func DefaultOptions() Options {
	return Options{Length: 16, Lower: true, Upper: true, Digits: true, Symbols: true}
}

func Generate(opts Options) (string, error) {
	if opts.Length <= 0 {
		return "", errors.New("length must be positive")
	}

	charset := ""
	if opts.Lower {
		charset += lowerChars
	}
	if opts.Upper {
		charset += upperChars
	}
	if opts.Digits {
		charset += digitChars
	}
	if opts.Symbols {
		charset += symbolChars
	}
	if charset == "" {
		return "", errors.New("at least one character set must be enabled")
	}

	out := make([]byte, opts.Length)
	max := big.NewInt(int64(len(charset)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = charset[n.Int64()]
	}
	return string(out), nil
}
