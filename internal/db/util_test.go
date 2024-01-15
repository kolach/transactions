package db

import (
	"fmt"
	"testing"
)

func TestStringToInt32Ptr(t *testing.T) {
	tests := []struct {
		input string
		want  *int32
		err   error
	}{
		{"123", int32Ptr(123), nil},
		{"-456", int32Ptr(-456), nil},
		{"0", int32Ptr(0), nil},
		{"", nil, nil},
		{"abc", nil, fmt.Errorf("failed to convert string to int32: expected integer")},
	}

	for _, test := range tests {
		got, err := stringToInt32Ptr(test.input)
		if err != nil && test.err == nil {
			t.Errorf("stringToInt32Ptr(%q) returned unexpected error: %v", test.input, err)
		} else if err == nil && test.err != nil {
			t.Errorf("stringToInt32Ptr(%q) did not return expected error", test.input)
		} else if err != nil && test.err != nil && err.Error() != test.err.Error() {
			t.Errorf("stringToInt32Ptr(%q) returned unexpected error. Got: %v, Want: %v", test.input, err, test.err)
		}

		if got == nil && test.want != nil {
			t.Errorf("stringToInt32Ptr(%q) returned nil, want %v", test.input, *test.want)
		} else if got != nil && test.want == nil {
			t.Errorf("stringToInt32Ptr(%q) returned %v, want nil", test.input, *got)
		} else if got != nil && test.want != nil && *got != *test.want {
			t.Errorf("stringToInt32Ptr(%q) returned %v, want %v", test.input, *got, *test.want)
		}
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
