// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"encoding/json" // because we explicitly want to work with the standard library json package
	"strings"
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
		Err   error
	}
	type jsonTest struct {
		Input    Status
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Expected: testResponse{[]byte("0"), nil},
		},
		jsonTest{ // Free
			Input:    Free,
			Expected: testResponse{[]byte("0"), nil},
		},
		jsonTest{ // Busy
			Input:    Busy,
			Expected: testResponse{[]byte("1"), nil},
		},
		jsonTest{ // Occupied
			Input:    Occupied,
			Expected: testResponse{[]byte("2"), nil},
		},
		jsonTest{ // Out of Range
			Input:    Occupied + 1,
			Expected: testResponse{[]byte(""), ErrOutOfRange},
		},
	}
	for _, st := range tests {
		if actual, err := json.Marshal(st.Input); !rootError(err, st.Expected.Err) || string(actual) != string(st.Expected.Value) {
			t.Errorf("\nexpected:\t%q\t%q\n  actual:\t%q\t%q", string(st.Expected.Value), st.Expected.Err, string(actual), err)
		}
	}
}

// compare two errors to see if they are both nil or e contains root
func rootError(e, root error) bool {
	if e == nil && root == nil {
		return true
	} else if e == nil || root == nil {
		return false
	}
	return strings.Contains(e.Error(), root.Error())
}

func TestStatusUnmarshalJSON(t *testing.T) {
	type testResponse struct {
		Value Status
		Err   error
	}
	type jsonTest struct {
		Input    []byte
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Free
			Input:    []byte("0"),
			Expected: testResponse{Free, nil},
		},
		jsonTest{ // Busy
			Input:    []byte("1"),
			Expected: testResponse{Busy, nil},
		},
		jsonTest{ // Occupied
			Input:    []byte("2"),
			Expected: testResponse{Occupied, nil},
		},
		jsonTest{ // Out of Range
			Input:    []byte("3"),
			Expected: testResponse{0, ErrOutOfRange},
		},
	}
	for _, st := range tests {
		actual := new(Status)
		if err := json.Unmarshal(st.Input, actual); !rootError(err, st.Expected.Err) || *actual != st.Expected.Value {
			t.Errorf("\nexpected:\t%v\t%q\n  actual:\t%v\t%q", uint8(st.Expected.Value), st.Expected.Err, uint8(*actual), err)
		}
	}
}

func TestStatusMarshalUnmarshalJSON(t *testing.T) {
	// Expects identical status to Input
	type testResponse struct {
		Value Status
		Err   error
	}
	type jsonTest struct {
		Input    Status
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Expected: testResponse{Free, nil},
		},
		jsonTest{ // Free
			Input:    Free,
			Expected: testResponse{Free, nil},
		},
		jsonTest{ // Busy
			Input:    Busy,
			Expected: testResponse{Busy, nil},
		},
		jsonTest{ // OccupiedFree
			Input:    Occupied,
			Expected: testResponse{Occupied, nil},
		},
		jsonTest{ // Out of Range
			Input:    Occupied + 1,
			Expected: testResponse{Free, ErrOutOfRange},
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
		if !rootError(err, st.Expected.Err) || actual != st.Expected.Value {
			t.Errorf("\nexpected:\t%v\t%v\n  actual:\t%v\t%v", uint8(st.Expected.Value), st.Expected.Err, uint8(actual), err)
		}
	}
}
