package logwriter

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

	"github.com/lorenzobenvenuti/loco/utils"
)

type State struct {
	FullName  string
	CreatedAt time.Time
	RotatedAt time.Time
	Interval  time.Duration
	Counter   int
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
	return s.RotatedAt.IsZero() || (now.Sub(s.RotatedAt) > s.Interval)
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

type stateStorage interface {
	store(state *State) error
	load(fullName string) (*State, error)
	list() ([]*State, error)
}

type fileStateStorage struct {
	dir       string
	marshaler stateMarshaller
}

func (s *fileStateStorage) filename(fullName string) string {
	return fmt.Sprintf("%s.json", utils.MD5(fullName))
}

func (s *fileStateStorage) store(state *State) error {
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

func (s *fileStateStorage) load(fullName string) (*State, error) {
	filename := path.Join(s.dir, s.filename(fullName))
	return s.loadFromFile(filename)
}

func (s *fileStateStorage) remove(fullName string) error {
	filename := path.Join(s.dir, s.filename(fullName))
	return os.Remove(filename)
}

func (s *fileStateStorage) list() ([]*State, error) {
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

func newHomeDirStateStorage() (*fileStateStorage, error) {
	appDir, err := utils.AppDir()
	if err != nil {
		return nil, err
	}
	return newFileStateStorage(path.Join(appDir, "logfiles"))
}

func newFileStateStorage(dir string) (*fileStateStorage, error) {
	return &fileStateStorage{
		dir:       dir,
		marshaler: &jsonStateMarshaler{},
	}, nil
}

func List() ([]*State, error) {
	storage, err := newHomeDirStateStorage()
	if err != nil {
		return nil, err
	}
	return storage.list()
}

func Remove(name string) error {
	storage, err := newHomeDirStateStorage()
	if err != nil {
		return err
	}
	return storage.remove(name)
}

func WriteStates(w io.Writer, states []*State) error {
	t, err := template.New("list").Parse("FILE\tCREATED AT\tROTATED AT\n{{range .}}{{.FullName}}\t{{.PrettyCreatedAt}}\t{{.PrettyRotatedAt}}\n{{end}}\n")
	if err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.TabIndent)
	t.Execute(tw, states)
	tw.Flush()
	return nil
}
