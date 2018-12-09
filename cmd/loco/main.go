package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/lorenzobenvenuti/loco/defaults"
	"github.com/lorenzobenvenuti/loco/intervals"
	"github.com/lorenzobenvenuti/loco/logwriter"
	"github.com/lorenzobenvenuti/loco/state"
)

var logger = log.New(os.Stderr, "", 0)

func createConfig(file string, interval string, suffix string) {
	absPath, err := filepath.Abs(file)
	if err != nil {
		logger.Fatalf("Cannot convert path %s: %s", file, err)
	}
	err = intervals.Validate(interval)
	if err != nil {
		logger.Fatalf("Cannot parse interval %s: %s", interval, err)
	}
	storage := state.MustCreateHomeDirStateStorage()
	c := state.NewConfig(interval, suffix)
	_, err = state.NewState(storage, absPath, *c)
	if err != nil {
		logger.Fatalf("Cannot store configuration: %s", err)
	}
}

func collectLogs(file string, tee bool) {
	absPath, err := filepath.Abs(file)
	if err != nil {
		logger.Fatalf("Cannot convert path %s: %s", file, err)
	}
	lw, err := logwriter.LoadWriter(absPath)
	if err != nil {
		p := defaults.NewRuntimeDefaultsProvider()
		interval := defaults.MustGetInterval(p)
		suffix := defaults.MustGetSuffix(p)
		lw, err = logwriter.NewWriter(absPath, interval, suffix)
		if err != nil {
			logger.Fatalf("Cannot create a new writer: %s", err)
		}
	}
	defer lw.Close()
	var w io.Writer
	if tee {
		w = io.MultiWriter(lw, os.Stdout)
	} else {
		w = lw
	}
	io.Copy(w, os.Stdin)
}

func listLogFiles() {
	l, err := state.List()
	if err != nil {
		logger.Fatal(err)
	}
	err = state.WriteStates(os.Stdout, l)
	if err != nil {
		logger.Fatal(err)
	}
}

func removeLogFile(name string) {
	err := state.Remove(name)
	if err != nil {
		logger.Fatalf("Cannot remove file %s: %s", name, err)
	}
}

func showOrSetDefaults(interval string, suffix string) {
	if interval == "" && suffix == "" {
		defaults.WriteDefaults(os.Stdout, defaults.NewStaticDefaultsProvider())
		return
	}
	if interval != "" {
		err := defaults.SetDefaultInterval(interval)
		if err == nil {
			logger.Printf("Default interval set to %s", interval)
		} else {
			logger.Fatalf("Cannot set default interval: %s", err.Error())
		}
	}
	if suffix != "" {
		err := defaults.SetDefaultSuffix(suffix)
		if err == nil {
			logger.Printf("Default suffix set to %s", suffix)
		} else {
			logger.Fatalf("Cannot set default suffix: %s", err.Error())
		}
	}
}

func main() {
	app := kingpin.New("loco", "A log collector")
	config := app.Command("config", "Configures a log file")
	configInterval := config.Flag("interval", "Rotate interval").Short('i').String()
	configSuffix := config.Flag("suffix", "Rotated file suffix").Short('s').String()
	configFile := config.Arg("file", "Log file").Required().String()
	collect := app.Command("collect", "Collects stdin and redirects to a log file")
	collectTee := collect.Flag("tee", "Write to log file and stdout").Short('t').Bool()
	collectFile := collect.Arg("file", "Log file").Required().String()
	list := app.Command("list", "Lists the registered log files")
	remove := app.Command("remove", "Removes a log file")
	removeFile := remove.Arg("file", "Log file").Required().String()
	defaults := app.Command("defaults", "Shows or sets default options")
	defaultsInterval := defaults.Flag("interval", "Rotate interval").Short('i').String()
	defaultsSuffix := defaults.Flag("suffix", "Rotated file suffix").Short('s').String()
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case config.FullCommand():
		createConfig(*configFile, *configInterval, *configSuffix)
	case collect.FullCommand():
		collectLogs(*collectFile, *collectTee)
	case list.FullCommand():
		listLogFiles()
	case remove.FullCommand():
		removeLogFile(*removeFile)
	case defaults.FullCommand():
		showOrSetDefaults(*defaultsInterval, *defaultsSuffix)
	}
}
