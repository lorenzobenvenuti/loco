package utils

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustCreate(path string, err error) string {
	if err != nil {
		panic(err)
	}
	return path
}

func TestTouchAndExists(t *testing.T) {
	dir := mustCreate(ioutil.TempDir("", ""))
	defer os.RemoveAll(dir)
	file := path.Join(dir, "foo")
	assert.False(t, Exists(file), "File must not exist")
	err := touch(file)
	assert.NoError(t, err)
	assert.True(t, Exists(file), "File must exist")
}

func TestCreateDirIfNotExists(t *testing.T) {
	temp := mustCreate(ioutil.TempDir("", ""))
	defer os.RemoveAll(temp)
	dir := path.Join(temp, "foo")
	assert.False(t, Exists(dir), "Directory must not exist")
	CreateDirIfNotExists(dir)
	assert.True(t, Exists(dir), "Directory must exist")
}

func TestMD5(t *testing.T) {
	assert.Equal(t, "acbd18db4cc2f85cedef654fccc4a4d8", MD5("foo"))
	assert.Equal(t, "37b51d194a7513e45b56f6524f2d51f2", MD5("bar"))
	assert.Equal(t, "e789bf22faae2374879bd5a6922fe558", MD5("/path/to/file.log"))
}

func TestWrapError(t *testing.T) {
	err := errors.New("First error")
	assert.Equal(t, "An error occurred: First error", Wrap(err, "An error occurred").Error())
}

func TestWrapfError(t *testing.T) {
	err := errors.New("First error")
	assert.Equal(t, "A foo error occurred: First error", Wrapf(err, "A %s error occurred", "foo").Error())
}
