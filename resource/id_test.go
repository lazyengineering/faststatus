package resource_test

import (
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
