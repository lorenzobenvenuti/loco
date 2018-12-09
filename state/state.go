package state

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"text/tabwriter"
	"time"

	"github.com/lorenzobenvenuti/loco/intervals"
	"github.com/lorenzobenvenuti/loco/utils"
)

type Config struct {
	Interval string
	Suffix   string
}

type State struct {
	FullName  string
	CreatedAt time.Time
	RotatedAt time.Time
	Counter   int
	Config    Config
}

func (s *State) formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.RFC822)
}

func (s *State) PrettyCreatedAt() string {
	return s.formatDate(s.CreatedAt)
}

func (s *State) PrettyRotatedAt() string {
	return s.formatDate(s.RotatedAt)
}

func (s *State) FileMustBeCreated() bool {
	return s.CreatedAt.IsZero()
}

func (s *State) FileMustBeRotated(now time.Time) bool {
	interval := intervals.MustParse(s.Config.Interval)
	return s.RotatedAt.IsZero() || (now.Sub(s.RotatedAt) > interval)
}

type stateMarshaller interface {
	marshal(state *State) ([]byte, error)
	unmarshal(bytes []byte) (*State, error)
}

type jsonStateMarshaler struct{}

func (m *jsonStateMarshaler) marshal(state *State) ([]byte, error) {
	bytes, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (m *jsonStateMarshaler) unmarshal(value []byte) (*State, error) {
	s := &State{}
	err := json.Unmarshal([]byte(value), s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

type StateStorage interface {
	Store(state *State) error
	Load(fullName string) (*State, error)
	List() ([]*State, error)
	Remove(fullName string) error
}

type fileStateStorage struct {
	dir       string
	marshaler stateMarshaller
}

func (s *fileStateStorage) filename(fullName string) string {
	return fmt.Sprintf("%s.json", utils.MD5(fullName))
}

func (s *fileStateStorage) Store(state *State) error {
	utils.CreateDirIfNotExists(s.dir)
	b, err := s.marshaler.marshal(state)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(s.dir, s.filename(state.FullName)), b, os.ModePerm)
}

func (s *fileStateStorage) loadFromFile(filename string) (*State, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return s.marshaler.unmarshal(b)
}

func (s *fileStateStorage) Load(fullName string) (*State, error) {
	filename := path.Join(s.dir, s.filename(fullName))
	return s.loadFromFile(filename)
}

func (s *fileStateStorage) Remove(fullName string) error {
	filename := path.Join(s.dir, s.filename(fullName))
	return os.Remove(filename)
}

func (s *fileStateStorage) List() ([]*State, error) {
	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	states := make([]*State, 0)
	for _, file := range files {
		state, err := s.loadFromFile(path.Join(s.dir, file.Name()))
		if err == nil {
			states = append(states, state)
		} else {
			// TODO: log?
		}
	}
	return states, nil
}

func NewHomeDirStateStorage() (StateStorage, error) {
	appDir, err := utils.AppDir()
	if err != nil {
		return nil, err
	}
	return newFileStateStorage(path.Join(appDir, "logfiles"))
}

func MustCreateHomeDirStateStorage() StateStorage {
	s, err := NewHomeDirStateStorage()
	if err != nil {
		panic(err)
	}
	return s
}

func newFileStateStorage(dir string) (*fileStateStorage, error) {
	return &fileStateStorage{
		dir:       dir,
		marshaler: &jsonStateMarshaler{},
	}, nil
}

func List() ([]*State, error) {
	storage, err := NewHomeDirStateStorage()
	if err != nil {
		return nil, err
	}
	return storage.List()
}

func Remove(name string) error {
	storage, err := NewHomeDirStateStorage()
	if err != nil {
		return err
	}
	return storage.Remove(name)
}

type mapStorage struct {
	states map[string]*State
}

func (s *mapStorage) Store(state *State) error {
	s.states[state.FullName] = state
	return nil
}

func (s *mapStorage) Load(fullName string) (*State, error) {
	return s.states[fullName], nil
}

func (s *mapStorage) List() ([]*State, error) {
	return nil, nil
}

func (s *mapStorage) Remove(fullName string) error {
	delete(s.states, fullName)
	return nil
}

func NewMapStorage() StateStorage {
	return &mapStorage{
		states: make(map[string]*State),
	}
}

func WriteStates(w io.Writer, states []*State) error {
	t, err := template.New("list").Parse("FILE\tCREATED AT\tROTATED AT\tINTERVAL\tSUFFIX\n" +
		"{{range .}}{{.FullName}}\t{{.PrettyCreatedAt}}\t{{.PrettyRotatedAt}}\t{{.Config.Interval}}\t{{.Config.Suffix}}\n{{end}}\n")
	if err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.TabIndent)
	t.Execute(tw, states)
	tw.Flush()
	return nil
}

func NewState(storage StateStorage, fullName string, config Config) (*State, error) {
	s := &State{
		FullName: fullName,
		Config:   config,
	}
	return s, storage.Store(s)
}

func NewConfig(interval string, suffix string) *Config {
	return &Config{
		Interval: interval,
		Suffix:   suffix,
	}
}
