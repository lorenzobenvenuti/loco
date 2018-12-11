package defaults

import (
	"bytes"
	"errors"
	"os"
	"path"
	"testing"
	"time"

	"github.com/lorenzobenvenuti/loco/state"
	"github.com/lorenzobenvenuti/loco/utils"
	"github.com/stretchr/testify/assert"
)

func TestMapConfigReaderCanRetrievesAStringValueCorrectly(t *testing.T) {
	sut := &mapConfigReader{map[string]interface{}{"key": "value"}}
	v, err := sut.GetString("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestMapConfigReaderFailsRetrievingANonExistingStringValue(t *testing.T) {
	sut := &mapConfigReader{map[string]interface{}{}}
	_, err := sut.GetString("key")
	assert.Error(t, err)
}

func TestMapConfigReaderFailsRetrievingAStringValueIfItsNotAString(t *testing.T) {
	sut := &mapConfigReader{map[string]interface{}{"key": 42}}
	_, err := sut.GetString("key")
	assert.Error(t, err)
}

func TestMapConfigReaderCanRetrievesAnIntValueCorrectly(t *testing.T) {
	sut := &mapConfigReader{map[string]interface{}{"key": 42}}
	v, err := sut.GetInt("key")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), v)
}

func TestMapConfigReaderFailsRetrievingANonExistingIntValue(t *testing.T) {
	sut := &mapConfigReader{map[string]interface{}{}}
	_, err := sut.GetInt("key")
	assert.Error(t, err)
}

func TestMapConfigReaderFailsRetrievingAnIntValueIfItsNotAnInt(t *testing.T) {
	sut := &mapConfigReader{map[string]interface{}{"key": "value"}}
	_, err := sut.GetInt("key")
	assert.Error(t, err)
}

type testEnvReader struct {
	values map[string]string
}

func (r *testEnvReader) getEnv(key string) string {
	return r.values[key]
}

func TestEnvConfigReaderSuccessfullyRetrievesAStringValue(t *testing.T) {
	sut := &envConfigReader{
		keys:      map[string]string{"key": "ENV_KEY"},
		envReader: &testEnvReader{map[string]string{"ENV_KEY": "value"}},
	}
	v, err := sut.GetString("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestEnvConfigReaderGetStringFailsIfAKeyIsNotMapped(t *testing.T) {
	sut := &envConfigReader{
		keys:      map[string]string{},
		envReader: &testEnvReader{map[string]string{"ENV_KEY": "value"}},
	}
	_, err := sut.GetString("key")
	assert.Error(t, err)
}

func TestEnvConfigReaderGetStringFailsIfEnvVariableIsNotDefined(t *testing.T) {
	sut := &envConfigReader{
		keys:      map[string]string{"key": "ENV_KEY"},
		envReader: &testEnvReader{map[string]string{}},
	}
	_, err := sut.GetString("key")
	assert.Error(t, err)
}

func TestEnvConfigReaderSuccessfullyRetrievesAnIntValue(t *testing.T) {
	sut := &envConfigReader{
		keys:      map[string]string{"key": "ENV_KEY"},
		envReader: &testEnvReader{map[string]string{"ENV_KEY": "42"}},
	}
	v, err := sut.GetInt("key")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), v)
}

func TestEnvConfigReaderGetIntFailsIfAKeyIsNotMapped(t *testing.T) {
	sut := &envConfigReader{
		keys:      map[string]string{},
		envReader: &testEnvReader{map[string]string{"ENV_KEY": "value"}},
	}
	_, err := sut.GetInt("key")
	assert.Error(t, err)
}

func TestEnvConfigReaderGetIntFailsIfEnvVariableIsNotDefined(t *testing.T) {
	sut := &envConfigReader{
		keys:      map[string]string{"key": "ENV_KEY"},
		envReader: &testEnvReader{map[string]string{}},
	}
	_, err := sut.GetInt("key")
	assert.Error(t, err)
}

func newJsonFileConfigReader() *jsonFileConfigReader {
	dir := utils.MustCreateTempDir()
	return &jsonFileConfigReader{path.Join(dir, "defaults.json")}
}

func TestJsonFileConfigReaderGetStringReturnsAnErrorIfFileDoesNotExist(t *testing.T) {
	sut := newJsonFileConfigReader()
	defer os.RemoveAll(path.Dir(sut.path))
	_, err := sut.GetString("key")
	assert.Error(t, err)
}

func TestJsonFileConfigReaderSuccessfullyReturnsAString(t *testing.T) {
	sut := newJsonFileConfigReader()
	defer os.RemoveAll(path.Dir(sut.path))
	sut.writeDefaults(map[string]interface{}{"key": "value"})
	v, err := sut.GetString("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestJsonFileConfigReaderGetStringFailsIfKeyIsNotDefined(t *testing.T) {
	sut := newJsonFileConfigReader()
	defer os.RemoveAll(path.Dir(sut.path))
	sut.writeDefaults(map[string]interface{}{})
	_, err := sut.GetString("key")
	assert.Error(t, err)
}

func TestJsonFileConfigReaderGetStringFailsIfValueIsNotAString(t *testing.T) {
	sut := newJsonFileConfigReader()
	defer os.RemoveAll(path.Dir(sut.path))
	sut.writeDefaults(map[string]interface{}{"key": 42})
	_, err := sut.GetString("key")
	assert.Error(t, err)
}

func TestJsonFileConfigReaderGteIntReturnsAnErrorIfFileDoesNotExist(t *testing.T) {
	sut := newJsonFileConfigReader()
	defer os.RemoveAll(path.Dir(sut.path))
	_, err := sut.GetInt("key")
	assert.Error(t, err)
}

func TestJsonFileConfigReaderSuccessfullyReturnsAnInt(t *testing.T) {
	sut := newJsonFileConfigReader()
	defer os.RemoveAll(path.Dir(sut.path))
	sut.writeDefaults(map[string]interface{}{"key": 42})
	v, err := sut.GetInt("key")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), v)
}

