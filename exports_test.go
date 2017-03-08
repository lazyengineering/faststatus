// Copyright 2016-2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package faststatus

import (
	"math/rand"
	"reflect"
	"time"
	"unicode/utf8"
)

var availableLocations []*time.Location

func init() {
	availableLocations = []*time.Location{
		mustLocation(time.LoadLocation("Europe/London")),
		mustLocation(time.LoadLocation("America/New_York")),
		mustLocation(time.LoadLocation("America/Los_Angeles")),
		mustLocation(time.LoadLocation("Australia/Sydney")),
		mustLocation(time.LoadLocation("Asia/Tokyo")),
		mustLocation(time.LoadLocation("Asia/Shanghai")),
		mustLocation(time.LoadLocation("Asia/Kolkata")),
		mustLocation(time.LoadLocation("Europe/Istanbul")),
		mustLocation(time.LoadLocation("Europe/Zurich")),
		time.UTC,
	}
}

func mustLocation(loc *time.Location, err error) *time.Location {
	if err != nil {
		panic(err)
	}
	return loc
}

// Generate is used in testing to generate random valid Resource values
func (r Resource) Generate(rgen *rand.Rand, size int) reflect.Value {
	rr := Resource{}

	rr.ID, _ = NewID()
	rr.FriendlyName = func(rgen *rand.Rand, size int) string {
		txt := make([]byte, 0, size)
		for len(txt) < size {
			p := make([]byte, 1)
			n, err := rgen.Read(p)
			if err != nil {
				panic(err)
			}
			if n != 1 {
				continue
			}
			if utf8.Valid(p) {
				txt = append(txt, p...)
			}
		}
		return string(txt)
	}(rgen, rgen.Intn(100))
	rr.Status = Status(rgen.Int() % int(Occupied))
	rr.Since = time.Date(
		2016+rgen.Intn(10),
		time.Month(rgen.Intn(11)+1),
		rgen.Intn(27)+1,
		rgen.Intn(24),
		rgen.Intn(60),
		rgen.Intn(60),
		0,
		availableLocations[rgen.Int()%len(availableLocations)],
	)

	return reflect.ValueOf(rr)
}

// Generate is used in testing to generate only valid Status values
func (s Status) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(Status(rand.Int() % int(Occupied)))
}

const BinaryVersion = binaryVersion
