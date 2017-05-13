package store

import "strings"

type dataError struct {
	old  bool
	noID bool
}

func (e dataError) Error() string {
	reasons := []string{"bad data"}
	if e.old {
		reasons = append(reasons, "a more recent version of this resource already exists")
	}
	if e.noID {
		reasons = append(reasons, "resource ID cannot be zero-value")
	}
	return strings.Join(reasons, ", ")
}

func (e dataError) Stale() bool {
	return e.old
}

func (e dataError) ZeroValue() bool {
	return e.noID
}

func StaleError(e error) bool {
	type staler interface {
		Stale() bool
	}
	if e, ok := e.(staler); ok {
		return e.Stale()
	}
	return false
}

func ZeroValueError(e error) bool {
	type zeroValuer interface {
		ZeroValue() bool
	}
	if e, ok := e.(zeroValuer); ok {
		return e.ZeroValue()
	}
	return false
}