func TestJsonFileConfigReaderGetIntFailsIfKeyIsNotDefined(t *testing.T) {
	sut := newJsonFileConfigReader()
	defer os.RemoveAll(path.Dir(sut.path))
	sut.writeDefaults(map[string]interface{}{})
	_, err := sut.GetInt("key")
	assert.Error(t, err)
}

func TestJsonFileConfigReaderGetStringFailsIfValueIsNotAnInt(t *testing.T) {
	sut := newJsonFileConfigReader()
	defer os.RemoveAll(path.Dir(sut.path))
	sut.writeDefaults(map[string]interface{}{"key": "value"})
	_, err := sut.GetInt("key")
	assert.Error(t, err)
}

type testConfigReader struct {
	stringValue string
	intValue    int64
	err         error
}

func (r *testConfigReader) GetString(key string) (string, error) {
	return r.stringValue, r.err
}

func (r *testConfigReader) GetInt(key string) (int64, error) {
	return r.intValue, r.err
}

func TestCompositeConfigReaderGetStringFailsIfAllReadersReturnAnError(t *testing.T) {
	sut := &compositeConfigReader{[]ConfigReader{
		&testConfigReader{err: errors.New("Error 1")},
		&testConfigReader{err: errors.New("Error 2")},
	}}
	_, err := sut.GetString("key")
	assert.Error(t, err)
}

func TestCompositeConfigReaderReturnsAStringIfFirstReaderReturnsAString(t *testing.T) {
	sut := &compositeConfigReader{[]ConfigReader{
		&testConfigReader{stringValue: "value"},
		&testConfigReader{err: errors.New("Error 2")},
	}}
	v, err := sut.GetString("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestCompositeConfigReaderReturnsAStringIfSecondReaderReturnsAString(t *testing.T) {
	sut := &compositeConfigReader{[]ConfigReader{
		&testConfigReader{err: errors.New("Error 2")},
		&testConfigReader{stringValue: "value"},
	}}
	v, err := sut.GetString("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestCompositeConfigReaderGetIntFailsIfAllReadersReturnAnError(t *testing.T) {
	sut := &compositeConfigReader{[]ConfigReader{
		&testConfigReader{err: errors.New("Error 1")},
		&testConfigReader{err: errors.New("Error 2")},
	}}
	_, err := sut.GetInt("key")
	assert.Error(t, err)
}

func TestCompositeConfigReaderReturnsAnIntIfFirstReaderReturnsAnInt(t *testing.T) {
	sut := &compositeConfigReader{[]ConfigReader{
		&testConfigReader{intValue: 42},
		&testConfigReader{err: errors.New("Error 2")},
	}}
	v, err := sut.GetInt("key")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), v)
}

func TestCompositeConfigReaderReturnsAnIntIfSecondReaderReturnsAnInt(t *testing.T) {
	sut := &compositeConfigReader{[]ConfigReader{
		&testConfigReader{err: errors.New("Error 2")},
		&testConfigReader{intValue: 42},
	}}
	v, err := sut.GetInt("key")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), v)
}

func TestMustGetStringOk(t *testing.T) {
	assert.Equal(t, "value", mustGetString(&testConfigReader{stringValue: "value"}, "key"))
}

func TestMustGetStringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "Panic was expected")
		}
	}()
	mustGetString(&testConfigReader{err: errors.New("Error")}, "key")
}

func TestMustGetIntOk(t *testing.T) {
	assert.Equal(t, int64(42), mustGetInt(&testConfigReader{intValue: 42}, "key"))
}

func TestMustGetIntPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "Panic was expected")
		}
	}()
	mustGetInt(&testConfigReader{err: errors.New("Error")}, "key")
}

func TestMergeWithDefaultConfig(t *testing.T) {
	d := state.NewConfig(time.Hour*2, "foo")
	assert.Equal(t, state.NewConfig(time.Hour*3, "bar"), mergeWithDefaultConfig(state.NewConfig(time.Hour*3, "bar"), d))
	assert.Equal(t, state.NewConfig(time.Hour*2, "bar"), mergeWithDefaultConfig(state.NewConfig(0, "bar"), d))
	assert.Equal(t, state.NewConfig(time.Hour*4, "foo"), mergeWithDefaultConfig(state.NewConfig(time.Hour*4, ""), d))
	assert.Equal(t, d, mergeWithDefaultConfig(state.NewConfig(0, ""), d))
}

func TestWriteDefaultConfig(t *testing.T) {
	var buf bytes.Buffer
	c := state.NewConfig(time.Hour*3, "foo")
	writeDefaultConfig(&buf, c)
	assert.Equal(t, "Interval: 3h0m0s\nSuffix:   foo\n", buf.String())
}
