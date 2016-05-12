// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

// Status represents how busy a given resource is on a scale from 0–2,
// where 0 (Free) is a completely unoccupied resource, 2 (Occupied) is
// completely occupied, and 1 (Busy) is anything between. The simplicity
// and flexibility of this scheme allows this to be used for any number
// of applications.
type Status uint

const (
	Free     Status = iota // completely free resource
	Busy                   // resource is busy
	Occupied               // resource completely busy
)
