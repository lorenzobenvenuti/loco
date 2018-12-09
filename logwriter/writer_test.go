package logwriter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/lorenzobenvenuti/loco/state"
	"github.com/lorenzobenvenuti/loco/utils"
	"github.com/stretchr/testify/assert"
)

type fakeNowProvider struct {
	now time.Time
}

func (p *fakeNowProvider) Now() time.Time {
	return p.now
}

func newFakeNowProvider(nanos int64) *fakeNowProvider {
	return &fakeNowProvider{now: time.Unix(0, nanos)}
}

type fakeFileNameGenerator struct {
}

func (g *fakeFileNameGenerator) FileName(state *state.State) string {
	return fmt.Sprintf("%s.bak", state.FullName)
}

func newFakeFileNameGenerator() *fakeFileNameGenerator {
	return &fakeFileNameGenerator{}
}

func TestLoadWriter(t *testing.T) {
	s := &state.State{
		FullName:  "/path/to/file",
		Interval:  time.Duration(100),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 34),
	}
	storage := state.NewMapStorage()
	storage.Store(s)
	lw, err := loadWriter(storage, newFakeNowProvider(42), newFakeFileNameGenerator(), "/path/to/file")
	assert.Nil(t, err, "Loading the writer should not return an error")
	expected := &state.State{
		FullName:  "/path/to/file",
		Interval:  time.Duration(100),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 34),
	}
	assert.Equal(t, expected, lw.state)
}

func TestNewWriter(t *testing.T) {
	storage := state.NewMapStorage()
	lw, err := newWriter(storage, newFakeNowProvider(42), newFakeFileNameGenerator(), "/path/to/file", time.Duration(123), "%c")
	assert.Nil(t, err, "Creating the writer should not return an error")
	s, err := storage.Load("/path/to/file")
	assert.Nil(t, err, "Retrieving the writer from storage should not return an error")
	assert.Equal(t, s, lw.state)
	assert.Nil(t, err, "Error should be nil")
	expected := &state.State{
		FullName: "/path/to/file",
		Interval: time.Duration(123),
		Suffix:   "%c",
	}
	assert.Equal(t, expected, lw.state)
}

func TestLogWriterFirstWrite(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	fullpath := path.Join(dir, "file.log")
	s := &state.State{
		FullName: fullpath,
		Interval: time.Duration(100),
	}
	storage := state.NewMapStorage()
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
	updated, err := storage.Load(fullpath)
	assert.NoError(t, err)
	expected := &state.State{
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
	storage := state.NewMapStorage()
	s := &state.State{
		FullName:  fullpath,
		Interval:  time.Duration(100),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 12),
	}
	storage.Store(s)
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
	updated, err := storage.Load(fullpath)
	assert.NoError(t, err)
	expected := &state.State{
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
	rotatedPath := path.Join(dir, "file.log.bak")
	err := ioutil.WriteFile(fullpath, []byte("bar"), 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(rotatedPath, []byte("This should be overwritten"), 0755)
	assert.NoError(t, err)
	storage := state.NewMapStorage()
	s := &state.State{
		FullName:  fullpath,
		Interval:  time.Duration(10),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 12),
	}
	storage.Store(s)
	lw := &LogWriter{
		state:             s,
		nowProvider:       newFakeNowProvider(42),
		stateStorage:      storage,
		fileNameGenerator: newFakeFileNameGenerator(),
	}
	_, err = lw.Write([]byte("foo"))
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fullpath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), bytes)
	bytes, err = ioutil.ReadFile(rotatedPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("bar"), bytes)
	updated, err := storage.Load(fullpath)
	assert.NoError(t, err)
	expected := &state.State{
		FullName:  fullpath,
		Interval:  time.Duration(10),
		CreatedAt: time.Unix(0, 12),
		RotatedAt: time.Unix(0, 42),
		Counter:   1,
	}
	assert.Equal(t, expected, updated)
}
