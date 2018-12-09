package filename

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/lorenzobenvenuti/loco/state"
)

type FileNameGenerator interface {
	FileName(state *state.State) string
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

type patternTranslator func(s *state.State) string

type suffixFileNameGenerator struct {
	patternTranslators map[string]patternTranslator
}

func (m *suffixFileNameGenerator) add(pattern string, t patternTranslator) {
	m.patternTranslators[pattern] = t
}

func (m *suffixFileNameGenerator) suffixFromState(state *state.State) string {
	tokens := strings.Split(state.Suffix, "%%")
	for i, _ := range tokens {
		for k, v := range m.patternTranslators {
			tokens[i] = strings.Replace(tokens[i], k, v(state), -1)
		}
	}
	return strings.Join(tokens, "%")
}

func (m *suffixFileNameGenerator) FileName(state *state.State) string {
	dir, file := path.Split(state.FullName)
	basename, ext := splitBaseNameAndExtension(file)
	suffix := m.suffixFromState(state)
	return path.Join(dir, fmt.Sprintf("%s.%s%s", basename, suffix, ext))
}

func NewFileNameGenerator() FileNameGenerator {
	generator := &suffixFileNameGenerator{make(map[string]patternTranslator)}
	generator.add("%c", func(s *state.State) string { return strconv.Itoa(s.Counter) })
	generator.add("%Y", func(s *state.State) string { return strconv.Itoa(s.RotatedAt.Year()) })
	generator.add("%m", func(s *state.State) string { return fmt.Sprintf("%02d", int(s.RotatedAt.Month())) })
	generator.add("%d", func(s *state.State) string { return fmt.Sprintf("%02d", s.RotatedAt.Day()) })
	generator.add("%H", func(s *state.State) string { return fmt.Sprintf("%02d", s.RotatedAt.Hour()) })
	generator.add("%M", func(s *state.State) string { return fmt.Sprintf("%02d", s.RotatedAt.Minute()) })
	generator.add("%S", func(s *state.State) string { return fmt.Sprintf("%02d", s.RotatedAt.Second()) })
	return generator
}
