// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
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
