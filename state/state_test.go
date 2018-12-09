package state

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/lorenzobenvenuti/loco/utils"
	"github.com/stretchr/testify/assert"
)

func TestJsonStateMarshaler(t *testing.T) {
	s := &State{
		FullName:  "/path/to/file.log",
		CreatedAt: time.Unix(0, 1000000),
		RotatedAt: time.Unix(0, 2000000),
		Interval:  time.Duration(1000),
		Counter:   42,
	}
	m := &jsonStateMarshaler{}
	v, err := m.marshal(s)
	assert.Nil(t, err, "Should marshal without errors")
	actual, err := m.unmarshal(v)
	assert.Nil(t, err, "Should unmarshal without errors")
	assert.Equal(t, s, actual, "Expected %v, actual %v", s, actual)
}

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
		Interval:  time.Duration(100),
		Counter:   42,
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
		Interval:  time.Duration(100),
		Counter:   42,
	}
	s2 := &State{
		FullName:  "/path/to/another/file",
		CreatedAt: time.Unix(0, 30000),
		RotatedAt: time.Unix(0, 40000),
		Interval:  time.Duration(200),
		Counter:   77,
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
		Interval:  time.Duration(100),
		Counter:   42,
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

func TestFileMustBeCreated(t *testing.T) {
	s := &State{}
	assert.True(t, s.FileMustBeCreated())
}

func TestFileMustNotBeCreated(t *testing.T) {
	s := &State{CreatedAt: time.Unix(0, 42)}
	assert.False(t, s.FileMustBeCreated())
}

func TestFileMustBeRotated(t *testing.T) {
	s := &State{
		RotatedAt: time.Unix(0, 30),
		Interval:  time.Duration(10),
	}
	assert.True(t, s.FileMustBeRotated(time.Unix(0, 42)))
}

func TestFileMustNotBeRotated(t *testing.T) {
	s := &State{
		RotatedAt: time.Unix(0, 40),
		Interval:  time.Duration(10),
	}
	assert.False(t, s.FileMustBeRotated(time.Unix(0, 42)))
}

func TestPrettyCreatedAt(t *testing.T) {
	s := &State{
		CreatedAt: time.Date(2018, 11, 18, 17, 15, 0, 0, time.UTC),
	}
	assert.Equal(t, "18 Nov 18 17:15 UTC", s.PrettyCreatedAt())
}

func TestEmptyPrettyCreatedAt(t *testing.T) {
	s := &State{}
	assert.Equal(t, "-", s.PrettyCreatedAt())
}

func TestPrettyRotatedAt(t *testing.T) {
	s := &State{
		RotatedAt: time.Date(2018, 11, 18, 17, 15, 0, 0, time.UTC),
	}
	assert.Equal(t, "18 Nov 18 17:15 UTC", s.PrettyRotatedAt())
}

func TestEmptyPrettyRotatedAt(t *testing.T) {
	s := &State{}
	assert.Equal(t, "-", s.PrettyRotatedAt())
}

func TestWriteStates(t *testing.T) {
	s1 := &State{
		FullName: "/path/to/file1",
	}
	s2 := &State{
		FullName:  "/path/to/file2",
		CreatedAt: time.Date(2018, 11, 18, 17, 15, 12, 0, time.UTC),
		RotatedAt: time.Date(2018, 11, 18, 18, 15, 12, 0, time.UTC),
	}
	var buf bytes.Buffer
	err := WriteStates(&buf, []*State{s1, s2})
	assert.NoError(t, err)
	assert.Equal(t, "FILE           CREATED AT          ROTATED AT\n/path/to/file1 -                   -\n/path/to/file2 18 Nov 18 17:15 UTC 18 Nov 18 18:15 UTC\n\n", buf.String())
}

func TestNewConfig(t *testing.T) {
	storage := NewMapStorage()
	s, err := NewConfig(storage, "/path/to/file", time.Duration(123))
	assert.NoError(t, err, "Creating the config should not return an error")
	loaded, err := storage.Load("/path/to/file")
	assert.NoError(t, err, "Retrieving the writer from storage should not return an error")
	assert.Equal(t, s, loaded)
	assert.Nil(t, err, "Error should be nil")
	expected := &State{
		FullName: "/path/to/file",
		Interval: time.Duration(123),
	}
	assert.Equal(t, expected, s)
}
