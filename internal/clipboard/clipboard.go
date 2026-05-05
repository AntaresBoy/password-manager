package clipboard

import "time"

type Clipboard interface {
	Copy(text string) error
	Clear() error
}

func ClearAfter(c Clipboard, delay time.Duration) <-chan error {
	errs := make(chan error, 1)

	go func() {
		defer close(errs)
		time.Sleep(delay)
		errs <- c.Clear()
	}()

	return errs
}
