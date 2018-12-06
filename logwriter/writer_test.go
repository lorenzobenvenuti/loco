package logwriter

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/lorenzobenvenuti/loco/utils"
	"github.com/stretchr/testify/assert"
)

type fakeStorage struct {
	states map[string]*State
}

func (s *fakeStorage) store(state *State) error {
	s.states[state.FullName] = state
	return nil
}

func (s *fakeStorage) load(fullName string) (*State, error) {
	return s.states[fullName], nil
}

func (s *fakeStorage) list() ([]*State, error) {
	return nil, nil
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{
		states: make(map[string]*State),
	}
}

type fakeNowProvider struct {
	now time.Time
}

func (p *fakeNowProvider) Now() time.Time {
	return p.now
}

func newFakeNowProvider(nanos int64) *fakeNowProvider {
	return &fakeNowProvider{now: time.Unix(0, nanos)}
}

func TestLoadWriter(t *testing.T) {
	s := &State{
		FullName:  "/path/to/file",
		Interval:  time.Duration(100),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 34),
	}
	storage := newFakeStorage()
	storage.store(s)
	lw, err := loadWriter(storage, newFakeNowProvider(42), "/path/to/file")
	assert.Nil(t, err, "Loading the writer should not return an error")
	expected := &State{
		FullName:  "/path/to/file",
		Interval:  time.Duration(100),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 34),
	}
	assert.Equal(t, expected, lw.state)
}

func TestNewWriter(t *testing.T) {
	storage := newFakeStorage()
	lw, err := newWriter(storage, newFakeNowProvider(42), "/path/to/file", time.Duration(123))
	assert.Nil(t, err, "Creating the writer should not return an error")
	s, err := storage.load("/path/to/file")
	assert.Nil(t, err, "Retrieving the writer from storage should not return an error")
	assert.Equal(t, s, lw.state)
	assert.Nil(t, err, "Error should be nil")
	expected := &State{
		FullName: "/path/to/file",
		Interval: time.Duration(123),
	}
	assert.Equal(t, expected, lw.state)
}

func TestNewConfig(t *testing.T) {
	storage := newFakeStorage()
	s, err := newConfig(storage, "/path/to/file", time.Duration(123))
	assert.NoError(t, err, "Creating the config should not return an error")
	loaded, err := storage.load("/path/to/file")
	assert.NoError(t, err, "Retrieving the writer from storage should not return an error")
	assert.Equal(t, s, loaded)
	assert.Nil(t, err, "Error should be nil")
	expected := &State{
		FullName: "/path/to/file",
		Interval: time.Duration(123),
	}
	assert.Equal(t, expected, s)
}

func TestLogWriterFirstWrite(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	fullpath := path.Join(dir, "file.log")
	s := &State{
		FullName: fullpath,
		Interval: time.Duration(100),
	}
	storage := newFakeStorage()
	lw := &LogWriter{
		state:        s,
		nowProvider:  newFakeNowProvider(42),
		stateStorage: storage,
	}
	_, err := lw.Write([]byte("foo"))
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fullpath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), bytes)
	updated, err := storage.load(fullpath)
	assert.NoError(t, err)
	expected := &State{
		FullName:  fullpath,
		Interval:  time.Duration(100),
		CreatedAt: time.Unix(0, 42),
		RotatedAt: time.Unix(0, 42),
		Counter:   0,
	}
	assert.Equal(t, expected, updated)
	assert.False(t, utils.Exists(path.Join(dir, "file.log.0")))
}

func TestLogWriterFileExists(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	fullpath := path.Join(dir, "file.log")
	ioutil.WriteFile(fullpath, []byte("bar"), 0755)
	storage := newFakeStorage()
	s := &State{
		FullName:  fullpath,
		Interval:  time.Duration(100),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 12),
	}
	storage.store(s)
	lw := &LogWriter{
		state:        s,
		nowProvider:  newFakeNowProvider(42),
		stateStorage: storage,
	}
	_, err := lw.Write([]byte("foo"))
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fullpath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("barfoo"), bytes)
	updated, err := storage.load(fullpath)
	assert.NoError(t, err)
	expected := &State{
		FullName:  fullpath,
		Interval:  time.Duration(100),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 12),
		Counter:   0,
	}
	assert.Equal(t, expected, updated)
	assert.False(t, utils.Exists(path.Join(dir, "file.log.0")))
}

func TestLogWriterFileRotation(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	fullpath := path.Join(dir, "file.log")
	rotatedPath := path.Join(dir, "file.log.0")
	err := ioutil.WriteFile(fullpath, []byte("bar"), 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(rotatedPath, []byte("This should be overwritten"), 0755)
	assert.NoError(t, err)
	storage := newFakeStorage()
	s := &State{
		FullName:  fullpath,
		Interval:  time.Duration(10),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 12),
	}
	storage.store(s)
	lw := &LogWriter{
		state:        s,
		nowProvider:  newFakeNowProvider(42),
		stateStorage: storage,
	}
	_, err = lw.Write([]byte("foo"))
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fullpath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), bytes)
	bytes, err = ioutil.ReadFile(rotatedPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("bar"), bytes)
	updated, err := storage.load(fullpath)
	assert.NoError(t, err)
	expected := &State{
		FullName:  fullpath,
		Interval:  time.Duration(10),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 42),
		Counter:   1,
	}
	assert.Equal(t, expected, updated)
}
