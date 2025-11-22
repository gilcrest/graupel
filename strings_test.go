// strings.go and strings_test.go are adapted from the Go-GitHub project
// Copyright 2013 The go-github AUTHORS. All rights reserved.

package graupel

import (
	"fmt"
	"testing"
)

func TestStringify(t *testing.T) {
	t.Parallel()
	var nilPointer *string

	tests := []struct {
		in  any
		out string
	}{
		// basic types
		{"foo", `"foo"`},
		{123, `123`},
		{1.5, `1.5`},
		{false, `false`},
		{
			[]string{"a", "b"},
			`["a" "b"]`,
		},
		{
			struct {
				A []string
			}{nil},
			// nil slice is skipped
			`{}`,
		},
		{
			struct {
				A string
			}{"foo"},
			// structs not of a named type get no prefix
			`{A:"foo"}`,
		},

		// pointers
		{nilPointer, `<nil>`},
		{Ptr("foo"), `"foo"`},
		{Ptr(123), `123`},
		{Ptr(false), `false`},
		{
			//nolint:sliceofpointers
			[]*string{Ptr("a"), Ptr("b")},
			`["a" "b"]`,
		},
	}

	for i, tt := range tests {
		s := Stringify(tt.in)
		if s != tt.out {
			t.Errorf("%v. Stringify(%q) => %q, want %q", i, tt.in, s, tt.out)
		}
	}
}

// Directly test the String() methods on various GitHub types. We don't do an
// exhaustive test of all the various field types, since TestStringify() above
// takes care of that. Rather, we just make sure that Stringify() is being
// used to build the strings, which we do by verifying that pointers are
// stringified as their underlying value.
func TestString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in  any
		out string
	}{
		{ModelConfig{Orchestration: Ptr("o")}, `graupel.ModelConfig{Orchestration:"o"}`},
	}

	for i, tt := range tests {
		s := tt.in.(fmt.Stringer).String()
		if s != tt.out {
			t.Errorf("%v. String() => %q, want %q", i, tt.in, tt.out)
		}
	}
}
