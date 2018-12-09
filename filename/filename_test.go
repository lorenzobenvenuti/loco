package filename

import (
	"testing"
	"time"

	"github.com/lorenzobenvenuti/loco/logwriter"
	"github.com/stretchr/testify/assert"
)

func TestSuffixWithExtension(t *testing.T) {
	sut := &suffixFileNameGenerator{}
	state := &logwriter.State{
		FullName: "/path/to/file.log",
		Counter:  1,
		Suffix:   "foo",
	}
	assert.Equal(t, "/path/to/file.foo.log", sut.FileName(state))
}

func TestSuffixWithoutExtension(t *testing.T) {
	sut := &suffixFileNameGenerator{}
	state := &logwriter.State{
		FullName: "/path/to/file",
		Counter:  1,
		Suffix:   "bar",
	}
	assert.Equal(t, "/path/to/file.bar", sut.FileName(state))
}

func TestSuffixWithCounter(t *testing.T) {
	state := &logwriter.State{
		Counter: 1,
		Suffix:  "X%cY",
	}
	assert.Equal(t, "X1Y", suffixFromState(state))
}

func TestSuffixWithTrailingCounter(t *testing.T) {
	state := &logwriter.State{
		Counter: 1,
		Suffix:  "%cX",
	}
	assert.Equal(t, "1X", suffixFromState(state))
}

func TestEscapedPercent(t *testing.T) {
	state := &logwriter.State{
		Counter: 1,
		Suffix:  "c%%%cc",
	}
	assert.Equal(t, "c%1c", suffixFromState(state))
}

func TestSuffixWithDate(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2018-12-09")
	state := &logwriter.State{
		Counter:   1,
		Suffix:    "X%Y%m%dY",
		RotatedAt: now,
	}
	assert.Equal(t, "X20181209Y", suffixFromState(state))
}
