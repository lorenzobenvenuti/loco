package state

import "encoding/json"

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
