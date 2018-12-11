package state

import (
	"os"
	"testing"
	"time"

	"github.com/lorenzobenvenuti/loco/utils"
	"github.com/stretchr/testify/assert"
)

func mustCreateStorage(dir string) *fileStateStorage {
	s, err := newFileStateStorage(dir)
	if err != nil {
		panic(err)
	}
	return s
}

func TestStoreReturnsAnErrorWhenStateIsNotFound(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	s := mustCreateStorage(dir)
	_, err := s.Load("/path/to/file")
	assert.NotNil(t, err, "Should raise an error if session is not found")
}

func TestStoreSuccessfullyStoresAState(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	s := mustCreateStorage(dir)
	expected := &State{
		FullName:  "/path/to/file",
		CreatedAt: time.Unix(0, 10000),
		RotatedAt: time.Unix(0, 20000),
		Counter:   42,
		Config: Config{
			Interval: time.Hour * 24 * 21,
			Suffix:   "%Y%m%d",
		},
	}
	s.Store(expected)
	actual, err := s.Load("/path/to/file")
	assert.Nil(t, err, "No error is returned when state is found")
	assert.Equal(t, expected, actual, "Loaded state is equal to the stored")
}

func TestStoreSuccessfullyListStates(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	s := mustCreateStorage(dir)
	s1 := &State{
		FullName:  "/path/to/file",
		CreatedAt: time.Unix(0, 10000),
		RotatedAt: time.Unix(0, 20000),
		Counter:   42,
		Config: Config{
			Interval: time.Hour * 48,
			Suffix:   "%c",
		},
	}
	s2 := &State{
		FullName:  "/path/to/another/file",
		CreatedAt: time.Unix(0, 30000),
		RotatedAt: time.Unix(0, 40000),
		Counter:   77,
		Config: Config{
			Interval: time.Hour * 48,
			Suffix:   "%c",
		},
	}
	s.Store(s1)
	s.Store(s2)
	states, err := s.List()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(states))
}

func TestStoreSuccessfullyRemoveStates(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	s := mustCreateStorage(dir)
	s1 := &State{
		FullName:  "/path/to/file",
		CreatedAt: time.Unix(0, 10000),
		RotatedAt: time.Unix(0, 20000),
		Counter:   42,
		Config: Config{
			Interval: time.Hour * 48,
			Suffix:   "%c",
		},
	}
	s.Store(s1)
	states, err := s.List()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(states))
	err = s.Remove("/path/to/file")
	assert.NoError(t, err)
	states, err = s.List()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(states))
}
