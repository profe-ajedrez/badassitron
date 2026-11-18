package errors

import (
	"strings"
	"testing"
)

func TestMigrationInvalidErrorStateDoesNotPanic(t *testing.T) {
	for _, code := range []Error{11, 127, 255} {
		if got := code.Value(); got == nil || !strings.Contains(got.Error(), "invalid state code") {
			t.Fatalf("Value(%d) = %v", code, got)
		}
		if got := code.Error(); !strings.Contains(got, "invalid state code") {
			t.Fatalf("Error(%d) = %q", code, got)
		}
	}
	if None.Value() != nil || None.Error() != "" {
		t.Fatal("None is not an error")
	}
	if Overflow.Value() == nil || Overflow.Error() != "overflow" {
		t.Fatal("historical error code changed")
	}
}
