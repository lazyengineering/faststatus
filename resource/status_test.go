// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"encoding/json" // because we explicitly want to work with the standard library json package
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
	type jsonTest struct {
		Expected string
		Status   Status
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Expected: "0",
		},
		jsonTest{ // Free
			Expected: "0",
			Status:   Free,
		},
		jsonTest{ // Busy
			Expected: "1",
			Status:   Busy,
		},
		jsonTest{ // Occupied
			Expected: "2",
			Status:   Occupied,
		},
		jsonTest{ // Out of Range
			Expected: "0",
			Status:   Occupied + 1,
		},
	}
	for _, st := range tests {
		if actual, err := json.Marshal(st.Status); err != nil {
			t.Error(err)
		} else if string(actual) != st.Expected {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", string(actual))
		}
	}
}

func TestStatusUnmarshalJSON(t *testing.T) {
	type jsonTest struct {
		Raw      []byte
		Expected Status
	}
	tests := []jsonTest{
		jsonTest{ // Free
			Raw:      []byte("0"),
			Expected: Free,
		},
		jsonTest{ // Busy
			Raw:      []byte("1"),
			Expected: Busy,
		},
		jsonTest{ // Occupied
			Raw:      []byte("2"),
			Expected: Occupied,
		},
		jsonTest{ // Out of Range
			Raw:      []byte("3"),
			Expected: Free,
		},
	}
	for _, st := range tests {
		actual := new(Status)
		if err := json.Unmarshal(st.Raw, actual); err != nil {
			t.Error(err)
		} else if *actual != st.Expected {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", actual)
		}
	}
}
