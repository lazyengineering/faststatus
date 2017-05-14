// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package store

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"

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
		return dataError{noID: true}
	}
	key, err := r.ID.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "marshaling binary key from resource ID")
	}

	err = s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return errors.Wrap(err, "creating bucket")
		}

		latest := b.Get(key)
		if len(latest) > 0 {
			latestResource := new(faststatus.Resource)
			if err := latestResource.UnmarshalBinary(latest); err != nil {
				return errors.Wrap(err, "unmarshaling latest stored resource")
			}
			if latestResource.Since.After(r.Since) {
				return dataError{old: true}
			}
		}
		payload, err := r.MarshalBinary()
		if err != nil {
			return errors.Wrap(err, "marshaling text for resource payload")
		}
		if err := b.Put(key, payload); err != nil {
			return errors.Wrap(err, "putting resource in bucket")
		}
		return nil
	})
	return errors.Wrap(err, "updating database with resource")
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
		return faststatus.Resource{}, dataError{noID: true}
	}
	key, err := id.MarshalBinary()
	if err != nil {
		return faststatus.Resource{}, errors.Wrap(err, "failed to marshal key from id")
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
		if err := r.UnmarshalBinary(raw); err != nil {
			return errors.Wrap(err, "unmarshaling resource from stored value")
		}
		return nil
	})
	if err != nil {
		return faststatus.Resource{}, errors.Wrap(err, "viewing database with resource")
	}
	return *r, nil
}

var (
	errorStoreNotInitialized = fmt.Errorf("store not initialized")
	errorDBNotInitialized    = fmt.Errorf("no bolt database for store")
)

var (
	bucketName = []byte("faststatus/store")
)
