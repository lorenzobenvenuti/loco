package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/lorenzobenvenuti/loco/defaults"
	"github.com/lorenzobenvenuti/loco/intervals"
	"github.com/lorenzobenvenuti/loco/logwriter"
)

var logger = log.New(os.Stderr, "", 0)

func createConfig(file string, intervalExpr string) {
	absPath, err := filepath.Abs(file)
	if err != nil {
		logger.Fatalf("Cannot convert path %s: %s", file, err)
	}
	interval, err := intervals.Parse(intervalExpr)
	if err != nil {
		logger.Fatalf("Cannot parse interval %s: %s", intervalExpr, err)
	}
	err = logwriter.NewConfig(absPath, time.Duration(interval))
	if err != nil {
		logger.Fatalf("Cannot store configuration: %s", err)
	}
}

func defaultInterval() time.Duration {
	interval := defaults.MustGetInterval(defaults.NewRuntimeDefaultsProvider())
	return intervals.MustParse(interval)
}

func collectLogs(file string, tee bool) {
	absPath, err := filepath.Abs(file)
	if err != nil {
		logger.Fatalf("Cannot convert path %s: %s", file, err)
	}
	lw, err := logwriter.LoadWriter(absPath)
	if err != nil {
		lw, err = logwriter.NewWriter(absPath, defaultInterval())
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
	l, err := logwriter.List()
	if err != nil {
		logger.Fatal(err)
	}
	err = logwriter.WriteStates(os.Stdout, l)
	if err != nil {
		logger.Fatal(err)
	}
}

func removeLogFile(name string) {
	err := logwriter.Remove(name)
	if err != nil {
		logger.Fatalf("Cannot remove file %s: %s", name, err)
	}
}

func showOrSetDefaults(intervalExpr string) {
	if intervalExpr == "" {
		fmt.Println(defaults.DefaultsToString(defaults.NewStaticDefaultsProvider()))
	} else {
		err := defaults.SetDefaultInterval(intervalExpr)
		if err == nil {
			logger.Printf("Default interval set to %s", intervalExpr)
		} else {
			logger.Fatalf("Cannot set default interval: %s", err.Error())
		}
	}
}

func main() {
	app := kingpin.New("loco", "A log collector")
	config := app.Command("config", "Configures a log file")
	configInterval := config.Flag("interval", "Rotate interval").Short('i').String()
	configFile := config.Arg("file", "Log file").Required().String()
	collect := app.Command("collect", "Collects stdin and redirects to a log file")
	collectTee := collect.Flag("tee", "Write to log file and stdout").Short('t').Bool()
	collectFile := collect.Arg("file", "Log file").Required().String()
	list := app.Command("list", "Lists the registered log files")
	remove := app.Command("remove", "Removes a log file")
	removeFile := remove.Arg("file", "Log file").Required().String()
	defaults := app.Command("defaults", "Shows or sets default options")
	defaultsInterval := defaults.Flag("interval", "Rotate interval").Short('i').String()
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case config.FullCommand():
		createConfig(*configFile, *configInterval)
	case collect.FullCommand():
		collectLogs(*collectFile, *collectTee)
	case list.FullCommand():
		listLogFiles()
	case remove.FullCommand():
		removeLogFile(*removeFile)
	case defaults.FullCommand():
		showOrSetDefaults(*defaultsInterval)
	}
}
