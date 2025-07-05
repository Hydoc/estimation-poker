package assert

import (
	"reflect"
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()

	if got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()

	if !got {
		t.Errorf("got: %v, want: true", got)
	}
}

func False(t *testing.T, got bool) {
	t.Helper()

	if got {
		t.Errorf("got: %v, want: false", got)
	}
}

func DeepEqual[T any](t *testing.T, actual, expected T) {
	t.Helper()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got: %#v; want: %#v", actual, expected)
	}
}

func StringContains(t *testing.T, got, want string) {
	t.Helper()

	if !strings.Contains(got, want) {
		t.Errorf("got: %v should contain want: %v", got, want)
	}
}

func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %v; expected: nil", actual)
	}
}
