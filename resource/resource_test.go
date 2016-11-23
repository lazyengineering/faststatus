// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

// Generate is used in testing to generate random valid Resource values
func (r Resource) Generate(rand *rand.Rand, size int) reflect.Value {
	rr := Resource{}

	rr.Id = uint64(rand.Int())
	buf := make([]byte, rand.Intn(100))
	rand.Read(buf)
	rr.FriendlyName = string(buf)
	rr.Status = Status(rand.Int() % int(Occupied))
	rr.Since = time.Date(
		2016+rand.Intn(10),
		time.Month(rand.Intn(11)+1),
		rand.Intn(27)+1,
		rand.Intn(24),
		rand.Intn(60),
		rand.Intn(60),
		0,
		time.UTC,
	)

	return reflect.ValueOf(rr)
}

// Expects [0xId] [Status] [Since] [FriendlyName]
func TestResourceString(t *testing.T) {
	type stringTest struct {
		Expected string
		Resource Resource
	}
	tests := []stringTest{
		stringTest{ // Zero Value
			Expected: "0000000000000000 free 0001-01-01T00:00:00Z ",
		},
		stringTest{ // Valid Busy
			Expected: "0000000000000001 busy 2016-05-12T15:09:00-07:00 First One",
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
			Expected: "000000000000000f free 2016-05-12T15:39:00-07:00 Second One",
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
			Expected: "00000000000000af occupied 2016-05-12T15:40:00-07:00 Third One",
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
			Expected: "",
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

func TestResourceMarshalText(t *testing.T) {
	tests := []struct {
		name      string
		resource  Resource
		wantBytes []byte
		wantError bool
	}{
		{"Zero Value",
			Resource{},
			[]byte("0000000000000000 free 0001-01-01T00:00:00Z "),
			false,
		},
		{"Valid Busy",
			Resource{
				Id:           1,
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
			},
			[]byte("0000000000000001 busy 2016-05-12T15:09:00-07:00 First One"),
			false,
		},
		{"Valid Free",
			Resource{
				Id:           15,
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:39:00-07:00")
					return tt
				}(),
			},
			[]byte("000000000000000f free 2016-05-12T15:39:00-07:00 Second One"),
			false,
		},
		{"Valid Occupied",
			Resource{
				Id:           175,
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:40:00-07:00")
					return tt
				}(),
			},
			[]byte("00000000000000af occupied 2016-05-12T15:40:00-07:00 Third One"),
			false,
		},
		{"Out of Range",
			Resource{
				Id:           3503,
				FriendlyName: "Another One",
				Status:       Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:43:00-07:00")
					return tt
				}(),
			},
			[]byte(""),
			true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			txt, err := tc.resource.MarshalText()
			if (err != nil) != tc.wantError {
				t.Fatalf("%+v.MarshalText() = []byte, %+v, expected error? %+v", tc.resource, err, tc.wantError)
			}
			if !bytes.Equal(txt, tc.wantBytes) {
				t.Fatalf("%+v.MarshalText() = %q, error, expected %q", tc.resource, txt, tc.wantBytes)
			}
		})
	}
}

func TestResourceUnmarshalText(t *testing.T) {
	testCases := []struct {
		name         string
		txt          []byte
		wantError    bool
		wantResource Resource
	}{
		{"zero value",
			[]byte{},
			true,
			Resource{},
		},
		{"valid busy",
			[]byte("0000000000000001 busy 2016-05-12T16:25:00-07:00 First One"),
			false,
			Resource{
				Id:           1,
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
		},
		{"valid free",
			[]byte("000000000000000f free 2016-05-12T16:27:00-07:00 Second One"),
			false,
			Resource{
				Id:           0x0f,
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
			},
		},
		{"valid occupied",
			[]byte("00000000000000af occupied 2016-05-12T16:28:00-07:00 Third One"),
			false,
			Resource{
				Id:           0xaf,
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
			},
		},
		{"invalid id",
			[]byte("daf occupied 2016-05-12T16:30:00-07:00 Another One"),
			true,
			Resource{},
		},
		{"invalid status",
			[]byte("0000000000000daf 4 2016-05-12T16:30:00-07:00 Another One"),
			true,
			Resource{},
		},
		{"invalid since",
			[]byte("0000000000000daf busy 16-05-12T16:30:00-07:00 Another One"),
			true,
			Resource{},
		},
		{"missing friendly name",
			[]byte("0000000000000daf busy 2016-05-12T16:30:00-07:00"),
			false,
			Resource{
				Id:     0xdaf,
				Status: Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
			},
		},
		{"missing timestamp",
			[]byte("0000000000000daf busy"),
			true,
			Resource{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var got Resource
			err := (&got).UnmarshalText(tc.txt)
			if (err != nil) != tc.wantError {
				t.Fatalf("resource.UnmarshalText(%q) = %+v, expected error? %+v", tc.txt, err, tc.wantError)
			}
			if !reflect.DeepEqual(got, tc.wantResource) {
				t.Fatalf("%+v.UnmarshalText(%q) = error, expected %+v", got, tc.txt, tc.wantResource)
			}
		})
	}
}

func TestResourceMarshalUnmarshalText(t *testing.T) {
	f := func(r Resource) bool {
		b, err := r.MarshalText()
		if err != nil {
			return false
		}
		got := new(Resource)
		err = got.UnmarshalText(b)
		if err != nil {
			return false
		}
		return reflect.DeepEqual(*got, r)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
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
				[]byte(`{"id":"0","friendlyName":"","status":"free","since":"0001-01-01T00:00:00Z"}`),
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
				[]byte(`{"id":"1","friendlyName":"First One","status":"busy","since":"2016-05-12T16:25:00-07:00"}`),
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
				[]byte(`{"id":"F","friendlyName":"Second One","status":"free","since":"2016-05-12T16:27:00-07:00"}`),
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
				[]byte(`{"id":"AF","friendlyName":"Third One","status":"occupied","since":"2016-05-12T16:28:00-07:00"}`),
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
				"status":"1",
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
				"status":"0",
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
				"status":"2",
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
				"status":"3",
				"since":"2016-05-12T16:30:00-07:00"
			}`),
			Expected: testResponse{
				Resource{},
				func(e error) bool { return e != nil },
			},
		},
		jsonTest{ // Status Overflow
			Input: []byte(`{
				"id":"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
				"friendlyName":"Third One",
				"status":"2",
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
