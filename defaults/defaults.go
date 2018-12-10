package defaults

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/lorenzobenvenuti/loco/intervals"
	"github.com/lorenzobenvenuti/loco/state"
	"github.com/lorenzobenvenuti/loco/utils"
)

type DefaultConfigProvider interface {
	DefaultConfig() *state.Config
}

type confStringReader interface {
	confString() (string, error);
}

type constConfStringReader struct {
	value string
}

func (r *constConfStringReader) confString() (string, error) {
	return r.value, nil
}

type envConfStringReader struct {
	key string
}

func (r *envConfStringReader) confString() (string, error) {
	return os.Getenv(r.key), nil
}

type jsonFileConfStringReader struct {
	path string
	key string
}

func (r *jsonFileConfStringReader) confString() (string, error) {
	m, err := r.loadDefaults()
	if err != nil {
		return "", err
	}
	return m[r.key].(string), nil
}

func (r *jsonFileConfStringReader) hasDefaults() bool {
	return utils.Exists(r.path)
}

func (r *jsonFileConfStringReader) loadDefaults() (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(r.path)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *jsonFileConfStringReader) writeDefaults(m map[string]interface{}) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(r.path, bytes, 0755)
}

type compositeConfStringReader struct {
	readers []confStringReader;
}

func (r *compositeConfStringReader) confString() (string, error) {
	for _, reader := range r.readers {
		s, err := reader.confString()
		if err == nil && s != "" {
			return s, nil
		}
	}
	return "", errors.New("Cannot retrieve configuration string")
}

type intervalConfStringReader struct {
	delegate confStringReader
}

func (r *intervalConfStringReader) confString() (string, error) {
	s, err := r.delegate.confString()
	if err != nil {
		return "", err
	}
	err = intervals.Validate(s)
	if err != nil {
		return "", err
	}
	return s, nil
}

func mustGetConfString(confString string, err error) string {
	if err != nil {
		panic(err)
	}
	return confString
}

type defaultConfigProvider struct {
	intervalConfStringReader confStringReader
	suffixConfStringReader confStringReader
}

func (p *defaultConfigProvider) DefaultConfig() *state.Config {
	return &state.Config{
		Interval: mustGetConfString(p.intervalConfStringReader.confString()),
		Suffix: mustGetConfString(p.suffixConfStringReader.confString()),
	}
}

func NewDefaultConfigProvider() DefaultConfigProvider {
	intervalConfStringReader := []confStringReader {
		&intervalConfStringReader{&envConfStringReader{"LOCO_INTERVAL"}},
		&jsonFileConfStringReader{"/path/to/file", "interval"},
		&constConfStringReader{"1d"},
	}
	suffixConfStringReader := []confStringReader {
		&envConfStringReader{"LOCO_SUFFIX"},
		&jsonFileConfStringReader{"/path/to/file", "suffix"},
		&constConfStringReader{"%c"},
	}
	return &defaultConfigProvider{
		intervalConfStringReader: &compositeConfStringReader{intervalConfStringReader},
		suffixConfStringReader: &compositeConfStringReader{suffixConfStringReader},
	}
}

type DefaultsProvider interface {
	Interval() (string, error)
	Suffix() (string, error)
}

type constDefaultsProvider struct {
	interval string
	suffix   string
}

func (p *constDefaultsProvider) Interval() (string, error) {
	return p.interval, nil
}

func (p *constDefaultsProvider) Suffix() (string, error) {
	return p.suffix, nil
}

type fileDefaultsProvider struct {
	path string
}

func (p *fileDefaultsProvider) Interval() (string, error) {
	d, err := p.loadDefaults()
	if err != nil {
		return "", err
	}
	return d.Interval, nil
}

func (p *fileDefaultsProvider) Suffix() (string, error) {
	d, err := p.loadDefaults()
	if err != nil {
		return "", err
	}
	return d.Suffix, nil
}

func (p *fileDefaultsProvider) hasDefaults() bool {
	return utils.Exists(p.path)
}

