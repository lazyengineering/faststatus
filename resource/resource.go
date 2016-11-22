// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"bytes"
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

// String will return a single-line representation of a valid resource.
// In order to optimize for standard streams, the output is as follows:
//   {{Id}} {{Status}} {{Since}} {{FriendlyName}}
// Formatted as follows:
//   0123456789ABCDEF busy 2006-01-02T15:04:05Z07:00 My Resource
func (r Resource) String() string {
	txt, err := r.MarshalText()
	if err != nil {
		return ""
	}
	return string(txt)
}

// MarshalText encodes a Resource to the text representation. In order to
// better stream text, the output is as follows:
//   {{Id}} {{Status}} {{Since}} {{FriendlyName}}
// Formatted as follows:
//   0123456789ABCDEF busy 2006-01-02T15:04:05Z07:00 My Resource
// An invalid Status (out of range, etc.) will result in an error.
func (r Resource) MarshalText() ([]byte, error) {
	var b bytes.Buffer

	{ // ID
		idBuf := bytes.Repeat([]byte("0"), 16)
		idStr := strconv.FormatUint(r.Id, 16)
		idBuf = append(idBuf[:16-len(idStr)], idStr...)
		if _, err := b.Write(idBuf); err != nil {
			return nil, fmt.Errorf("writing ID to resource text: %+v", err)
		}
		if _, err := b.WriteString(" "); err != nil {
			return nil, fmt.Errorf("writing space to resource text: %+v", err)
		}
	}

	{ // Status
		txt, err := r.Status.MarshalText()
		if err != nil {
			return nil, fmt.Errorf("marshaling Status to text: %+v", err)
		}
		if _, err := b.Write(txt); err != nil {
			return nil, fmt.Errorf("writing Status to resource text: %+v", err)
		}
		if _, err := b.WriteString(" "); err != nil {
			return nil, fmt.Errorf("writing space to resource text: %+v", err)
		}
	}

	{ // Since
		txt, err := r.Since.MarshalText()
		if err != nil {
			return nil, fmt.Errorf("marshaling Since to text: %+v", err)
		}
		if _, err := b.Write(txt); err != nil {
			return nil, fmt.Errorf("writing Since to resource text: %+v", err)
		}
		if _, err := b.WriteString(" "); err != nil {
			return nil, fmt.Errorf("writing space to resource text: %+v", err)
		}
	}

	{ // Friendly Name
		if _, err := b.WriteString(r.FriendlyName); err != nil {
			return nil, fmt.Errorf("writing FriendlyName to resource text: %+v", err)
		}
	}

	return b.Bytes(), nil
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
	if r.Since.IsZero() {
		r.Since = time.Time{}
	}
	return nil
}
