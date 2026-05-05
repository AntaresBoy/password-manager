package clipboard

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type fakeClipboard struct {
	copiedText string
	copyCalls  int
	clearCalls int
	clearErr   error
}

func (f *fakeClipboard) Copy(text string) error {
	f.copyCalls++
	f.copiedText = text
	return nil
}

func (f *fakeClipboard) Clear() error {
	f.clearCalls++
	return f.clearErr
}

func TestClipboardInterfaceBehaviorWithFake(t *testing.T) {
	var c Clipboard = &fakeClipboard{}

	if err := c.Copy("secret"); err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}

	f := c.(*fakeClipboard)
	if f.copyCalls != 1 {
		t.Fatalf("Copy calls = %d, want 1", f.copyCalls)
	}
	if f.copiedText != "secret" {
		t.Fatalf("copied text = %q, want %q", f.copiedText, "secret")
	}
	if f.clearCalls != 1 {
		t.Fatalf("Clear calls = %d, want 1", f.clearCalls)
	}
}

func TestSystemClipboardDelegatesCopyAndClear(t *testing.T) {
	originalWriteAll := writeAll
	t.Cleanup(func() { writeAll = originalWriteAll })

	var writes []string
	writeAll = func(text string) error {
		writes = append(writes, text)
		return nil
	}

	c := NewSystemClipboard()
	if err := c.Copy("secret"); err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}

	want := []string{"secret", ""}
	if !reflect.DeepEqual(writes, want) {
		t.Fatalf("writes = %#v, want %#v", writes, want)
	}
}

func TestClearAfterClearsAfterDelay(t *testing.T) {
	f := &fakeClipboard{}

	errs := ClearAfter(f, time.Millisecond)

	select {
	case err, ok := <-errs:
		if !ok {
			t.Fatal("error channel closed before reporting clear result")
		}
		if err != nil {
			t.Fatalf("ClearAfter returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ClearAfter")
	}

	if f.clearCalls != 1 {
		t.Fatalf("Clear calls = %d, want 1", f.clearCalls)
	}
	if _, ok := <-errs; ok {
		t.Fatal("error channel still open after clear result")
	}
}

func TestClearAfterPropagatesClearError(t *testing.T) {
	wantErr := errors.New("clipboard unavailable")
	f := &fakeClipboard{clearErr: wantErr}

	errs := ClearAfter(f, time.Millisecond)

	select {
	case err := <-errs:
		if !errors.Is(err, wantErr) {
			t.Fatalf("ClearAfter error = %v, want %v", err, wantErr)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ClearAfter")
	}
}
