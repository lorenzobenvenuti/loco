package filename

import (
	"testing"
	"time"

	"github.com/lorenzobenvenuti/loco/state"
	"github.com/stretchr/testify/assert"
)

func TestSuffixWithExtension(t *testing.T) {
	sut := NewFileNameGenerator()
	state := &state.State{
		FullName: "/path/to/file.log",
		Counter:  1,
		Config:   state.Config{Suffix: "foo"},
	}
	assert.Equal(t, "/path/to/file.foo.log", sut.FileName(state))
}

func TestSuffixWithoutExtension(t *testing.T) {
	sut := NewFileNameGenerator()
	state := &state.State{
		FullName: "/path/to/file",
		Counter:  1,
		Config:   state.Config{Suffix: "bar"},
	}
	assert.Equal(t, "/path/to/file.bar", sut.FileName(state))
}

func TestSuffixWithCounter(t *testing.T) {
	state := &state.State{
		FullName: "/path/to/file.log",
		Counter:  1,
		Config:   state.Config{Suffix: "X%cY"},
	}
	sut := NewFileNameGenerator()
	assert.Equal(t, "/path/to/file.X1Y.log", sut.FileName(state))
}

func TestSuffixWithTrailingCounter(t *testing.T) {
	state := &state.State{
		FullName: "/path/to/file.log",
		Counter:  1,
		Config:   state.Config{Suffix: "%cX"},
	}
	sut := NewFileNameGenerator()
	assert.Equal(t, "/path/to/file.1X.log", sut.FileName(state))
}

func TestEscapedPercent(t *testing.T) {
	state := &state.State{
		FullName: "/path/to/file.log",
		Counter:  1,
		Config:   state.Config{Suffix: "c%%%cc"},
	}
	sut := NewFileNameGenerator()
	assert.Equal(t, "/path/to/file.c%1c.log", sut.FileName(state))
}

func TestTwoEscapedPercent(t *testing.T) {
	state := &state.State{
		FullName: "/path/to/file.log",
		Counter:  1,
		Config:   state.Config{Suffix: "c%%%%%cc"},
	}
	sut := NewFileNameGenerator()
	assert.Equal(t, "/path/to/file.c%%1c.log", sut.FileName(state))
}

func TestSuffixWithDate(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2018-12-09")
	state := &state.State{
		FullName:  "/path/to/file.log",
		Counter:   1,
		Config:    state.Config{Suffix: "X%Y%m%dY"},
		RotatedAt: now,
	}
	sut := NewFileNameGenerator()
	assert.Equal(t, "/path/to/file.X20181209Y.log", sut.FileName(state))
}

func TestSuffixWithDateAndTime(t *testing.T) {
	now, _ := time.Parse("2006-01-02 15:04:05", "2018-12-09 15:21:32")
	state := &state.State{
		FullName:  "/path/to/file.log",
		Counter:   1,
		Config:    state.Config{Suffix: "X%Y%m%d%H%M%SY"},
		RotatedAt: now,
	}
	sut := NewFileNameGenerator()
	assert.Equal(t, "/path/to/file.X20181209152132Y.log", sut.FileName(state))
}
