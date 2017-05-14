// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package store_test

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/boltdb/bolt"

	"github.com/lazyengineering/faststatus"
	"github.com/lazyengineering/faststatus/store"
)

func TestSave(t *testing.T) {
	db, cleanup := newEmptyDB(t)
	defer cleanup()

	// order of test cases matters here (these are not stateless)
	testCases := []struct {
		name      string
		store     *store.Store
		resource  faststatus.Resource
		wantError func(error) bool
	}{
		{"Save should return an error if the store is nil",
			nil,
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
			},
			func(e error) bool { return e != nil },
		},
		{"Save should return an error if the database is not initialized",
			&store.Store{},
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
			},
			func(e error) bool { return e != nil },
		},
		{"Save should return an error for a resource with a zero-value ID",
			&store.Store{DB: db},
			faststatus.Resource{
				ID:     faststatus.ID{},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
			},
			store.ZeroValueError,
		},
		{"Save should not return an error for a new resource",
			&store.Store{DB: db},
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:09:00-07:00")
					return tt
				}(),
			},
			func(e error) bool { return e == nil },
		},
		{"Save should return an error if there is a more recent version already saved",
			&store.Store{DB: db},
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Busy,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:00:00-07:00")
					return tt
				}(),
			},
			faststatus.ConflictError,
		},
		{"Save should not return an error for the most recent valid resource",
			&store.Store{DB: db},
			faststatus.Resource{
				ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
				Status: faststatus.Free,
				Since: func() time.Time {
					tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:15:00-07:00")
					return tt
				}(),
			},
			func(e error) bool { return e == nil },
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.store.Save(tc.resource)
			if !tc.wantError(err) {
				t.Fatalf("%+v.Save(%+v) error checks false, expected true from %+v", tc.store, tc.resource, err)
			}
		})
	}
}

func TestSaveIsIdempotent(t *testing.T) {
	db, cleanup := newEmptyDB(t)
	defer cleanup()

	s := &store.Store{DB: db}

	r := faststatus.Resource{
		ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		Status: faststatus.Free,
		Since: func() time.Time {
			tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
			return tt
		}(),
	}

	for i := 0; i < 20; i++ {
		err := s.Save(r)
		if err != nil {
			t.Fatalf("unexpected error saving resource: %+v", err)
		}
		got, err := s.Get(r.ID)
		if err != nil {
			t.Fatalf("unexpected error getting final resource: %+v", err)
		}
		if !got.Equal(r) {
			t.Fatalf("getting resource for the %d time: got %+v, expected %+v", i+1, got, r)
		}
	}
}

func TestSaveIsConcurrencySafe(t *testing.T) {
	db, cleanup := newEmptyDB(t)
	defer cleanup()

	s := &store.Store{DB: db}

	testResources := []faststatus.Resource{
		{
			ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Status: faststatus.Busy,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
				return tt
			}(),
		},
		{
			ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
			Status: faststatus.Free,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
				return tt
			}(),
		},
		{
			ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
			Status: faststatus.Occupied,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
				return tt
			}(),
		},
		{
			ID: faststatus.ID{
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
			},
			Status: faststatus.Busy,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
				return tt
			}(),
		},
		{
			ID: faststatus.ID{
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
			},
			Status: faststatus.Free,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:01-07:00")
				return tt
			}(),
		},
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, r := range testResources {
		wg.Add(1)
		go func(r faststatus.Resource) {
			defer wg.Done()
			<-start
			err := s.Save(r)
			if err != nil {
				t.Fatalf("no errors expected for concurrency test: %+v", err)
			}
		}(r)
	}
	close(start)
	wg.Wait()
}

func TestSaveStoresOnlyLatest(t *testing.T) {
	// Save should store only the version of a resource with the highest timestamp,
	// regardless the order of arrival. This test should focus on concurrent saves
	// (sequential saves are covered in the table test above)
	db, cleanup := newEmptyDB(t)
	defer cleanup()

	s := &store.Store{DB: db}

	testResources := map[string]faststatus.Resource{
		"first": {
			ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Status: faststatus.Free,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
				return tt
			}(),
		},
		"second": {
			ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Status: faststatus.Busy,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:01-07:00")
				return tt
			}(),
		},
		"third": {
			ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Status: faststatus.Occupied,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:02-07:00")
				return tt
			}(),
		},
		"fourth": {
			ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Status: faststatus.Free,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:03-07:00")
				return tt
			}(),
		},
		"final": {
			ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Status: faststatus.Busy,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:04-07:00")
				return tt
			}(),
		},
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, r := range testResources {
		wg.Add(1)
		go func(r faststatus.Resource) {
			defer wg.Done()
			<-start
			// some errors are expected, but not always
			_ = s.Save(r)
		}(r)
	}
	close(start)
	wg.Wait()
	got, err := s.Get(testResources["final"].ID)
	if err != nil {
		t.Fatalf("unexpected error getting final resource: %+v", err)
	}
	if !got.Equal(testResources["final"]) {
		t.Fatalf("getting final resource: got %+v, expected %+v", got, testResources["final"])
	}
}

