// strings.go and strings_test.go are adapted from the Go-GitHub project
// Copyright 2013 The go-github AUTHORS. All rights reserved.

package graupel

import (
	"bytes"
	"fmt"
	"reflect"
)

// Stringify attempts to create a reasonable string representation of types in
// the Graupel library. It does things like resolve pointers to their values
// and omits struct fields with nil values.
func Stringify(message any) string {
	var buf bytes.Buffer
	v := reflect.ValueOf(message)
	stringifyValue(&buf, v)
	return buf.String()
}

// stringifyValue was heavily inspired by the goprotobuf library.
func stringifyValue(w *bytes.Buffer, val reflect.Value) {
	if val.Kind() == reflect.Pointer && val.IsNil() {
		w.WriteString("<nil>")
		return
	}

	v := reflect.Indirect(val)

	switch v.Kind() {
	case reflect.String:
		fmt.Fprintf(w, `"%v"`, v)
	case reflect.Slice:
		w.WriteByte('[')
		for i := range v.Len() {
			if i > 0 {
				w.WriteByte(' ')
			}

			stringifyValue(w, v.Index(i))
		}

		w.WriteByte(']')
		return
	case reflect.Struct:
		if v.Type().Name() != "" {
			w.WriteString(v.Type().String())
		}

		w.WriteByte('{')

		var sep bool
		for i := range v.NumField() {
			fv := v.Field(i)
			if fv.Kind() == reflect.Pointer && fv.IsNil() {
				continue
			}
			if fv.Kind() == reflect.Slice && fv.IsNil() {
				continue
			}
			if fv.Kind() == reflect.Map && fv.IsNil() {
				continue
			}

			if sep {
				w.WriteString(", ")
			} else {
				sep = true
			}

			w.WriteString(v.Type().Field(i).Name)
			w.WriteByte(':')
			stringifyValue(w, fv)
		}

		w.WriteByte('}')
	default:
		if v.CanInterface() {
			fmt.Fprint(w, v.Interface())
		}
	}
}
