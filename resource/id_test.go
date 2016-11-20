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

func TestMarshalBinary(t *testing.T) {
	is16Bytes := func(id resource.ID) bool {
		b, err := id.MarshalBinary()
		return err == nil && len(b) == 16
	}
	if err := quick.Check(is16Bytes, nil); err != nil {
		t.Error(err)
	}
}

func TestUnmarshalBinary(t *testing.T) {
	f := func(b []byte) bool {
		id := new(resource.ID)
		err := id.UnmarshalBinary(b)
		return (err == nil) == (len(b) == 16)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestMarshalUnmarshalBinary(t *testing.T) {
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

func TestUnmarshalMarshalBinary(t *testing.T) {
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

func TestMarshalTextIs36Bytes(t *testing.T) {
	// is 36 bytes
	is36bytes := func(id resource.ID) bool {
		s, err := id.MarshalText()
		return err == nil && len(s) == 36
	}
	if err := quick.Check(is36bytes, nil); err != nil {
		t.Fatal(err)
	}
}

func TestMarshalTextIsValidChars(t *testing.T) {
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

func TestMarshalTextHasCorrectDashes(t *testing.T) {
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
