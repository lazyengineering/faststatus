// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package resource

import (
	"time"
)

// A Resource represents any resource (a person, a bathroom, a server, etc.)
// that needs to communicate how busy it is.
type Resource struct {
	Id           uint
	FriendlyName string
	Status       Status
	Since        time.Time
}
