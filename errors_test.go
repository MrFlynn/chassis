package chassis

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestError_Error(t *testing.T) {
	e := Error{
		Display:  errors.New("external"),
		Internal: errors.New("internal"),
	}

	if diff := cmp.Diff(e.Internal.Error(), e.Error()); diff != "" {
		t.Errorf("Error string mismatch (-want +got):\n%s", diff)
	}
}

func TestError_Presentable(t *testing.T) {
	testCases := []struct {
		name     string
		err      Error
		expected string
	}{
		{
			name:     "simple",
			err:      Error{Display: errors.New("hello world")},
			expected: "Hello world.",
		},
		{
			name: "joined",
			err: Error{Display: errors.Join(
				errors.New("hello world"),
				errors.New("lorem ipsum"),
			)},
			expected: "Hello world.\nLorem ipsum.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if diff := cmp.Diff(tc.expected, tc.err.Presentable()); diff != "" {
				t.Errorf("Mismatch in presentable format (-want +got):\n%s", diff)
			}
		})
	}
}
