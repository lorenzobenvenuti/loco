package state

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
		RotatedAt: time.Unix(0, int64(time.Hour)),
		Config: Config{
			Interval: time.Hour * 48,
		},
	}
	assert.True(t, s.FileMustBeRotated(time.Unix(0, int64(time.Hour*50))))
}

func TestFileMustNotBeRotated(t *testing.T) {
	s := &State{
		RotatedAt: time.Unix(0, int64(time.Hour)),
		Config: Config{
			Interval: time.Hour * 24,
		},
	}
	assert.False(t, s.FileMustBeRotated(time.Unix(0, int64(time.Hour*22))))
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
		Config:   Config{Interval: time.Hour * 24, Suffix: "%c"},
	}
	s2 := &State{
		FullName:  "/path/to/file2",
		CreatedAt: time.Date(2018, 11, 18, 17, 15, 12, 0, time.UTC),
		RotatedAt: time.Date(2018, 11, 18, 18, 15, 12, 0, time.UTC),
		Config:    Config{Interval: time.Hour * 24 * 14, Suffix: "%Y%m%d"},
	}
	var buf bytes.Buffer
	err := WriteStates(&buf, []*State{s1, s2})
	assert.NoError(t, err)
	assert.Equal(t, "FILE           CREATED AT          ROTATED AT          INTERVAL SUFFIX\n/path/to/file1 -                   -                   24h0m0s  %c\n/path/to/file2 18 Nov 18 17:15 UTC 18 Nov 18 18:15 UTC 336h0m0s %Y%m%d\n\n", buf.String())
}
