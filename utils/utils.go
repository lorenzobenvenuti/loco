package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

func touch(file string) error {
	f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func Exists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateDirIfNotExists(dir string) error {
	if !Exists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func MD5(value string) string {
	hasher := md5.New()
	hasher.Write([]byte(value))
	return hex.EncodeToString(hasher.Sum(nil))
}

type wrappedError struct {
	message string
	err     error
}

func (e *wrappedError) Error() string {
	return fmt.Sprintf("%s: %s", e.message, e.err.Error())
}

func Wrap(err error, msg string) error {
	return &wrappedError{
		message: msg,
		err:     err,
	}
}

func Wrapf(err error, msg string, a ...interface{}) error {
	return Wrap(err, fmt.Sprintf(msg, a...))
}

func MustCreateTempDir() string {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	return dir
}

func HomeDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return user.HomeDir, nil
}

func AppDir() (string, error) {
	dir, err := HomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(dir, ".loco"), err
}
