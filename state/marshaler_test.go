package state

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJsonStateMarshaler(t *testing.T) {
	s := &State{
		FullName:  "/path/to/file.log",
		CreatedAt: time.Unix(0, 1000000),
		RotatedAt: time.Unix(0, 2000000),
		Counter:   42,
		Config: Config{
			Interval: time.Hour * 48,
			Suffix:   "%c",
		},
	}
	m := &jsonStateMarshaler{}
	v, err := m.marshal(s)
	assert.Nil(t, err, "Should marshal without errors")
	actual, err := m.unmarshal(v)
	assert.Nil(t, err, "Should unmarshal without errors")
	assert.Equal(t, s, actual, "Expected %v, actual %v", s, actual)
}
