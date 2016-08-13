// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/lazyengineering/faststatus/resource"
)

type boltStore struct {
	db *bolt.DB
}

// BoltStore initializes a storage engine for the current server built on boltdb persistance. Use as an option when calling `Current()`.
func BoltStore(dbFile string) func(*current) error {
	return func(c *current) error {
		s := new(boltStore)
		if err := s.init(dbFile); err != nil {
			return fmt.Errorf("creating new store: %+v", err)
		}
		c.store = s
		return nil
	}
}

func (s *boltStore) init(dbPath string) error {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("Error initializing database, %q: %v", dbPath, err)
	}
	s.db = db
	return nil
}

func (s *boltStore) save(r resource.Resource) error {
	return fmt.Errorf("Not implemented!")
}

// get returns a slice of resources from the underlying bolt database.
// An empty resultset is not an error, just an empty slice.
func (s *boltStore) get(ids ...uint64) ([]resource.Resource, error) {
	var resources []resource.Resource

	if len(ids) == 0 {
		return resources, nil
	}

	rch := make(chan []byte)
	done := make(chan struct{})
	defer close(done)
	go s.db.View(func(tx *bolt.Tx) error {
		defer close(rch)
		b := tx.Bucket([]byte("resources"))
		for _, id := range ids {
			raw := b.Get([]byte(strconv.FormatUint(id, 16)))
			select {
			case rch <- raw:
			case <-done:
				return nil
			}
		}
		return nil
	})

	for raw := range rch {
		if raw == nil {
			continue
		}
		rc := new(resource.Resource)
		err := rc.UnmarshalJSON(raw)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling Resource JSON: %+v", err)
		}
		resources = append(resources, *rc)
	}
	return resources, nil
}