func TestGet(t *testing.T) {
	db, cleanup := newEmptyDB(t)
	defer cleanup()

	stockResources := map[string]faststatus.Resource{
		"valid": {
			ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Status: faststatus.Busy,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:00:00-07:00")
				return tt
			}(),
		},
		"not-found": {
			ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
			Status: faststatus.Busy,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T15:00:00-07:00")
				return tt
			}(),
		},
	}

	if err := (&store.Store{DB: db}).Save(stockResources["valid"]); err != nil {
		t.Fatalf("saving stock resource for test: %+v", err)
	}

	testCases := []struct {
		name         string
		store        *store.Store
		id           faststatus.ID
		wantError    func(error) bool
		wantResource faststatus.Resource
	}{
		{"Get should return an error if the store is nil",
			nil,
			stockResources["valid"].ID,
			func(e error) bool { return e != nil },
			faststatus.Resource{},
		},
		{"Get should return an error if the database is not initialized",
			&store.Store{},
			stockResources["valid"].ID,
			func(e error) bool { return e != nil },
			faststatus.Resource{},
		},
		{"Get should return an error for a zero-value ID",
			&store.Store{DB: db},
			faststatus.ID{},
			store.ZeroValueError,
			faststatus.Resource{},
		},
		{"Get should return a valid Resource with the input ID",
			&store.Store{DB: db},
			stockResources["valid"].ID,
			func(e error) bool { return e == nil },
			stockResources["valid"],
		},
		{"Get should return a zero-value Resource when none found",
			&store.Store{DB: db},
			stockResources["not-found"].ID,
			func(e error) bool { return e == nil },
			faststatus.Resource{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gotResource, err := tc.store.Get(tc.id)
			if !tc.wantError(err) {
				t.Fatalf("%+v.Get(%+v) error checks false, expected true from %+v", tc.store, tc.id, err)
			}
			if !gotResource.Equal(tc.wantResource) {
				t.Fatalf("%+v.Get(%+v) = %+v <error>, expected %+v", tc.store, tc.id, gotResource, tc.wantResource)
			}
		})
	}
}

func TestGetIsConcurrencySafe(t *testing.T) {
	db, cleanup := newEmptyDB(t)
	defer cleanup()

	s := &store.Store{DB: db}

	testResources := []faststatus.Resource{
		{
			ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			Status: faststatus.Busy,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
				return tt
			}(),
		},
		{
			ID:     faststatus.ID{0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01},
			Status: faststatus.Free,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:27:00-07:00")
				return tt
			}(),
		},
		{
			ID:     faststatus.ID{0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23},
			Status: faststatus.Occupied,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:28:00-07:00")
				return tt
			}(),
		},
		{
			ID:     faststatus.ID{0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45},
			Status: faststatus.Busy,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
				return tt
			}(),
		},
		{
			ID:     faststatus.ID{0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67},
			Status: faststatus.Free,
			Since: func() time.Time {
				tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:01-07:00")
				return tt
			}(),
		},
	}

	for _, r := range testResources {
		err := s.Save(r)
		if err != nil {
			t.Fatalf("no errors expected for concurrency test: %+v", err)
		}
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, r := range testResources {
		wg.Add(1)
		go func(r faststatus.Resource) {
			defer wg.Done()
			<-start
			got, err := s.Get(r.ID)
			if err != nil {
				t.Fatalf("no errors expected for concurrency test: %+v", err)
			}
			if !got.Equal(r) {
				t.Fatalf("getting test resource from store: got %+v, expected %+v", got, r)
			}
		}(r)
	}
	close(start)
	wg.Wait()
}

func TestGetIsIdempotent(t *testing.T) {
	db, cleanup := newEmptyDB(t)
	defer cleanup()

	s := &store.Store{DB: db}

	r := faststatus.Resource{
		ID:     faststatus.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		Status: faststatus.Free,
		Since: func() time.Time {
			tt, _ := time.Parse(time.RFC3339, "2016-05-12T16:25:00-07:00")
			return tt
		}(),
	}
	err := s.Save(r)
	if err != nil {
		t.Fatalf("unexpected error saving resource: %+v", err)
	}

	for i := 0; i < 20; i++ {
		got, err := s.Get(r.ID)
		if err != nil {
			t.Fatalf("unexpected error getting final resource: %+v", err)
		}
		if !got.Equal(r) {
			t.Fatalf("getting resource for the %d time: got %+v, expected %+v", i+1, got, r)
		}
	}
}

func newEmptyDB(t *testing.T) (*bolt.DB, func()) {
	path, cleanup := tempfile(t)
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		cleanup()
		t.Fatalf("opening new database for tests: %+v", err)
	}
	return db, func() {
		defer cleanup()
		if err := db.Close(); err != nil {
			t.Fatalf("closing database: %+v", err)
		}
	}
}

func tempfile(t *testing.T) (string, func()) {
	tmpfile, err := ioutil.TempFile("", "_test")
	if err != nil {
		t.Fatalf("creating test file: %+v", err)
	}
	fileName := tmpfile.Name()
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("closing test file: %+v", err)
	}
	return fileName, func() {
		os.Remove(fileName)
	}
}
