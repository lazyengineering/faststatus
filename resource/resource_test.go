// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// Expects [0xId] [Status] [Since] [FriendlyName]
func TestResourceString(t *testing.T) {
	type stringTest struct {
		Expected string
		Resource Resource
	}
	tests := []stringTest{
		stringTest{ // Zero Value
			Expected: "0000000000000000 0 0001-01-01T00:00:00Z ",
		},
		stringTest{ // Valid Busy
			Expected: "0000000000000001 1 2016-05-12T15:09:00-07:00 First One",
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
			Expected: "000000000000000F 0 2016-05-12T15:39:00-07:00 Second One",
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
			Expected: "00000000000000AF 2 2016-05-12T15:40:00-07:00 Third One",
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
			Expected: "0000000000000DAF 0 2016-05-12T15:43:00-07:00 Another One",
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
	type testResponse struct {
		Value []byte
		ErrOK func(error) bool
	}
	type jsonTest struct {
		Input    Resource
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Expected: testResponse{
				[]byte(`{"id":"0","friendlyName":"","status":0,"since":"0001-01-01T00:00:00Z"}`),
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Busy
			Input: Resource{
				Id:           1,
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
			Expected: testResponse{
				[]byte(`{"id":"1","friendlyName":"First One","status":1,"since":"2016-05-12T16:25:00-07:00"}`),
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Free
			Input: Resource{
				Id:           15,
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
			},
			Expected: testResponse{
				[]byte(`{"id":"F","friendlyName":"Second One","status":0,"since":"2016-05-12T16:27:00-07:00"}`),
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Occupied
			Input: Resource{
				Id:           175,
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
			},
			Expected: testResponse{
				[]byte(`{"id":"AF","friendlyName":"Third One","status":2,"since":"2016-05-12T16:28:00-07:00"}`),
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Out of Range
			Input: Resource{
				Id:           3503,
				FriendlyName: "Another One",
				Status:       Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
			},
			Expected: testResponse{
				nil,
				func(e error) bool { return !IsOutOfRange(e) },
			},
		},
	}
	for _, st := range tests {
		actual, err := json.Marshal(st.Input)
		if !st.Expected.ErrOK(err) {
			t.Errorf("Resource.MarshalJSON(%v) = '...', %v; expected: %#v", st.Input, err, st.Expected.ErrOK)
		}
		if !reflect.DeepEqual(actual, st.Expected.Value) {
			t.Errorf("Resource.MarshalJSON(%v) = %v, error; expected: %v", st.Input, actual, st.Expected.Value)
		}
	}
}

func TestResourceUnmarshalJSON(t *testing.T) {
	type testResponse struct {
		Value Resource
		ErrOK func(error) bool
	}
	type jsonTest struct {
		Input    []byte
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Input: []byte(`{}`),
			Expected: testResponse{
				Resource{},
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Busy
			Input: []byte(`{
				"id":"1",
				"friendlyName":"First One",
				"status":1,
				"since":"2016-05-12T16:25:00-07:00"
			}`),
			Expected: testResponse{
				Resource{
					Id:           1,
					FriendlyName: "First One",
					Status:       Busy,
					Since: func() time.Time {
						tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
						return tt
					}(),
				},
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Free
			Input: []byte(`{
				"id":"F",
				"friendlyName":"Second One",
				"status":0,
				"since":"2016-05-12T16:27:00-07:00"
			}`),
			Expected: testResponse{
				Resource{
					Id:           15,
					FriendlyName: "Second One",
					Status:       Free,
					Since: func() time.Time {
						tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
						return tt
					}(),
				},
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Occupied
			Input: []byte(`{
				"id":"AF",
				"friendlyName":"Third One",
				"status":2,
				"since":"2016-05-12T16:28:00-07:00"
			}`),
			Expected: testResponse{
				Resource{
					Id:           175,
					FriendlyName: "Third One",
					Status:       Occupied,
					Since: func() time.Time {
						tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
						return tt
					}(),
				},
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Out of Range
			Input: []byte(`{
				"id":"DAF",
				"friendlyName":"Another One",
				"status":3,
				"since":"2016-05-12T16:30:00-07:00"
			}`),
			Expected: testResponse{
				Resource{},
				func(e error) bool { return IsOutOfRange(e) },
			},
		},
		jsonTest{ // Status Overflow
			Input: []byte(`{
				"id":"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
				"friendlyName":"Third One",
				"status":2,
				"since":"2016-05-12T16:28:00-07:00"
			}`),
			Expected: testResponse{
				Resource{},
				func(e error) bool { return e != nil },
			},
		},
	}
	for _, st := range tests {
		var actual Resource
		err := json.Unmarshal(st.Input, &actual)
		if !st.Expected.ErrOK(err) {
			t.Errorf("Resource.UnmarshalJSON(%v) = %v; expected: %#v", st.Input, err, st.Expected.ErrOK)
		}
		if !reflect.DeepEqual(actual, st.Expected.Value) {
			t.Errorf("Resource.UnmarshalJSON(%v), Resource: %v; expected: %v", st.Input, actual, st.Expected.Value)
		}
	}
}

func TestResourceMarshalUnmarshalJSON(t *testing.T) {
	type testResponse struct {
		Value Resource
		ErrOK func(error) bool
	}
	type jsonTest struct {
		Input    Resource
		Expected testResponse
	}
	tests := []jsonTest{
		jsonTest{ // Zero Value
			Input: Resource{},
			Expected: testResponse{
				Resource{},
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Busy
			Input: Resource{
				Id:           1,
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
			Expected: testResponse{
				Resource{
					Id:           1,
					FriendlyName: "First One",
					Status:       Busy,
					Since: func() time.Time {
						tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
						return tt
					}(),
				},
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Free
			Input: Resource{
				Id:           15,
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
			},
			Expected: testResponse{
				Resource{
					Id:           15,
					FriendlyName: "Second One",
					Status:       Free,
					Since: func() time.Time {
						tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
						return tt
					}(),
				},
				func(e error) bool { return e == nil },
			},
		},
		jsonTest{ // Valid Occupied
			Input: Resource{
				Id:           175,
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
			},
			Expected: testResponse{
				Resource{
					Id:           175,
					FriendlyName: "Third One",
					Status:       Occupied,
					Since: func() time.Time {
						tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
						return tt
					}(),
				},
				// func(e error) bool { return e == nil },
				func(e error) bool { return !IsOutOfRange(e) },
			},
		},
		jsonTest{ // Out of Range
			Input: Resource{
				Id:           3503,
				FriendlyName: "Another One",
				Status:       Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
			},
			Expected: testResponse{
				Resource{},
				func(e error) bool { return !IsOutOfRange(e) },
			},
		},
	}
	for _, st := range tests {
		actual, err := func(r Resource) (Resource, error) {
			ac := new(Resource)
			tmp, erx := json.Marshal(r)
			if erx != nil {
				return *ac, erx
			}
			erx = json.Unmarshal(tmp, ac)
			return *ac, erx
		}(st.Input)
		if !st.Expected.ErrOK(err) {
			t.Errorf("Resource.UnmarshalJSON(%v) = %v; expected: %#v", st.Input, err, st.Expected.ErrOK)
		}
		if !reflect.DeepEqual(actual, st.Expected.Value) {
			t.Errorf("Resource.UnmarshalJSON(%v), Resource: %#v; expected: %#v", st.Input, actual, st.Expected.Value)
		}
	}
}
