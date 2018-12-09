package logwriter

import (
	"io"
	"os"
	"time"

	"github.com/lorenzobenvenuti/loco/filename"
	"github.com/lorenzobenvenuti/loco/state"
	"github.com/lorenzobenvenuti/loco/utils"
)

type nowProvider interface {
	Now() time.Time
}

type timeNowProvider struct{}

func (t *timeNowProvider) Now() time.Time {
	return time.Now()
}

var defaultNowProvider = &timeNowProvider{}

var fileNameGenerator = filename.NewFileNameGenerator()

type LogWriter struct {
	state             *state.State
	file              *os.File
	stateStorage      state.StateStorage
	nowProvider       nowProvider
	fileNameGenerator filename.FileNameGenerator
}

func (lw *LogWriter) openLogFile() (*os.File, error) {
	return os.OpenFile(lw.state.FullName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func (lw *LogWriter) createLogFile() error {
	f, err := lw.openLogFile()
	if err != nil {
		return utils.Wrap(err, "Cannot open log file")
	}
	lw.state.CreatedAt = lw.nowProvider.Now()
	lw.state.RotatedAt = lw.nowProvider.Now()
	lw.stateStorage.Store(lw.state)
	lw.file = f
	return nil
}

func (lw *LogWriter) rotateLogFile() error {
	err := lw.Close()
	if err != nil {
		return utils.Wrap(err, "Error closing log writer")
	}
	lw.state.RotatedAt = lw.nowProvider.Now()
	lw.state.Counter++
	if utils.Exists(lw.state.FullName) {
		rotated := lw.fileNameGenerator.FileName(lw.state)
		err = os.Rename(lw.state.FullName, rotated)
		if err != nil {
			return utils.Wrapf(err, "Error renaming log file to %s", rotated)
		}
	}
	f, err := lw.openLogFile()
	if err != nil {
		return utils.Wrap(err, "Error opening log writer")
	}
	lw.file = f
	lw.stateStorage.Store(lw.state)
	return nil
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	if lw.state.FileMustBeCreated() {
		err := lw.createLogFile()
		if err != nil {
			return 0, utils.Wrap(err, "Error creating log file")
		}
	} else if lw.state.FileMustBeRotated(lw.nowProvider.Now()) {
		lw.rotateLogFile()
	} else if lw.file == nil {
		f, err := lw.openLogFile()
		if err != nil {
			return 0, utils.Wrap(err, "Error opening log writer")
		}
		lw.file = f
	}
	return lw.file.Write(p)
}

func (lw *LogWriter) Close() error {
	if lw.file == nil {
		return nil
	}
	return lw.file.Close()
}

func LoadWriter(fullName string) (io.WriteCloser, error) {
	storage, err := state.NewHomeDirStateStorage()
	if err != nil {
		return nil, err
	}
	return loadWriter(storage, defaultNowProvider, fileNameGenerator, fullName)
}

func loadWriter(
	storage state.StateStorage,
	nowProvider nowProvider,
	fileNameGenerator filename.FileNameGenerator,
	fullName string,
) (*LogWriter, error) {
	s, err := storage.Load(fullName)
	if err != nil {
		return nil, err
	}
	return writer(s, storage, nowProvider, fileNameGenerator), nil
}

func NewWriter(fullName string, interval time.Duration, suffix string) (io.WriteCloser, error) {
	storage, err := state.NewHomeDirStateStorage()
	if err != nil {
		return nil, err
	}
	return newWriter(storage, defaultNowProvider, fileNameGenerator, fullName, interval, suffix)
}

func newWriter(
	storage state.StateStorage,
	nowProvider nowProvider,
	fileNameGenerator filename.FileNameGenerator,
	fullName string,
	interval time.Duration,
	suffix string,
) (*LogWriter, error) {
	s, err := state.NewConfig(storage, fullName, interval, suffix)
	if err != nil {
		return nil, err
	}
	return writer(s, storage, nowProvider, fileNameGenerator), nil
}

func writer(
	s *state.State,
	storage state.StateStorage,
	nowProvider nowProvider,
	fileNameGenerator filename.FileNameGenerator,
) *LogWriter {
	return &LogWriter{
		state:             s,
		stateStorage:      storage,
		nowProvider:       nowProvider,
		fileNameGenerator: fileNameGenerator,
	}
}
