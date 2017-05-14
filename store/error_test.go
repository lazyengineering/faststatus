// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package store

import (
	"errors"
	"testing"

	"github.com/lazyengineering/faststatus"
)

func TestDataError(t *testing.T) {
	var allFalse1 error = &dataError{}
	var allFalse2 error = &dataError{}
	if allFalse1.Error() != allFalse2.Error() {
		t.Fatalf("expected %s == %s", allFalse1.Error(), allFalse2.Error())
	}
	if allFalse1.Error() != allFalse1.Error() {
		t.Fatalf("expected %s == %s", allFalse1.Error(), allFalse1.Error())
	}

	var oldOnly1 error = &dataError{old: true}
	var oldOnly2 error = &dataError{old: true}
	if oldOnly1.Error() != oldOnly2.Error() {
		t.Fatalf("expected %s == %s", oldOnly1.Error(), oldOnly2.Error())
	}
	if oldOnly1.Error() != oldOnly1.Error() {
		t.Fatalf("expected %s == %s", oldOnly1.Error(), oldOnly1.Error())
	}
	if oldOnly1.Error() == allFalse1.Error() {
		t.Fatalf("expected %s != %s", oldOnly1.Error(), allFalse1.Error())
	}
	if !faststatus.ConflictError(oldOnly1) {
		t.Fatalf("faststatus.ConflictError(%+v) = false, expected true", oldOnly1)
	}

	var noIDOnly1 error = &dataError{noID: true}
	var noIDOnly2 error = &dataError{noID: true}
	if noIDOnly1.Error() != noIDOnly2.Error() {
		t.Fatalf("expected %s == %s", noIDOnly1.Error(), noIDOnly2.Error())
	}
	if noIDOnly1.Error() != noIDOnly1.Error() {
		t.Fatalf("expected %s == %s", noIDOnly1.Error(), noIDOnly1.Error())
	}
	if noIDOnly1.Error() == allFalse1.Error() {
		t.Fatalf("expected %s != %s", noIDOnly1.Error(), allFalse1.Error())
	}
	if noIDOnly1.Error() == oldOnly1.Error() {
		t.Fatalf("expected %s != %s", noIDOnly1.Error(), oldOnly1.Error())
	}

	var allTrue1 error = &dataError{old: true, noID: true}
	var allTrue2 error = &dataError{old: true, noID: true}
	if allTrue1.Error() != allTrue2.Error() {
		t.Fatalf("expected %s == %s", allTrue1.Error(), allTrue2.Error())
	}
	if allTrue1.Error() != allTrue1.Error() {
		t.Fatalf("expected %s == %s", allTrue1.Error(), allTrue1.Error())
	}
	if allTrue1.Error() == allFalse1.Error() {
		t.Fatalf("expected %s != %s", allTrue1.Error(), allFalse1.Error())
	}
	if allTrue1.Error() == oldOnly1.Error() {
		t.Fatalf("expected %s != %s", noIDOnly1.Error(), oldOnly1.Error())
	}
	if allTrue1.Error() == noIDOnly1.Error() {
		t.Fatalf("expected %s != %s", allTrue1.Error(), noIDOnly1.Error())
	}
	if !faststatus.ConflictError(allTrue1) {
		t.Fatalf("faststatus.ConflictError(%+v) = false, expected true", allTrue1)
	}
}

func TestZeroValueError(t *testing.T) {
	testCases := []struct {
		name          string
		err           error
		wantZeroValue bool
	}{
		{"nil",
			nil,
			false,
		},
		{"new string",
			errors.New("an error"),
			false,
		},
		{"zero-value dataError",
			dataError{},
			false,
		},
		{"zero-id dataError",
			dataError{noID: true},
			true,
		},
		{"old dataError",
			dataError{old: true},
			false,
		},
		{"old zero-id dataError",
			dataError{old: true, noID: true},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := ZeroValueError(tc.err)
			if got != tc.wantZeroValue {
				t.Fatalf("ZeroValueError(%+v) = %v, expected %v", tc.err, got, tc.wantZeroValue)
			}
		})
	}
}
