// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// A Resource represents any resource (a person, a bathroom, a server, etc.)
// that needs to communicate how busy it is.
type Resource struct {
	Id           uint64    `json:"id"`
	FriendlyName string    `json:"friendlyName"`
	Status       Status    `json:"status"`
	Since        time.Time `json:"since"`
}

const resourceFmtString = "%s %v %016X %s\n"

// String will return a single-line representation of a resource.
// In order to optimize for standard streams, the output is as follows:
//   {{Since}}\t{{Status}}\t{{Id}}\t{{FriendlyName}}
// Formatted as follows:
//   2006-01-02T15:04:05Z07:00 1 0123456789ABCDEF My Resource
func (r Resource) String() string {
	return fmt.Sprintf(resourceFmtString, r.Since.Format(time.RFC3339), r.Status, r.Id, r.FriendlyName)
}

// MarshalJSON will return simple a simple json structure for a resource.
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
// according to the same format as MarshalJSON
func (r *Resource) UnmarshalJson(json []byte) error {
	return errors.New("Not Implemented")
}
