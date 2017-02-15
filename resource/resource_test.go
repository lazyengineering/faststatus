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

	rr.ID, _ = NewID()
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

// Expects [ID] [Status] [Since] [FriendlyName]
func TestResourceString(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		resource Resource
	}{
		{"Zero Value",
			"00000000-0000-0000-0000-000000000000 free 0001-01-01T00:00:00Z",
			Resource{},
		},
		{"Valid Busy",
			"01234567-89ab-cdef-0123-456789abcdef busy 2016-05-12T15:09:00-07:00 First One",
			Resource{
				ID:     ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
		},
		{"Valid Free",
			"23456789-abcd-ef01-2345-6789abcdef01 free 2016-05-12T15:39:00-07:00 Second One",
			Resource{
				ID:     ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:39:00-07:00")
					return tt
				}(),
				FriendlyName: "Second One",
			},
		},
		{"Valid Occupied",
			"456789ab-cdef-0123-4567-89abcdef0123 occupied 2016-05-12T15:40:00-07:00 Third One",
			Resource{
				ID:     ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:40:00-07:00")
					return tt
				}(),
				FriendlyName: "Third One",
			},
		},
		{"Out of Range",
			"",
			Resource{
				ID:     ID{0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45},
				Status: Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:43:00-07:00")
					return tt
				}(),
				FriendlyName: "Another One",
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if actual := tc.resource.String(); actual != tc.expected {
				t.Fatalf("%+v.String() = %q, expected %q", tc.resource, actual, tc.expected)
			}
		})
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
			[]byte("00000000-0000-0000-0000-000000000000 free 0001-01-01T00:00:00Z"),
			false,
		},
		{"Valid Busy",
			Resource{
				ID:     ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			[]byte("01234567-89ab-cdef-0123-456789abcdef busy 2016-05-12T15:09:00-07:00 First One"),
			false,
		},
		{"Valid Free",
			Resource{
				ID:     ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:39:00-07:00")
					return tt
				}(),
				FriendlyName: "Second One",
			},
			[]byte("23456789-abcd-ef01-2345-6789abcdef01 free 2016-05-12T15:39:00-07:00 Second One"),
			false,
		},
		{"Valid Occupied",
			Resource{
				ID:     ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:40:00-07:00")
					return tt
				}(),
				FriendlyName: "Third One",
			},
			[]byte("456789ab-cdef-0123-4567-89abcdef0123 occupied 2016-05-12T15:40:00-07:00 Third One"),
			false,
		},
		{"Out of Range",
			Resource{
				ID:     ID{0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45},
				Status: Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:43:00-07:00")
					return tt
				}(),
				FriendlyName: "Another One",
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
			[]byte("01234567-89ab-cdef-0123-456789abcdef busy 2016-05-12T16:25:00-07:00 First One"),
			false,
			Resource{
				ID:           ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
		},
		{"valid busy (numeric status)",
			[]byte("01234567-89ab-cdef-0123-456789abcdef 1 2016-05-12T16:25:00-07:00 First One"),
			false,
			Resource{
				ID:           ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
		},
		{"valid free",
			[]byte("23456789-abcd-ef01-2345-6789abcdef01 free 2016-05-12T16:27:00-07:00 Second One"),
			false,
			Resource{
				ID:     ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
				FriendlyName: "Second One",
			},
		},
		{"valid free (numeric status)",
			[]byte("23456789-abcd-ef01-2345-6789abcdef01 0 2016-05-12T16:27:00-07:00 Second One"),
			false,
			Resource{
				ID:     ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
				FriendlyName: "Second One",
			},
		},
		{"valid occupied",
			[]byte("456789ab-cdef-0123-4567-89abcdef0123 occupied 2016-05-12T16:28:00-07:00 Third One"),
			false,
			Resource{
				ID:     ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
				FriendlyName: "Third One",
			},
		},
		{"valid occupied (numeric status)",
			[]byte("456789ab-cdef-0123-4567-89abcdef0123 2 2016-05-12T16:28:00-07:00 Third One"),
			false,
			Resource{
				ID:     ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
				FriendlyName: "Third One",
			},
		},
		{"invalid id",
			[]byte("0123456--0000-0000-0000-000000000000 occupied 2016-05-12T16:30:00-07:00 Another One"),
			true,
			Resource{},
		},
		{"invalid status",
			[]byte("01234567-89ab-cdef-0123-456789abcdef 4 2016-05-12T16:30:00-07:00 Another One"),
			true,
			Resource{},
		},
		{"invalid since",
			[]byte("01234567-89ab-cdef-0123-456789abcdef busy 16-05-12T16:30:00-07:00 Another One"),
			true,
			Resource{},
		},
		{"missing friendly name",
			[]byte("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa busy 2016-05-12T16:30:00-07:00"),
			false,
			Resource{
				ID:     ID{0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa},
				Status: Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
			},
		},
		{"missing timestamp",
			[]byte("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb busy"),
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
	testCases := []struct {
		name      string
		resource  Resource
		wantValue []byte
		wantError bool
	}{
		{"Zero Value",
			Resource{},
			[]byte(`{"id":"00000000-0000-0000-0000-000000000000","status":"free","since":"0001-01-01T00:00:00Z","friendlyName":""}`),
			false,
		},
		{"Valid Busy",
			Resource{
				ID:     ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			[]byte(`{"id":"01234567-89ab-cdef-0123-456789abcdef","status":"busy","since":"2016-05-12T16:25:00-07:00","friendlyName":"First One"}`),
			false,
		},
		{"Valid Free",
			Resource{
				ID:     ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
				FriendlyName: "Second One",
			},
			[]byte(`{"id":"23456789-abcd-ef01-2345-6789abcdef01","status":"free","since":"2016-05-12T16:27:00-07:00","friendlyName":"Second One"}`),
			false,
		},
		{"Valid Occupied",
			Resource{
				ID:     ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
				FriendlyName: "Third One",
			},
			[]byte(`{"id":"456789ab-cdef-0123-4567-89abcdef0123","status":"occupied","since":"2016-05-12T16:28:00-07:00","friendlyName":"Third One"}`),
			false,
		},
		{"Out of Range",
			Resource{
				ID:     ID{0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45},
				Status: Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
				FriendlyName: "Another One",
			},
			nil,
			true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual, err := json.Marshal(tc.resource)
			if (err != nil) != tc.wantError {
				t.Fatalf("json.Marshal(%+v) = <[]byte>, %+v; expected error? %+v", tc.resource, err, tc.wantError)
			}
			if !reflect.DeepEqual(actual, tc.wantValue) {
				t.Fatalf("json.Marshal(%+v) = %s, <error>, expected %s", tc.resource, actual, tc.wantValue)
			}
		})
	}
}

func TestResourceUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name         string
		input        []byte
		wantResource Resource
		wantError    bool
	}{
		{"Zero Value",
			[]byte(`{}`),
			Resource{},
			false,
		},
		{"Valid Busy",
			[]byte(`{
				"id":"01234567-89ab-cdef-0123-456789abcdef",
				"status":"1",
				"since":"2016-05-12T16:25:00-07:00",
				"friendlyName":"First One"
			}`),
			Resource{
				ID:     ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			false,
		},
		{"Valid Busy text value",
			[]byte(`{
				"id":"01234567-89ab-cdef-0123-456789abcdef",
				"status":"busy",
				"since":"2016-05-12T16:25:00-07:00",
				"friendlyName":"First One"
			}`),
			Resource{
				ID:     ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			false,
		},
		{"Valid Free",
			[]byte(`{
				"friendlyName":"Second One",
				"id":"23456789-abcd-ef01-2345-6789abcdef01",
				"status":"0",
				"since":"2016-05-12T16:27:00-07:00"
			}`),
			Resource{
				ID:     ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
				FriendlyName: "Second One",
			},
			false,
		},
		{"Valid Free text value",
			[]byte(`{
				"friendlyName":"Second One",
				"id":"23456789-abcd-ef01-2345-6789abcdef01",
				"status":"FrEe",
				"since":"2016-05-12T16:27:00-07:00"
			}`),
			Resource{
				ID:     ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
				FriendlyName: "Second One",
			},
			false,
		},
		{"Valid Occupied",
			[]byte(`{
				"since":"2016-05-12T16:28:00-07:00",
				"status":"2",
				"friendlyName":"Third One",
				"id":"456789ab-cdef-0123-4567-89abcdef0123"
			}`),
			Resource{
				ID:     ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
				FriendlyName: "Third One",
			},
			false,
		},
		{"Valid Occupied text value",
			[]byte(`{
				"since":"2016-05-12T16:28:00-07:00",
				"status":"OCCUPIED",
				"friendlyName":"Third One",
				"id":"456789ab-cdef-0123-4567-89abcdef0123"
			}`),
			Resource{
				ID:     ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
				FriendlyName: "Third One",
			},
			false,
		},
		{"Out of Range",
			[]byte(`{
				"id":"6789abcd-ef01-2345-6789-abcdef012345",
				"friendlyName":"Another One",
				"status":"3",
				"since":"2016-05-12T16:30:00-07:00"
			}`),
			Resource{},
			true,
		},
		{"Bad ID",
			[]byte(`{
				"id":"01234567-89ab-cdef-0123-456789abcdef0",
				"friendlyName":"Third One",
				"status":"2",
				"since":"2016-05-12T16:28:00-07:00"
			}`),
			Resource{},
			true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var actual Resource
			err := json.Unmarshal(tc.input, &actual)
			if (err != nil) != tc.wantError {
				t.Fatalf("json.Unmarshal(%s, *<Resource>) = %+v, expected error? %+v", tc.input, err, tc.wantError)
			}
			if !reflect.DeepEqual(actual, tc.wantResource) {
				t.Fatalf("json.Unmarshal(%s, %+v) = <error>, expected %+v", tc.input, actual, tc.wantResource)
			}
		})
	}
}

func TestResourceMarshalUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name         string
		resource     Resource
		wantResource Resource
		wantError    bool
	}{
		{"Zero Value",
			Resource{},
			Resource{},
			false,
		},
		{"Valid Busy",
			Resource{
				ID:           ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
			Resource{
				ID:           ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				FriendlyName: "First One",
				Status:       Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
			false,
		},
		{"Valid Free",
			Resource{
				ID:           ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
			},
			Resource{
				ID:           ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				FriendlyName: "Second One",
				Status:       Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
					return tt
				}(),
			},
			false,
		},
		{"Valid Occupied",
			Resource{
				ID:           ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
			},
			Resource{
				ID:           ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				FriendlyName: "Third One",
				Status:       Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
					return tt
				}(),
			},
			false,
		},
		{"Out of Range",
			Resource{
				ID:           ID{0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45},
				FriendlyName: "Another One",
				Status:       Occupied + 1,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
			},
			Resource{},
			true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual, err := func(r Resource) (Resource, error) {
				ac := new(Resource)
				tmp, erx := json.Marshal(r)
				if erx != nil {
					return *ac, erx
				}
				erx = json.Unmarshal(tmp, ac)
				return *ac, erx
			}(tc.resource)
			if (err != nil) != tc.wantError {
				t.Fatalf("json.Unmarshal(json.Marshal(%+v)) = %+v, expected error? %+v", tc.resource, err, tc.wantError)
			}
			if !reflect.DeepEqual(actual, tc.wantResource) {
				t.Fatalf("json.Unmarshal(json.Marshal(%+v)) = <error>, expected %+v", tc.resource, actual, tc.wantResource)
			}
		})
	}
}
