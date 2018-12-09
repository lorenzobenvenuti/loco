package defaults

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/lorenzobenvenuti/loco/state"
	"github.com/lorenzobenvenuti/loco/utils"
	"github.com/stretchr/testify/assert"
)

func TestConstDefaultsProvider(t *testing.T) {
	p := &constDefaultsProvider{interval: "1d", suffix: "%c"}
	interval, err := p.Interval()
	assert.NoError(t, err)
	assert.Equal(t, "1d", interval)
	suffix, err := p.Suffix()
	assert.NoError(t, err)
	assert.Equal(t, "%c", suffix)
}

func TestFileDefaultsProviderWriteAndLoad(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	fullpath := path.Join(dir, "defaults.json")
	p := &fileDefaultsProvider{fullpath}
	d := &state.Config{Interval: "2w", Suffix: "%Y"}
	assert.NoError(t, p.writeDefaults(d))
	interval, err := p.Interval()
	assert.NoError(t, err)
	assert.Equal(t, "2w", interval)
	suffix, err := p.Suffix()
	assert.NoError(t, err)
	assert.Equal(t, "%Y", suffix)
}

func TestFileDefaultsProviderReturnsAnErrorIfFileDoesNotExist(t *testing.T) {
	p := &fileDefaultsProvider{"/path/to/file"}
	_, err := p.loadDefaults()
	assert.Error(t, err)
}

func TestFileDefaultsProviderReturnsAnErrorIfFileCantBeUnmarshaled(t *testing.T) {
	dir := utils.MustCreateTempDir()
	defer os.RemoveAll(dir)
	fullpath := path.Join(dir, "defaults.json")
	ioutil.WriteFile(fullpath, []byte("Hello world"), 0755)
	p := &fileDefaultsProvider{fullpath}
	_, err := p.loadDefaults()
	assert.Error(t, err)
}

type testEnvVariableReader struct {
	value string
}

func (r *testEnvVariableReader) GetEnv(key string) string {
	return r.value
}

func TestEnvDefaultsProviderReturnsAValidValue(t *testing.T) {
	p := &envDefaultsProvider{&testEnvVariableReader{"1w"}}
	interval, err := p.Interval()
	assert.NoError(t, err)
	assert.Equal(t, "1w", interval)
}

func TestEnvDefaultsProviderFailsForInvalidValues(t *testing.T) {
	p := &envDefaultsProvider{&testEnvVariableReader{"2a"}}
	_, err := p.Interval()
	assert.Error(t, err)
}

type testDefaultsProvider struct {
	interval string
	suffix   string
}

func (p *testDefaultsProvider) Interval() (string, error) {
	return p.interval, nil
}

func (p *testDefaultsProvider) Suffix() (string, error) {
	return p.suffix, nil
}

type errDefaultsProvider struct {
	err error
}

func (p *errDefaultsProvider) Interval() (string, error) {
	return "", p.err
}

func (p *errDefaultsProvider) Suffix() (string, error) {
	return "", p.err
}

func TestCompositeDefaultsProviderDoesntFindAValidValue(t *testing.T) {
	p := &compositeDefaultsProvider{[]DefaultsProvider{
		&errDefaultsProvider{errors.New("Invalid interval")},
		&errDefaultsProvider{errors.New("Invalid interval")},
	}}
	_, err := p.Interval()
	assert.Error(t, err)
}

func TestCompositeDefaultsProviderWhenFirstProviderReturnsAValue(t *testing.T) {
	p := &compositeDefaultsProvider{[]DefaultsProvider{
		&testDefaultsProvider{interval: "2w", suffix: ""},
		&errDefaultsProvider{errors.New("Invalid interval")},
	}}
	interval, err := p.Interval()
	assert.NoError(t, err)
	assert.Equal(t, "2w", interval)
}

func TestCompositeDefaultsProviderWhenSecondProviderReturnsAValue(t *testing.T) {
	p := &compositeDefaultsProvider{[]DefaultsProvider{
		&errDefaultsProvider{errors.New("Invalid interval")},
		&testDefaultsProvider{interval: "3d", suffix: ""},
	}}
	interval, err := p.Interval()
	assert.NoError(t, err)
	assert.Equal(t, "3d", interval)
}