func (p *fileDefaultsProvider) loadDefaults() (*state.Config, error) {
	bytes, err := ioutil.ReadFile(p.path)
	if err != nil {
		return nil, err
	}
	d := &state.Config{}
	err = json.Unmarshal(bytes, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (p *fileDefaultsProvider) writeDefaults(d *state.Config) error {
	bytes, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p.path, bytes, 0755)
}

type envVariableReader interface {
	GetEnv(key string) string
}

type defaultEnvVariableReader struct{}

func (r *defaultEnvVariableReader) GetEnv(key string) string {
	return os.Getenv(key)
}

type envDefaultsProvider struct {
	envVariableReader envVariableReader
}

func (p *envDefaultsProvider) Interval() (string, error) {
	interval := p.envVariableReader.GetEnv("LOCO_INTERVAL")
	err := intervals.Validate(interval)
	if err != nil {
		return "", err
	}
	return interval, nil
}

func (p *envDefaultsProvider) Suffix() (string, error) {
	return p.envVariableReader.GetEnv("LOCO_SUFFIX"), nil
}

type compositeDefaultsProvider struct {
	providers []DefaultsProvider
}

func (p *compositeDefaultsProvider) Interval() (string, error) {
	for _, p := range p.providers {
		interval, err := p.Interval()
		if err == nil && interval != "" {
			return interval, nil
		}
	}
	return "", errors.New("Cannot find a default interval")
}

func (p *compositeDefaultsProvider) Suffix() (string, error) {
	for _, p := range p.providers {
		suffix, err := p.Suffix()
		if err == nil && suffix != "" {
			return suffix, nil
		}
	}
	return "", errors.New("Cannot find a default suffix")
}

func MustGetInterval(defaultsProvider DefaultsProvider) string {
	interval, err := defaultsProvider.Interval()
	if err != nil {
		panic(err)
	}
	return interval
}

func MustGetSuffix(defaultsProvider DefaultsProvider) string {
	suffix, err := defaultsProvider.Suffix()
	if err != nil {
		panic(err)
	}
	return suffix
}

func appDirFileDefaultsProvider() (*fileDefaultsProvider, error) {
	appDir, err := utils.AppDir()
	if err != nil {
		return nil, err
	}
	return &fileDefaultsProvider{path.Join(appDir, "defaults.json")}, nil
}

func mustGetAppDirFileDefaultsProvider() *fileDefaultsProvider {
	p, err := appDirFileDefaultsProvider()
	if err != nil {
		panic(err)
	}
	return p
}

func environmentDefaultsProvider() DefaultsProvider {
	return &envDefaultsProvider{&defaultEnvVariableReader{}}
}

func builtInDefaultsProvider() DefaultsProvider {
	return &constDefaultsProvider{interval: "1d", suffix: "%c"}
}

func NewStaticDefaultsProvider() DefaultsProvider {
	return &compositeDefaultsProvider{[]DefaultsProvider{
		mustGetAppDirFileDefaultsProvider(),
		builtInDefaultsProvider(),
	}}
}

func NewRuntimeDefaultsProvider() DefaultsProvider {
	return &compositeDefaultsProvider{[]DefaultsProvider{
		environmentDefaultsProvider(),
		NewStaticDefaultsProvider(),
	}}
}

func SetDefaultInterval(interval string) error {
	err := intervals.Validate(interval)
	if err != nil {
		return err
	}
	p := NewStaticDefaultsProvider()
	d := &state.Config{}
	d.Interval = interval
	d.Suffix = MustGetSuffix(p)
	return mustGetAppDirFileDefaultsProvider().writeDefaults(d)
}

func SetDefaultSuffix(suffix string) error {
	p := NewStaticDefaultsProvider()
	d := &state.Config{}
	d.Interval = MustGetInterval(p)
	d.Suffix = suffix
	return mustGetAppDirFileDefaultsProvider().writeDefaults(d)
}

func WriteDefaults(writer io.Writer, provider DefaultsProvider) {
	interval := MustGetInterval(provider)
	suffix := MustGetSuffix(provider)
	io.WriteString(writer, fmt.Sprintf("Interval\t%s\nSuffix\t%s\n", interval, suffix))
}
