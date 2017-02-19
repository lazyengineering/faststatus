// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package store

import (
	"fmt"

	"github.com/boltdb/bolt"

	"github.com/lazyengineering/faststatus"
)

// Store persists the most recent version of Resources by ID
type Store struct {
	DB *bolt.DB
}

// Save persists a Resource to the Store iff it is the most recent
func (s *Store) Save(r faststatus.Resource) error {
	if s == nil {
		return fmt.Errorf("store not initialized")
	}
	if s.DB == nil {
		return fmt.Errorf("no bolt database for store")
	}
	var zeroID = faststatus.ID{}
	if r.ID == zeroID {
		return fmt.Errorf("cannot save a resource with a zero-value ID")
	}
	err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("faststatus/store"))
		if err != nil {
			return fmt.Errorf("creating bucket: %+v", err)
		}
		key, err := r.ID.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshaling binary key from resource ID: %+v", err)
		}

		latest := b.Get(key)
		if len(latest) > 0 {
			latestResource := new(faststatus.Resource)
			if err := latestResource.UnmarshalText(latest); err != nil {
				return fmt.Errorf("unmarshaling latest stored resource: %+v", err)
			}
			if latestResource.Since.After(r.Since) {
				return fmt.Errorf("a more recent versoin of this resource already exists in store")
			}
		}
		//TODO: Implement MarshalBinary() for Resource
		payload, err := r.MarshalText()
		if err != nil {
			return fmt.Errorf("marshaling text for resource payload: %+v", err)
		}
		if err := b.Put(key, payload); err != nil {
			return fmt.Errorf("putting resource in bucket: %+v", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("updating database with resource: %+v", err)
	}
	return nil
}
