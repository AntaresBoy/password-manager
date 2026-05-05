package clipboard

import "github.com/atotto/clipboard"

var writeAll = clipboard.WriteAll

type SystemClipboard struct{}

func NewSystemClipboard() *SystemClipboard {
	return &SystemClipboard{}
}

func (s *SystemClipboard) Copy(text string) error {
	return writeAll(text)
}

func (s *SystemClipboard) Clear() error {
	return writeAll("")
}
