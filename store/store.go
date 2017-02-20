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
		return errorStoreNotInitialized
	}
	if s.DB == nil {
		return errorDBNotInitialized
	}
	if r.ID == (faststatus.ID{}) {
		return fmt.Errorf("cannot save a resource with a zero-value ID")
	}
	key, err := r.ID.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshaling binary key from resource ID: %+v", err)
	}

	err = s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("creating bucket: %+v", err)
		}

		latest := b.Get(key)
		if len(latest) > 0 {
			latestResource := new(faststatus.Resource)
			if err := latestResource.UnmarshalText(latest); err != nil {
				return fmt.Errorf("unmarshaling latest stored resource: %+v", err)
			}
			if latestResource.Since.After(r.Since) {
				return errorMoreRecentVersion
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

// Get returns the most recent state of the Resource with the given valid ID
// or a zero-value Resource if it does not exist in the Store.
func (s *Store) Get(id faststatus.ID) (faststatus.Resource, error) {
	if s == nil {
		return faststatus.Resource{}, errorStoreNotInitialized
	}
	if s.DB == nil {
		return faststatus.Resource{}, errorDBNotInitialized
	}
	if id == (faststatus.ID{}) {
		return faststatus.Resource{}, fmt.Errorf("cannot get a resource with a zero-value ID")
	}
	key, err := id.MarshalBinary()
	if err != nil {
		return faststatus.Resource{}, fmt.Errorf("failed to marshal key from id: %+v", err)
	}

	r := new(faststatus.Resource)
	err = s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return nil
		}
		raw := b.Get(key)
		if len(raw) == 0 {
			return nil
		}
		if err := r.UnmarshalText(raw); err != nil {
			return fmt.Errorf("unmarshaling resource from stored value: %+v", err)
		}
		return nil
	})
	if err != nil {
		return faststatus.Resource{}, fmt.Errorf("viewing database with resource: %+v", err)
	}
	return *r, nil
}

var (
	errorStoreNotInitialized = fmt.Errorf("store not initialized")
	errorDBNotInitialized    = fmt.Errorf("no bolt database for store")
	errorMoreRecentVersion   = fmt.Errorf("a more recent version of this resource already exists in store")
)

var (
	bucketName = []byte("faststatus/store")
)
