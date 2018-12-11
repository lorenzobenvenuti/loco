package state

import (
	"html/template"
	"io"
	"text/tabwriter"
	"time"
)

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
	return s.RotatedAt.IsZero() || (now.Sub(s.RotatedAt) > s.Config.Interval)
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
