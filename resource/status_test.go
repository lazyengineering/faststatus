// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"bytes" // because we explicitly want to work with the standard library json package
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// Generate is used in testing to generate only valid Status values
func (s Status) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(Status(rand.Int() % int(Occupied)))
}

func TestStatusMarshalBinary(t *testing.T) {
	isOneByteWithNoError := func(s Status) bool {
		b, err := s.MarshalBinary()
		return err == nil && len(b) == 1
	}
	if err := quick.Check(isOneByteWithNoError, nil); err != nil {
		t.Error(err)
	}
}

func TestStatusUnmarshalBinary(t *testing.T) {
	f := func(b []byte) bool {
		s := new(Status)
		err := s.UnmarshalBinary(b)
		if len(b) == 1 {
			return (err != nil) == (b[0] > byte(Occupied))
		}
		return err != nil
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestStatusMarshalUnmarshalBinary(t *testing.T) {
	f := func(s Status) bool {
		b, err := s.MarshalBinary()
		if err != nil {
			t.Logf("marshaling binary from status: %+v", err)
			return false
		}
		gotStatus := new(Status)
		err = gotStatus.UnmarshalBinary(b)
		if err != nil {
			t.Logf("unmarshaling binary from status: %+v", err)
			return false
		}
		return reflect.DeepEqual(*gotStatus, s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestStatusMarshalText(t *testing.T) {
	tests := []struct {
		name                  string
		status                Status
		wantBytes             []byte
		wantError             bool
		wantErrorIsOutOfRange bool
	}{
		{"zero value",
			0,
			[]byte("free"),
			false,
			false,
		},
		{"free",
			Free,
			[]byte("free"),
			false,
			false,
		},
		{"busy",
			Busy,
			[]byte("busy"),
			false,
			false,
		},
		{"occupied",
			Occupied,
			[]byte("occupied"),
			false,
			false,
		},
		{"out of range",
			Occupied + 1,
			nil,
			true,
			true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.status.MarshalText()
			if (err == nil) == tc.wantError {
				t.Fatalf("%+v.MarshalText() = []byte, %+v, expected error? %+v", tc.status, err, tc.wantError)
			}
			if IsOutOfRange(err) != tc.wantErrorIsOutOfRange {
				t.Fatalf("%+v.MarshalText() = []byte, %+v, expected error out of range? %+v", tc.status, err, tc.wantErrorIsOutOfRange)
			}
			if !bytes.Equal(b, tc.wantBytes) {
				t.Fatalf("%+v.MarshalText() = %+v, error, expected %+v", tc.status, b, tc.wantBytes)
			}
		})
	}
}

func TestStatusUnmarshalText(t *testing.T) {
	testCases := []struct {
		text       []byte
		wantError  bool
		wantStatus Status
	}{
		{nil,
			true,
			Free,
		},
		{[]byte(""),
			true,
			Free,
		},
		{[]byte("Free"),
			false,
			Free,
		},
		{[]byte("Busy"),
			false,
			Busy,
		},
		{[]byte("Occupied"),
			false,
			Occupied,
		},
		{[]byte("free"),
			false,
			Free,
		},
		{[]byte("busy"),
			false,
			Busy,
		},
		{[]byte("occupied"),
			false,
			Occupied,
		},
		{[]byte("FREE"),
			false,
			Free,
		},
		{[]byte("BUSY"),
			false,
			Busy,
		},
		{[]byte("OCCUPIED"),
			false,
			Occupied,
		},
		{[]byte("fReE"),
			false,
			Free,
		},
		{[]byte("bUsY"),
			false,
			Busy,
		},
		{[]byte("oCcUpIeD"),
			false,
			Occupied,
		},
		{[]byte("0"),
			false,
			Free,
		},
		{[]byte("1"),
			false,
			Busy,
		},
		{[]byte("2"),
			false,
			Occupied,
		},
		{[]byte("foo"),
			true,
			Free,
		},
		{[]byte("freedom"),
			true,
			Free,
		},
		{[]byte("busyness"),
			true,
			Free,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.text), func(t *testing.T) {
			var s = new(Status)
			err := s.UnmarshalText(tc.text)
			if (err != nil) != tc.wantError {
				t.Fatalf("%+v.UnmarshalText(%+v) = %+v, expected error? %+v", *s, tc.text, err, tc.wantError)
			}
			if *s != tc.wantStatus {
				t.Fatalf("%+v.UnmarshalText(%+v) = error, expected %+v", *s, tc.text, tc.wantStatus)
			}
		})
	}
}

func TestStatusMarshalUnmarshalText(t *testing.T) {
	f := func(s Status) bool {
		txt, err := s.MarshalText()
		if err != nil {
			t.Logf("marshaling text from status: %+v", err)
			return false
		}
		gotStatus := new(Status)
		err = gotStatus.UnmarshalText(txt)
		if err != nil {
			t.Logf("unmarshaling text from status: %+v", err)
			return false
		}
		return reflect.DeepEqual(*gotStatus, s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestStatusString(t *testing.T) {
	type stringTest struct {
		Expected string
		Status   Status
	}
	tests := []stringTest{
		stringTest{ // Zero Value
			Expected: "free",
		},
		stringTest{ // Free
			Expected: "free",
			Status:   Free,
		},
		stringTest{ // Busy
			Expected: "busy",
			Status:   Busy,
		},
		stringTest{ // Occupied
			Expected: "occupied",
			Status:   Occupied,
		},
		stringTest{ // Out of Range
			Expected: "free",
			Status:   Occupied + 1,
		},
	}
	for _, st := range tests {
		if actual := st.Status.String(); actual != st.Expected {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", actual)
		}
	}
}
