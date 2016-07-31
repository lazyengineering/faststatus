// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"encoding/json" // because we explicitly want to work with the standard library json package
	"reflect"
	"testing"
)

func TestStatusString(t *testing.T) {
	type stringTest struct {
		Expected string
		Status   Status
	}
	tests := []stringTest{
		stringTest{ // Zero Value
			Expected: "0",
		},
		stringTest{ // Free
			Expected: "0",
			Status:   Free,
		},
		stringTest{ // Busy
			Expected: "1",
			Status:   Busy,
		},
		stringTest{ // Occupied
			Expected: "2",
			Status:   Occupied,
		},
		stringTest{ // Out of Range
			Expected: "0",
			Status:   Occupied + 1,
		},
	}
	for _, st := range tests {
		if actual := st.Status.String(); actual != st.Expected {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", actual)
		}
	}
}

func TestStatusPretty(t *testing.T) {
	type stringTest struct {
		Expected string
		Status   Status
	}
	tests := []stringTest{
		stringTest{ // Zero Value
			Expected: "Free",
		},
		stringTest{ // Free
			Expected: "Free",
			Status:   Free,
		},
		stringTest{ // Busy
			Expected: "Busy",
			Status:   Busy,
		},
		stringTest{ // Occupied
			Expected: "Occupied",
			Status:   Occupied,
		},
		stringTest{ // Out of Range
			Expected: "Free",
			Status:   Occupied + 1,
		},
	}
	for _, st := range tests {
		if actual := st.Status.Pretty(); actual != st.Expected {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", actual)
		}
	}
}

func TestStatusMarshalJSON(t *testing.T) {
	type testResponse struct {
		Value []byte
		ErrOK func(error) bool
	}
	type jsonTest struct {
		Input    Status
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Expected: testResponse{
				[]byte("0"),
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Free
			Input: Free,
			Expected: testResponse{
				[]byte("0"),
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Busy
			Input: Busy,
			Expected: testResponse{
				[]byte("1"),
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Occupied
			Input: Occupied,
			Expected: testResponse{
				[]byte("2"),
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Out of Range
			Input: Occupied + 1,
			Expected: testResponse{
				[]byte(nil),
				func(e error) bool { return !IsOutOfRange(e) },
			},
		},
	}
	for _, st := range tests {
		actual, err := json.Marshal(st.Input)
		if !st.Expected.ErrOK(err) {
			t.Errorf("Status.MarshalJSON(%v) = '...', %v; expected: %#v", st.Input, err, st.Expected.ErrOK)
		}
		if !reflect.DeepEqual(actual, st.Expected.Value) {
			t.Errorf("Status.MarshalJSON(%v) = %#v, error; expected: %#v", st.Input, actual, st.Expected.Value)
		}
	}
}

func TestStatusUnmarshalJSON(t *testing.T) {
	type testResponse struct {
		Value Status
		ErrOK func(error) bool
	}
	type jsonTest struct {
		Input    []byte
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Free
			Input: []byte("0"),
			Expected: testResponse{
				Free,
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Busy
			Input: []byte("1"),
			Expected: testResponse{
				Busy,
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Occupied
			Input: []byte("2"),
			Expected: testResponse{
				Occupied,
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Out of Range
			Input: []byte("3"),
			Expected: testResponse{
				0,
				func(e error) bool { return IsOutOfRange(e) },
			},
		},
		jsonTest{ // Non-number
			Input: []byte(`"π"`),
			Expected: testResponse{
				0,
				func(e error) bool { return e != nil },
			},
		},
	}
	for _, st := range tests {
		var actual Status
		err := json.Unmarshal(st.Input, &actual)
		if !st.Expected.ErrOK(err) {
			t.Errorf("Status.UnmarshalJSON(%v) = %v; expected: %#v", st.Input, err, st.Expected.ErrOK)
		}
		if !reflect.DeepEqual(actual, st.Expected.Value) {
			t.Errorf("Status.UnmarshalJSON(%v), Status: %v; expected: %v", st.Input, actual, st.Expected.Value)
		}
	}
}

func TestStatusMarshalUnmarshalJSON(t *testing.T) {
	// Expects identical status to Input
	type testResponse struct {
		Value Status
		ErrOK func(e error) bool
	}
	type jsonTest struct {
		Input    Status
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Expected: testResponse{
				Free,
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Free
			Input: Free,
			Expected: testResponse{
				Free,
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Busy
			Input: Busy,
			Expected: testResponse{
				Busy,
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // OccupiedFree
			Input: Occupied,
			Expected: testResponse{
				Occupied,
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Out of Range
			Input: Occupied + 1,
			Expected: testResponse{
				Free,
				func(e error) bool { return !IsOutOfRange(e) },
			},
		},
	}
	for _, st := range tests {
		actual, err := func(s Status) (Status, error) {
			ac := new(Status)
			tmp, erx := json.Marshal(s)
			if erx != nil {
				return *ac, erx
			}
			erx = json.Unmarshal(tmp, ac)
			return *ac, erx
		}(st.Input)
		if !st.Expected.ErrOK(err) {
			t.Errorf("Status.MarshalJSON(%v) = '...', %v; expected: %#v", st.Input, err, st.Expected.ErrOK)
		}
		if !reflect.DeepEqual(actual, st.Expected.Value) {
			t.Errorf("Status.MarshalJSON(%v) = %v, error; expected: %v", st.Input, actual, st.Expected.Value)
		}
	}
}
