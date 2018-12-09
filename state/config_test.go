package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	storage := NewMapStorage()
	c := NewConfig("1d", "%c")
	s, err := NewState(storage, "/path/to/file", *c)
	assert.NoError(t, err, "Creating the config should not return an error")
	loaded, err := storage.Load("/path/to/file")
	assert.NoError(t, err, "Retrieving the writer from storage should not return an error")
	assert.Equal(t, s, loaded)
	assert.Nil(t, err, "Error should be nil")
	expected := &State{
		FullName: "/path/to/file",
		Config:   *c,
	}
	assert.Equal(t, expected, s)
}
