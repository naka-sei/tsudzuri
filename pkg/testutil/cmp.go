package testutil

import (
	"errors"
	"reflect"
	"testing"
)

// EqualErr compares two errors and returns a human-readable string describing the differences.
func EqualErr(t *testing.T, want, got error) {
	t.Helper()

	if want == nil && got == nil {
		return
	}

	if want == nil || got == nil {
		t.Fatalf("error mismatch: want=%v, got=%v", want, got)
	}

	if errors.Is(got, want) {
		return
	}

	target := reflect.New(reflect.TypeOf(want)).Interface()
	if errors.As(got, &target) {
		if reflect.TypeOf(target) == reflect.TypeOf(want) {
			return
		}
	}

	t.Fatalf("error mismatch: want=%v, got=%v", want, got)
}
