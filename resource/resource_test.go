// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"encoding/json"
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

func TestResourceMarshalJSON(t *testing.T) {
	type jsonTest struct {
		Expected string
		Resource Resource
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Expected: "{\"id\":\"0\",\"friendlyName\":\"\",\"status\":0,\"since\":\"0001-01-01T00:00:00Z\"}",
		},
		jsonTest{ // Valid Busy
			Expected: "{\"id\":\"1\",\"friendlyName\":\"First One\",\"status\":1,\"since\":\"2016-05-12T16:25:00-07:00\"}",
			Resource: Resource{
				Id:           1,
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
		},
		jsonTest{ // Valid Free
			Expected: "{\"id\":\"F\",\"friendlyName\":\"Second One\",\"status\":0,\"since\":\"2016-05-12T16:27:00-07:00\"}",
			Resource: Resource{
				Id:           15,
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
			},
		},
		jsonTest{ // Valid Occupied
			Expected: "{\"id\":\"AF\",\"friendlyName\":\"Third One\",\"status\":2,\"since\":\"2016-05-12T16:28:00-07:00\"}",
			Resource: Resource{
				Id:           175,
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
			},
		},
		jsonTest{ // Out of Range
			Expected: "{\"id\":\"DAF\",\"friendlyName\":\"Another One\",\"status\":0,\"since\":\"2016-05-12T16:30:00-07:00\"}",
			Resource: Resource{
				Id:           3503,
				FriendlyName: "Another One",
				Status:       Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
			},
		},
	}
	for _, st := range tests {
		if actual, err := json.Marshal(st.Resource); err != nil {
			t.Error(err)
		} else if string(actual) != st.Expected {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", string(actual))
		}
	}
}

func (a Resource) Eq(b Resource) bool {
	return a.Id == b.Id && a.FriendlyName == b.FriendlyName && a.Status == b.Status && a.Since.Equal(b.Since)
}

func TestResourceUnmarshalJSON(t *testing.T) {
	type jsonTest struct {
		Raw      []byte
		Expected Resource
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Raw:      []byte(`{}`),
			Expected: Resource{},
		},
		jsonTest{ // Valid Busy
			Raw: []byte(`{
				"id":"1",
				"friendlyName":"First One",
				"status":1,
				"since":"2016-05-12T16:25:00-07:00"
			}`),
			Expected: Resource{
				Id:           1,
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
		},
		jsonTest{ // Valid Free
			Raw: []byte(`{
				"id":"F",
				"friendlyName":"Second One",
				"status":0,
				"since":"2016-05-12T16:27:00-07:00"
			}`),
			Expected: Resource{
				Id:           15,
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
			},
		},
		jsonTest{ // Valid Occupied
			Raw: []byte(`{
				"id":"AF",
				"friendlyName":"Third One",
				"status":2,
				"since":"2016-05-12T16:28:00-07:00"
			}`),
			Expected: Resource{
				Id:           175,
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
			},
		},
		jsonTest{ // Out of Range
			Raw: []byte(`{
				"id":"DAF",
				"friendlyName":"Another One",
				"status":3,
				"since":"2016-05-12T16:30:00-07:00"
			}`),
			Expected: Resource{
				Id:           3503,
				FriendlyName: "Another One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
			},
		},
	}
	for _, st := range tests {
		actual := new(Resource)
		if err := json.Unmarshal(st.Raw, actual); err != nil {
			t.Error(err)
		} else if !actual.Eq(st.Expected) {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", actual)
		}
	}
}
