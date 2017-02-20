// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package faststatus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// A Resource represents any resource (a person, a bathroom, a server, etc.)
// that needs to communicate how busy it is.
type Resource struct {
	ID           ID
	Status       Status
	Since        time.Time
	FriendlyName string
}

// String will return a single-line representation of a valid resource.
// In order to optimize for standard streams, the output is as follows:
//   {{ID}} {{Status}} {{Since}} {{FriendlyName}}
// Formatted as follows:
//   01234567-89ab-cdef-0123-456789abcdef busy 2006-01-02T15:04:05Z07:00 My Resource
func (r Resource) String() string {
	txt, err := r.MarshalText()
	if err != nil {
		return ""
	}
	return string(txt)
}

// MarshalText encodes a Resource to the text representation. In order to
// better stream text, the output is as follows:
//   {{ID}} {{Status}} {{Since}} {{FriendlyName}}
// Formatted as follows:
//   01234567-89ab-cdef-0123-456789abcdef busy 2006-01-02T15:04:05Z07:00 My Resource
// An invalid Status (out of range, etc.) will result in an error.
func (r Resource) MarshalText() ([]byte, error) {
	txt := make([]byte, 0, 128)

	id, err := r.ID.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("marshaling text for ID: %+v", err)
	}
	txt = append(txt, id...)

	txt = append(txt, ' ')
	status, err := r.Status.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("marshaling Status to text: %+v", err)
	}
	txt = append(txt, status...)

	txt = append(txt, ' ')
	since, err := r.Since.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("marshaling Since to text: %+v", err)
	}
	txt = append(txt, since...)

	if r.FriendlyName != "" {
		txt = append(txt, ' ')
		txt = append(txt, r.FriendlyName...)
	}

	return txt, nil
}

// UnmarshalText decodes a Resource from a line of text. This matches the
// output of the `MarshalText` method. Partial matches are only accepted missing
// `FriendlyName` or `FriendlyName` and `Since`.
func (r *Resource) UnmarshalText(txt []byte) error {
	elements := bytes.Split(txt, []byte(" "))

	if len(elements) < 3 {
		return fmt.Errorf("invalid resource text")
	}

	tmp := Resource{}

	if err := (&tmp.ID).UnmarshalText(elements[0]); err != nil {
		return fmt.Errorf("parsing ID from text: %+v", err)
	}

	if err := (&tmp.Status).UnmarshalText(elements[1]); err != nil {
		return fmt.Errorf("parsing Status from text: %+v", err)
	}

	if err := (&tmp.Since).UnmarshalText(elements[2]); err != nil {
		return fmt.Errorf("parsing Since from text: %+v", err)
	}

	tmp.FriendlyName = string(bytes.Join(elements[3:], []byte(" ")))

	*r = tmp

	return nil
}

// MarshalJSON will return simple a simple json structure for a resource.
// Will not accept any Status that is out of range; see Status documentation
// for more information.
func (r Resource) MarshalJSON() ([]byte, error) {
	tmpResource := struct {
		ID           ID        `json:"id"`
		Status       Status    `json:"status"`
		Since        time.Time `json:"since"`
		FriendlyName string    `json:"friendlyName"`
	}{
		r.ID,
		r.Status,
		r.Since,
		r.FriendlyName,
	}
	return json.Marshal(tmpResource)
}

// UnmarshalJSON will populate a Resource with data from a json struct
// according to the same format as MarshalJSON. Will overwrite any values
// already assigned to the Resource.
func (r *Resource) UnmarshalJSON(raw []byte) error {
	// allow zero values with omitempty
	tmp := new(struct {
		ID           ID        `json:",omitempty"`
		Status       Status    `json:",omitempty"`
		Since        time.Time `json:",omitempty"`
		FriendlyName string    `json:",omitempty"`
	})
	if err := json.Unmarshal(raw, tmp); err != nil {
		return err
	}

	r.ID = tmp.ID
	r.FriendlyName = tmp.FriendlyName
	r.Status = tmp.Status
	r.Since = tmp.Since
	if r.Since.IsZero() {
		r.Since = time.Time{}
	}
	return nil
}
