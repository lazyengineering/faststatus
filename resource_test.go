// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package faststatus_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/lazyengineering/faststatus"
)

func TestResourceString(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		resource faststatus.Resource
	}{
		{"Zero Value",
			"00000000-0000-0000-0000-000000000000 free 0001-01-01T00:00:00Z",
			faststatus.Resource{},
		},
		{"Valid Busy",
			"01234567-89ab-cdef-0123-456789abcdef busy 2016-05-12T15:09:00-07:00 First One",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
		},
		{"Valid Free",
			"23456789-abcd-ef01-2345-6789abcdef01 free 2016-05-12T15:39:00-07:00 Second One",
			faststatus.Resource{
				ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: faststatus.Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:39:00-07:00")
					return tt
				}(),
				FriendlyName: "Second One",
			},
		},
		{"Valid Occupied",
			"456789ab-cdef-0123-4567-89abcdef0123 occupied 2016-05-12T15:40:00-07:00 Third One",
			faststatus.Resource{
				ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: faststatus.Occupied,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:40:00-07:00")
					return tt
				}(),
				FriendlyName: "Third One",
			},
		},
		{"Out of Range",
			"",
			faststatus.Resource{
				ID:     faststatus.ID{0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45},
				Status: faststatus.Occupied + 1,
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
		resource  faststatus.Resource
		wantBytes []byte
		wantError bool
	}{
		{"Zero Value",
			faststatus.Resource{},
			[]byte("00000000-0000-0000-0000-000000000000 free 0001-01-01T00:00:00Z"),
			false,
		},
		{"Valid Busy",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: faststatus.Free,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: faststatus.Occupied,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45},
				Status: faststatus.Occupied + 1,
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
		wantResource faststatus.Resource
	}{
		{"zero value",
			[]byte{},
			true,
			faststatus.Resource{},
		},
		{"valid busy",
			[]byte("01234567-89ab-cdef-0123-456789abcdef busy 2016-05-12T16:25:00-07:00 First One"),
			false,
			faststatus.Resource{
				ID:           faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				FriendlyName: "First One",
				Status:       faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
		},
		{"valid busy (numeric status)",
			[]byte("01234567-89ab-cdef-0123-456789abcdef 1 2016-05-12T16:25:00-07:00 First One"),
			false,
			faststatus.Resource{
				ID:           faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				FriendlyName: "First One",
				Status:       faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
			},
		},
		{"valid free",
			[]byte("23456789-abcd-ef01-2345-6789abcdef01 free 2016-05-12T16:27:00-07:00 Second One"),
			false,
			faststatus.Resource{
				ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: faststatus.Free,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: faststatus.Free,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: faststatus.Occupied,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: faststatus.Occupied,
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
			faststatus.Resource{},
		},
		{"invalid status",
			[]byte("01234567-89ab-cdef-0123-456789abcdef 4 2016-05-12T16:30:00-07:00 Another One"),
			true,
			faststatus.Resource{},
		},
		{"invalid since",
			[]byte("01234567-89ab-cdef-0123-456789abcdef busy 16-05-12T16:30:00-07:00 Another One"),
			true,
			faststatus.Resource{},
		},
		{"missing friendly name",
			[]byte("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa busy 2016-05-12T16:30:00-07:00"),
			false,
			faststatus.Resource{
				ID:     faststatus.ID{0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:30:00-07:00")
					return tt
				}(),
			},
		},
		{"missing timestamp",
			[]byte("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb busy"),
			true,
			faststatus.Resource{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var got faststatus.Resource
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
	f := func(r faststatus.Resource) bool {
		b, err := r.MarshalText()
		if err != nil {
			return false
		}
		got := new(faststatus.Resource)
		err = got.UnmarshalText(b)
		if err != nil {
			return false
		}
		return got.Equal(r)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestResourceMarshalJSON(t *testing.T) {
	testCases := []struct {
		name      string
		resource  faststatus.Resource
		wantValue []byte
		wantError bool
	}{
		{"Zero Value",
			faststatus.Resource{},
			[]byte(`{"id":"00000000-0000-0000-0000-000000000000","status":"free","since":"0001-01-01T00:00:00Z","friendlyName":""}`),
			false,
		},
		{"Valid Busy",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: faststatus.Free,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: faststatus.Occupied,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45},
				Status: faststatus.Occupied + 1,
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
		wantResource faststatus.Resource
		wantError    bool
	}{
		{"Zero Value",
			[]byte(`{}`),
			faststatus.Resource{},
			false,
		},
		{"Valid Busy",
			[]byte(`{
				"id":"01234567-89ab-cdef-0123-456789abcdef",
				"status":"1",
				"since":"2016-05-12T16:25:00-07:00",
				"friendlyName":"First One"
			}`),
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: faststatus.Free,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
				Status: faststatus.Free,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: faststatus.Occupied,
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
			faststatus.Resource{
				ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
				Status: faststatus.Occupied,
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
			faststatus.Resource{},
			true,
		},
		{"Bad ID",
			[]byte(`{
				"id":"01234567-89ab-cdef-0123-456789abcdef0",
				"friendlyName":"Third One",
				"status":"2",
				"since":"2016-05-12T16:28:00-07:00"
			}`),
			faststatus.Resource{},
			true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var actual faststatus.Resource
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
	f := func(r faststatus.Resource) bool {
		b, err := r.MarshalJSON()
		if err != nil {
			return false
		}
		got := new(faststatus.Resource)
		err = got.UnmarshalJSON(b)
		if err != nil {
			return false
		}
		return got.Equal(r)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestResourceEqual(t *testing.T) {
	testCases := []struct {
		name     string
		resource faststatus.Resource
		change   func(faststatus.Resource) faststatus.Resource
		want     bool
	}{
		{"zero value and zero value",
			faststatus.Resource{},
			func(r faststatus.Resource) faststatus.Resource { return faststatus.Resource{} },
			true,
		},
		{"real value and self value",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			func(r faststatus.Resource) faststatus.Resource { return r },
			true,
		},
		{"change in ID",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			func(r faststatus.Resource) faststatus.Resource {
				r.ID = faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01}
				return r
			},
			false,
		},
		{"change in Status",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			func(r faststatus.Resource) faststatus.Resource {
				r.Status = faststatus.Free
				return r
			},
			false,
		},
		{"change in FriendlyName",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			func(r faststatus.Resource) faststatus.Resource {
				r.FriendlyName = "Second Resource"
				return r
			},
			false,
		},
		{"change in Since time",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			func(r faststatus.Resource) faststatus.Resource {
				r.Since = r.Since.Add(time.Minute)
				return r
			},
			false,
		},
		{"change in Since location (actual time does not change)",
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
					return tt
				}(),
				FriendlyName: "First One",
			},
			func(r faststatus.Resource) faststatus.Resource {
				r.Since = r.Since.UTC()
				return r
			},
			true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			other := tc.change(tc.resource)
			got := tc.resource.Equal(other)
			if got != tc.want {
				t.Fatalf("%+v.Equal(%+v) = %+v, expected %+v", tc.resource, other, got, tc.want)
			}
		})
	}
}

func TestResourceEqualCommutative(t *testing.T) {
	f := func(a, b faststatus.Resource) bool {
		return a.Equal(b) == b.Equal(a) && a.Equal(a) && b.Equal(b)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestResourceMarshalBinaryHasMagicBytes(t *testing.T) {
	f := func(r faststatus.Resource) bool {
		b, _ := r.MarshalBinary()
		return len(b) >= 2 && bytes.Equal(b[0:2], faststatus.MagicBytes[:])
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestResourceMarshalBinaryVersionByte(t *testing.T) {
	f := func(r faststatus.Resource) bool {
		b, _ := r.MarshalBinary()
		return len(b) >= 3 && b[2] == faststatus.BinaryVersion
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestResourceMarshalBinaryLengthByte(t *testing.T) {
	f := func(r faststatus.Resource) bool {
		b, _ := r.MarshalBinary()
		return len(b) >= 4 && uint8(b[3]) == uint8(len(r.FriendlyName))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestResourceMarshalBinaryLength(t *testing.T) {
	f := func(r faststatus.Resource) bool {
		b, _ := r.MarshalBinary()
		return len(b) == 4+32+len(r.FriendlyName)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestResourceUnmarshalBinaryBadData(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte
	}{
		{"nil",
			nil,
		},
		{"empty",
			[]byte{},
		},
		{"bad magic bytes",
			func() []byte {
				b := make([]byte, 36)
				copy(b[0:4], []byte{0x47, 0x49, 0x46, 0x38}) // this would be a GIF
				return b
			}(),
		},
		{"version too high",
			func() []byte {
				var b = make([]byte, 36)
				copy(b, faststatus.MagicBytes[:])
				b[2] = byte(uint(faststatus.BinaryVersion) + 1)
				return b
			}(),
		},
		{"truncated data",
			func() []byte {
				b, _ := faststatus.Resource{
					ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
					Status: faststatus.Free,
					Since: func() time.Time {
						tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
						return tt
					}(),
					FriendlyName: "Second One",
				}.MarshalBinary()
				return b[0 : len(b)-5]
			}(),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var r = new(faststatus.Resource)
			err := r.UnmarshalBinary(tc.input)
			if err == nil {
				t.Fatalf("Resource.UnmarshalBinary(%v) = %v, expected error", tc.input, err)
			}
		})
	}
}

func TestResourceMarshalUnmarshalBinaryQuick(t *testing.T) {
	f := func(r faststatus.Resource) bool {
		b, err := r.MarshalBinary()
		if err != nil {
			return false
		}
		got := new(faststatus.Resource)
		err = got.UnmarshalBinary(b)
		if err != nil {
			return false
		}
		if !got.Equal(r) {
			t.Logf("Unmarshal(Marshal(%+v)) = %+v", r, got)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
