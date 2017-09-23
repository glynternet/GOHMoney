package GOHMoney

import (
	"time"

	"github.com/lib/pq"
)

// NullTime is a wrapper to a pq.NullTime object to extend functionality by adding more methods
type NullTime pq.NullTime

// Equal returns true if the two NullTime objects are exactly the same.
// Equal even evaluates the Time fields of both NullTime objects if they are both not Valid
func (a NullTime) Equal(b NullTime) bool {
	if a.Valid != b.Valid || !a.Time.Equal(b.Time) {
		return false
	}
	return true
}

// EqualTime return true a NullTime represents a given time.
// EqualTime will always return false if the NullTime is not Valid.
func (nt NullTime) EqualTime(t time.Time) bool {
	if !nt.Valid {
		return false
	}
	if !nt.Time.Equal(t) {
		return false
	}
	return true
}
