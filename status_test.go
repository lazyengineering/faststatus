// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package faststatus_test

import (
	"bytes" // because we explicitly want to work with the standard library json package
	"reflect"
	"testing"
	"testing/quick"

	"github.com/lazyengineering/faststatus"
)

func TestStatusMarshalBinary(t *testing.T) {
	isOneByteWithNoError := func(s faststatus.Status) bool {
		b, err := s.MarshalBinary()
		return err == nil && len(b) == 1
	}
	if err := quick.Check(isOneByteWithNoError, nil); err != nil {
		t.Error(err)
	}
}

func TestStatusUnmarshalBinary(t *testing.T) {
	f := func(b []byte) bool {
		s := new(faststatus.Status)
		err := s.UnmarshalBinary(b)
		if len(b) == 1 {
			return (err != nil) == (b[0] > byte(faststatus.Occupied))
		}
		return err != nil
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestStatusMarshalUnmarshalBinary(t *testing.T) {
	f := func(s faststatus.Status) bool {
		b, err := s.MarshalBinary()
		if err != nil {
			t.Logf("marshaling binary from status: %+v", err)
			return false
		}
		gotStatus := new(faststatus.Status)
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
		status                faststatus.Status
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
			faststatus.Free,
			[]byte("free"),
			false,
			false,
		},
		{"busy",
			faststatus.Busy,
			[]byte("busy"),
			false,
			false,
		},
		{"occupied",
			faststatus.Occupied,
			[]byte("occupied"),
			false,
			false,
		},
		{"out of range",
			faststatus.Occupied + 1,
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
			if faststatus.IsOutOfRange(err) != tc.wantErrorIsOutOfRange {
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
		wantStatus faststatus.Status
	}{
		{nil,
			true,
			faststatus.Free,
		},
		{[]byte(""),
			true,
			faststatus.Free,
		},
		{[]byte("Free"),
			false,
			faststatus.Free,
		},
		{[]byte("Busy"),
			false,
			faststatus.Busy,
		},
		{[]byte("Occupied"),
			false,
			faststatus.Occupied,
		},
		{[]byte("free"),
			false,
			faststatus.Free,
		},
		{[]byte("busy"),
			false,
			faststatus.Busy,
		},
		{[]byte("occupied"),
			false,
			faststatus.Occupied,
		},
		{[]byte("FREE"),
			false,
			faststatus.Free,
		},
		{[]byte("BUSY"),
			false,
			faststatus.Busy,
		},
		{[]byte("OCCUPIED"),
			false,
			faststatus.Occupied,
		},
		{[]byte("fReE"),
			false,
			faststatus.Free,
		},
		{[]byte("bUsY"),
			false,
			faststatus.Busy,
		},
		{[]byte("oCcUpIeD"),
			false,
			faststatus.Occupied,
		},
		{[]byte("0"),
			false,
			faststatus.Free,
		},
		{[]byte("1"),
			false,
			faststatus.Busy,
		},
		{[]byte("2"),
			false,
			faststatus.Occupied,
		},
		{[]byte("foo"),
			true,
			faststatus.Free,
		},
		{[]byte("freedom"),
			true,
			faststatus.Free,
		},
		{[]byte("busyness"),
			true,
			faststatus.Free,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.text), func(t *testing.T) {
			var s = new(faststatus.Status)
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
	f := func(s faststatus.Status) bool {
		txt, err := s.MarshalText()
		if err != nil {
			t.Logf("marshaling text from status: %+v", err)
			return false
		}
		gotStatus := new(faststatus.Status)
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
		Status   faststatus.Status
	}
	tests := []stringTest{
		{ // Zero Value
			Expected: "free",
		},
		{ // Free
			Expected: "free",
			Status:   faststatus.Free,
		},
		{ // Busy
			Expected: "busy",
			Status:   faststatus.Busy,
		},
		{ // Occupied
			Expected: "occupied",
			Status:   faststatus.Occupied,
		},
		{ // Out of Range
			Expected: "free",
			Status:   faststatus.Occupied + 1,
		},
	}
	for _, st := range tests {
		if actual := st.Status.String(); actual != st.Expected {
			t.Error("\nexpected:\t", st.Expected, "\n  actual:\t", actual)
		}
	}
}
