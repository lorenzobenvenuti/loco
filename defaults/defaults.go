package defaults

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/lorenzobenvenuti/loco/intervals"
	"github.com/lorenzobenvenuti/loco/utils"
)

type DefaultsProvider interface {
	Interval() (string, error)
}

type constDefaultsProvider struct {
	interval string
}

func (p *constDefaultsProvider) Interval() (string, error) {
	return p.interval, nil
}

type fileDefaultsProvider struct {
	path string
}

type defaults struct {
	Interval string
}

func (p *fileDefaultsProvider) Interval() (string, error) {
	d, err := p.loadDefaults()
	if err != nil {
		return "", err
	}
	return d.Interval, nil
}

func (p *fileDefaultsProvider) hasDefaults() bool {
	return utils.Exists(p.path)
}

func (p *fileDefaultsProvider) loadDefaults() (*defaults, error) {
	bytes, err := ioutil.ReadFile(p.path)
	if err != nil {
		return nil, err
	}
	d := &defaults{}
	err = json.Unmarshal(bytes, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (p *fileDefaultsProvider) writeDefaults(d *defaults) error {
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

func MustGetInterval(defaultsProvider DefaultsProvider) string {
	interval, err := defaultsProvider.Interval()
	if err != nil {
		panic(err)
	}
	return interval
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
	return &constDefaultsProvider{"1d"}
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
	d := &defaults{}
	d.Interval = MustGetInterval(p)
	return mustGetAppDirFileDefaultsProvider().writeDefaults(d)
}

func DefaultsToString(provider DefaultsProvider) string {
	interval := MustGetInterval(provider)
	return fmt.Sprintf("Interval: %s", interval)
}
