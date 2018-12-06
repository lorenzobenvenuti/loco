package intervals

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseReturnsAnError(t *testing.T) {
	values := []string{"ciao", "10s", "2z", "10hH", "xW", "1.2w"}
	for _, value := range values {
		_, err := Parse(value)
		assert.Error(t, err)
	}
}

func TestValidate(t *testing.T) {
	assert.NoError(t, Validate("2h"))
	assert.NoError(t, Validate("1d"))
	assert.NoError(t, Validate("2M"))
	assert.NoError(t, Validate("4w"))
	assert.Error(t, Validate("4H"))
	assert.Error(t, Validate("xyz"))
	assert.Error(t, Validate("3y"))
}

func TestParseOk(t *testing.T) {
	values := []string{"1h", "2M", "2d", "11w", "2m"}
	expected := []time.Duration{
		time.Duration(1000 * 1000 * 1000 * 60 * 60),
		time.Duration(1000 * 1000 * 1000 * 60 * 60 * 24 * 30 * 2),
		time.Duration(1000 * 1000 * 1000 * 60 * 60 * 24 * 2),
		time.Duration(1000 * 1000 * 1000 * 60 * 60 * 24 * 7 * 11),
		time.Duration(1000 * 1000 * 1000 * 60 * 2),
	}
	for i, value := range values {
		interval, err := Parse(value)
		assert.NoError(t, err)
		assert.Equal(t, expected[i], interval)
	}
}
