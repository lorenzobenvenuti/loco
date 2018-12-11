package state

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/lorenzobenvenuti/loco/utils"
)

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
	if !utils.Exists(s.dir) {
		return []*State{}, nil
	}
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
