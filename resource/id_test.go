// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource_test

import (
	"bytes"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/lazyengineering/faststatus/resource"
)

func TestNewIDIsV4(t *testing.T) {
	isV4 := func() bool {
		id, err := resource.NewID()
		return err == nil && uint(id[6]>>4) == 4
	}
	if err := quick.Check(isV4, nil); err != nil {
		t.Error(err)
	}
}

func TestNewIDIsUUID(t *testing.T) {
	isUUID := func() bool {
		id, err := resource.NewID()
		return err == nil && (id[8]&0xc0)|0x80 == 0x80
	}
	if err := quick.Check(isUUID, nil); err != nil {
		t.Error(err)
	}
}

func TestIDMarshalBinary(t *testing.T) {
	is16Bytes := func(id resource.ID) bool {
		b, err := id.MarshalBinary()
		return err == nil && len(b) == 16
	}
	if err := quick.Check(is16Bytes, nil); err != nil {
		t.Error(err)
	}
}

func TestIDUnmarshalBinary(t *testing.T) {
	f := func(b []byte) bool {
		id := new(resource.ID)
		err := id.UnmarshalBinary(b)
		return (err == nil) == (len(b) == 16)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestIDMarshalUnmarshalBinary(t *testing.T) {
	f := func(id resource.ID) bool {
		b, err := id.MarshalBinary()
		if err != nil {
			t.Fatalf("unexpected error marshaling binary from id: %+v", err)
		}
		gotID := new(resource.ID)
		err = gotID.UnmarshalBinary(b)
		if err != nil {
			t.Fatalf("unexpected error unmarshaling id from binary: %+v", err)
		}
		return reflect.DeepEqual(*gotID, id)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestIDUnmarshalMarshalBinary(t *testing.T) {
	f := func(ba [16]byte) bool {
		id := new(resource.ID)
		err := id.UnmarshalBinary(ba[:])
		if err != nil {
			t.Fatalf("unexpected error unmarshaling id from binary: %+v", err)
		}
		b, err := id.MarshalBinary()
		if err != nil {
			t.Fatalf("unexpected error marshaling binary from id: %+v", err)
		}
		return reflect.DeepEqual(b, ba[:])
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestIDMarshalTextIs36Bytes(t *testing.T) {
	// is 36 bytes
	is36bytes := func(id resource.ID) bool {
		s, err := id.MarshalText()
		return err == nil && len(s) == 36
	}
	if err := quick.Check(is36bytes, nil); err != nil {
		t.Fatal(err)
	}
}

func TestIDMarshalTextIsValidChars(t *testing.T) {
	// contains only lowercase hex and dashes
	onlyHexAndDashes := func(id resource.ID) bool {
		s, err := id.MarshalText()
		if err != nil {
			return false
		}
		for _, r := range bytes.Runes(s) {
			switch r {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', '-':
			default:
				return false
			}
		}
		return true
	}
	if err := quick.Check(onlyHexAndDashes, nil); err != nil {
		t.Fatal(err)
	}
}

func TestIDMarshalTextHasCorrectDashes(t *testing.T) {
	// contains dashes where expected
	dashesWhereExpected := func(id resource.ID) bool {
		s, err := id.MarshalText()
		if err != nil {
			return false
		}
		for _, i := range []int{8, 13, 18, 23} {
			if s[i] != '-' {
				return false
			}
		}
		return true
	}
	if err := quick.Check(dashesWhereExpected, nil); err != nil {
		t.Fatal(err)
	}
}

func TestIDUnmarshalText(t *testing.T) {
	testCases := []struct {
		name      string
		txt       []byte
		wantError bool
		wantID    resource.ID
	}{
		{"nilish",
			nil,
			true,
			resource.ID{},
		},
		{"non-hex in first block",
			[]byte("0123456--0000-0000-0000-000000000000"),
			true,
			resource.ID{},
		},
		{"non-hex in second block",
			[]byte("01234567-89az-0000-0000-000000000000"),
			true,
			resource.ID{},
		},
		{"non-hex in second block, no dash",
			[]byte("0123456789az-0000-0000-000000000000"),
			true,
			resource.ID{},
		},
		{"non-hex in third block",
			[]byte("01234567-89ab--def-0000-000000000000"),
			true,
			resource.ID{},
		},
		{"non-hex in third block, no first dash",
			[]byte("0123456789ab--def-0000-000000000000"),
			true,
			resource.ID{},
		},
		{"non-hex in third block, no dashes",
			[]byte("0123456789abjdef-0000-000000000000"),
			true,
			resource.ID{},
		},
		{"non-hex in fourth block",
			[]byte("01234567-89ab-cdef-0g23-000000000000"),
			true,
			resource.ID{},
		},
		{"non-hex in fourth block, missing a dash",
			[]byte("01234567-89abcdef-0g23-000000000000"),
			true,
			resource.ID{},
		},
		{"non-hex in final block",
			[]byte("01234567-89ab-cdef-0123-456789@bcdef"),
			true,
			resource.ID{},
		},
		{"non-hex in final block, no dashes",
			[]byte("0123456789abcdef0123456789@bcdef"),
			true,
			resource.ID{},
		},
		{"non-hex in final block, missing a dash",
			[]byte("01234567-89ab-cdef-0123456789@bcdef"),
			true,
			resource.ID{},
		},
		{"too long",
			[]byte("01234567-89ab-cdef-0123-456789abcdef0"),
			true,
			resource.ID{},
		},
		{"happy path",
			[]byte("01234567-89ab-cdef-0123-456789abcdef"),
			false,
			resource.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			id := new(resource.ID)
			err := id.UnmarshalText(tc.txt)
			if (err != nil) != tc.wantError {
				t.Fatalf("id.UnmarshalText(%+v) = %+v, expected error? %+v", tc.txt, err, tc.wantError)
			}
			if !reflect.DeepEqual(*id, tc.wantID) {
				t.Fatalf("id.UnmarshalText(%+v) -> id = %+v, expected %+v", tc.txt, *id, &tc.wantID)
			}
		})
	}
}

func TestIDMarshalUnmarshalText(t *testing.T) {
	f := func(id resource.ID) bool {
		b, err := id.MarshalText()
		if err != nil {
			t.Fatalf("unexpected error marshaling text from id: %+v", err)
		}
		gotID := new(resource.ID)
		err = gotID.UnmarshalText(b)
		if err != nil {
			t.Fatalf("unexpected error unmarshaling id from text: %+v", err)
		}
		return reflect.DeepEqual(*gotID, id)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
