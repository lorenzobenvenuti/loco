package defaults

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/lorenzobenvenuti/loco/state"
	"github.com/lorenzobenvenuti/loco/utils"
)

const INTERVAL = "interval"
const SUFFIX = "suffix"

type ConfigReader interface {
	GetString(key string) (string, error)
	GetInt(key string) (int64, error)
}

type mapConfigReader struct {
	config map[string]interface{}
}

func (r *mapConfigReader) GetString(key string) (string, error) {
	if v, ok := r.config[key]; ok {
		s, ok := v.(string)
		if ok {
			return s, nil
		}
		return "", fmt.Errorf("Cannot convert %v to string", v)
	}
	return "", fmt.Errorf("Cannot find key %s", key)
}

func toInt(v interface{}) (int64, error) {
	switch v.(type) {
	case int:
		return int64(v.(int)), nil
	case int32:
		return int64(v.(int32)), nil
	case int64:
		return v.(int64), nil
	case float64:
		return int64(v.(float64)), nil
	}
	return 0, fmt.Errorf("Cannot convert %v to int", v)
}

func (r *mapConfigReader) GetInt(key string) (int64, error) {
	if v, ok := r.config[key]; ok {
		return toInt(v)
	}
	return 0, fmt.Errorf("Cannot find key %s", key)
}

type envReader interface {
	getEnv(key string) string
}

type defaultEnvReader struct{}

func (r *defaultEnvReader) getEnv(key string) string {
	return os.Getenv(key)
}

type envConfigReader struct {
	keys      map[string]string
	envReader envReader
}

func (r *envConfigReader) GetString(key string) (string, error) {
	if envKey, ok := r.keys[key]; ok {
		v := r.envReader.getEnv(envKey)
		if v == "" {
			return "", fmt.Errorf("Variable %s is not set", envKey)
		}
		return r.envReader.getEnv(envKey), nil
	}
	return "", fmt.Errorf("Cannot find key %s", key)
}

func (r *envConfigReader) GetInt(key string) (int64, error) {
	if envKey, ok := r.keys[key]; ok {
		v := r.envReader.getEnv(envKey)
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("Cannot convert %s to int", v)
		}
		return i, nil
	}
	return 0, fmt.Errorf("Cannot find key %s", key)
}

type jsonFileConfigReader struct {
	path string
}

func (r *jsonFileConfigReader) GetString(key string) (string, error) {
	m, err := r.loadDefaults()
	if err != nil {
		return "", err
	}
	return (&mapConfigReader{m}).GetString(key)
}

func (r *jsonFileConfigReader) GetInt(key string) (int64, error) {
	m, err := r.loadDefaults()
	if err != nil {
		return 0, err
	}
	return (&mapConfigReader{m}).GetInt(key)
}

func (r *jsonFileConfigReader) hasDefaults() bool {
	return utils.Exists(r.path)
}

func (r *jsonFileConfigReader) loadDefaults() (map[string]interface{}, error) {
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

func (r *jsonFileConfigReader) writeDefaults(m map[string]interface{}) error {
	utils.CreateDirIfNotExists(path.Dir(r.path))
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(r.path, bytes, 0755)
}

type compositeConfigReader struct {
	readers []ConfigReader
}

func (r *compositeConfigReader) GetString(key string) (string, error) {
	for _, reader := range r.readers {
		s, err := reader.GetString(key)
		if err == nil {
			return s, nil
		}
	}
	return "", fmt.Errorf("Cannot read a value for %s", key)
}

func (r *compositeConfigReader) GetInt(key string) (int64, error) {
	for _, reader := range r.readers {
		i, err := reader.GetInt(key)
		if err == nil {
			return i, nil
		}
	}
	return 0, fmt.Errorf("Cannot read a value for %s", key)
}

var homeDirConfigReader = &jsonFileConfigReader{path.Join(utils.MustGetAppDir(), "defaults.json")}

func newConfigReader() ConfigReader {
	return &compositeConfigReader{[]ConfigReader{
		&envConfigReader{
			keys:      map[string]string{INTERVAL: "LOCO_INTERVAL", SUFFIX: "LOCO_SUFFIX"},
			envReader: &defaultEnvReader{},
		},
		homeDirConfigReader,
		&mapConfigReader{map[string]interface{}{INTERVAL: int64(time.Hour * 24), SUFFIX: "%c"}},
	}}
}

func mustGetString(cr ConfigReader, key string) string {
	s, err := cr.GetString(key)
	if err != nil {
		panic(err)
	}
	return s
}

func mustGetInt(cr ConfigReader, key string) int64 {
	i, err := cr.GetInt(key)
	if err != nil {
		panic(err)
	}
	return i
}

var configReader = newConfigReader()

func DefaultConfig() *state.Config {
	return state.NewConfig(
		time.Duration(mustGetInt(configReader, INTERVAL)),
		mustGetString(configReader, SUFFIX),
	)

}
func SetDefaultConfig(c *state.Config) error {
	return homeDirConfigReader.writeDefaults(
		map[string]interface{}{INTERVAL: c.Interval, SUFFIX: c.Suffix},
	)
}

func mergeWithDefaultConfig(c *state.Config, d *state.Config) *state.Config {
	interval := d.Interval
	suffix := d.Suffix
	if c.Interval != time.Duration(0) {
		interval = c.Interval
	}
	if c.Suffix != "" {
		suffix = c.Suffix
	}
	return state.NewConfig(interval, suffix)
}

func MergeWithDefaultConfig(c *state.Config) *state.Config {
	return mergeWithDefaultConfig(c, DefaultConfig())
}

func writeDefaultConfig(w io.Writer, c *state.Config) {
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintf(tw, "Interval:\t%s\nSuffix:\t%s\n", c.Interval, c.Suffix)
	tw.Flush()
}

func WriteDefaultConfig(w io.Writer) {
	writeDefaultConfig(w, DefaultConfig())
}
