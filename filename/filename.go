package filename

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/lorenzobenvenuti/loco/logwriter"
)

type FileNameGenerator interface {
	FileName(state *logwriter.State) string
}

type suffixFileNameGenerator struct{}

type translator func(s *logwriter.State) string

type translatorRegistry struct {
	translators map[string]translator
}

func (r *translatorRegistry) add(pattern string, t translator) {
	r.translators[pattern] = t
}

func (r *translatorRegistry) translate(state *logwriter.State, text string) string {
	for k, v := range r.translators {
		text = strings.Replace(text, k, v(state), -1)
	}
	return text
}

var registry *translatorRegistry = &translatorRegistry{make(map[string]translator)}

func init() {
	registry.add("%c", func(s *logwriter.State) string { return strconv.Itoa(s.Counter) })
	registry.add("%Y", func(s *logwriter.State) string { return strconv.Itoa(s.RotatedAt.Year()) })
	registry.add("%m", func(s *logwriter.State) string { return fmt.Sprintf("%02d", int(s.RotatedAt.Month())) })
	registry.add("%d", func(s *logwriter.State) string { return fmt.Sprintf("%02d", s.RotatedAt.Day()) })
	registry.add("%H", func(s *logwriter.State) string { return fmt.Sprintf("%02d", s.RotatedAt.Hour()) })
	registry.add("%M", func(s *logwriter.State) string { return fmt.Sprintf("%02d", s.RotatedAt.Minute()) })
	registry.add("%S", func(s *logwriter.State) string { return fmt.Sprintf("%02d", s.RotatedAt.Second()) })
}

func splitBaseNameAndExtension(file string) (string, string) {
	ext := path.Ext(file)
	var basename string
	if ext == "" {
		basename = file
	} else {
		basename = file[:strings.LastIndex(file, ext)]
	}
	return basename, ext
}

func suffixFromState(state *logwriter.State) string {
	suffix := state.Suffix
	tokens := strings.Split(suffix, "%%")
	for i, token := range tokens {
		tokens[i] = registry.translate(state, token)
	}
	return strings.Join(tokens, "%")
}

func (m *suffixFileNameGenerator) FileName(state *logwriter.State) string {
	dir, file := path.Split(state.FullName)
	basename, ext := splitBaseNameAndExtension(file)
	suffix := suffixFromState(state)
	return path.Join(dir, fmt.Sprintf("%s.%s%s", basename, suffix, ext))
}

func NewFileNameGenerator() FileNameGenerator {
	return &suffixFileNameGenerator{}
}
