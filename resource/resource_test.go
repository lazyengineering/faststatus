// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"testing"
	"time"
)

func TestResourceString(t *testing.T) {
	type stringTest struct {
		Expected string
		Resource Resource
	}
	tests := []stringTest{
		stringTest{ // Zero Value
			Expected: "0001-01-01T00:00:00Z 0 0000000000000000 \n",
		},
		stringTest{ // Valid Busy
			Expected: "2016-05-12T15:09:00-07:00 1 0000000000000001 First One\n",
			Resource: Resource{
				Id:           1,
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
			},
		},
		stringTest{ // Valid Free
			Expected: "2016-05-12T15:39:00-07:00 0 000000000000000F Second One\n",
			Resource: Resource{
				Id:           15,
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:39:00-07:00")
					return tt
				}(),
			},
		},
		stringTest{ // Valid Occupied
			Expected: "2016-05-12T15:40:00-07:00 2 00000000000000AF Third One\n",
			Resource: Resource{
				Id:           175,
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:40:00-07:00")
					return tt
				}(),
			},
		},
		stringTest{ // Out of Range
			Expected: "2016-05-12T15:43:00-07:00 0 0000000000000DAF Another One\n",
			Resource: Resource{
				Id:           3503,
				FriendlyName: "Another One",
				Status:       Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:43:00-07:00")
					return tt
				}(),
			},
		},
	}
	for _, st := range tests {
		if actual := st.Resource.String(); actual != st.Expected {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", actual)
		}
	}
}
