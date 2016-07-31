// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// A Resource represents any resource (a person, a bathroom, a server, etc.)
// that needs to communicate how busy it is.
type Resource struct {
	Id           uint64
	FriendlyName string
	Status       Status
	Since        time.Time
}

const resourceFmtString = "%016X %v %s %s"

// String will return a single-line representation of a resource.
// In order to optimize for standard streams, the output is as follows:
//   {{Id}} {{Status}} {{Since}} {{FriendlyName}}
// Formatted as follows:
//   0123456789ABCDEF 1 2006-01-02T15:04:05Z07:00 My Resource
func (r Resource) String() string {
	return fmt.Sprintf(resourceFmtString, r.Id, r.Status, r.Since.Format(time.RFC3339), r.FriendlyName)
}

// MarshalJSON will return simple a simple json structure for a resource.
// Will not accept any Status that is out of range; see Status documentation
// for more information.
func (r Resource) MarshalJSON() ([]byte, error) {
	tmpResource := struct {
		Id           string    `json:"id"`
		FriendlyName string    `json:"friendlyName"`
		Status       Status    `json:"status"`
		Since        time.Time `json:"since"`
	}{
		fmt.Sprintf("%X", r.Id),
		r.FriendlyName,
		r.Status,
		r.Since,
	}
	return json.Marshal(tmpResource)
}

// UnmarshalJson will populate a Resource with data from a json struct
// according to the same format as MarshalJSON. Will overwrite any values
// already assigned to the Resource.
func (r *Resource) UnmarshalJSON(raw []byte) error {
	// allow zero values with omitempty
	tmp := new(struct {
		Id           string    `json:",omitempty"`
		FriendlyName string    `json:",omitempty"`
		Status       Status    `json:",omitempty"`
		Since        time.Time `json:",omitempty"`
	})
	if err := json.Unmarshal(raw, tmp); err != nil {
		return err
	}

	if len(tmp.Id) == 0 {
		tmp.Id = "0"
	}
	if id, err := strconv.ParseUint(tmp.Id, 16, 64); err != nil {
		return err
	} else {
		r.Id = id
	}

	r.FriendlyName = tmp.FriendlyName
	r.Status = tmp.Status
	r.Since = tmp.Since
	return nil
}
