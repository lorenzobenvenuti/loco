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
