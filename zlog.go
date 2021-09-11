package zlog

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/rightjoin/fig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type fileLog struct {
	File   *os.File
	Logger *zerolog.Logger
}

var files = map[string]fileLog{}

var isTerminal = false
var termLog *zerolog.Logger = nil
var initialized = false

// zlog reads config values and sets up
// the logging variables.
//
// log.output = [terminal] | file
// log.folder = [./logs ] location to put the file in
// log.filename = [app].log
//
func Initialize() {

	if initialized { // Do it only once
		return
	}

	output := fig.StringOr("terminal", "log.output")
	if output == "terminal" && //  user wants terminal
		!(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) { // terminal not detected
		output = "file" // default to file output
	}

	// Defualt to terminal
	isTerminal = true
	t := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	termLog = &t
	log.Logger = *termLog

	// If "file", then setup accordingly
	if output == "file" {
		filename := fig.StringOr("app", "log.filename")
		log.Logger = *Name(filename)
		isTerminal = false
	}

	initialized = true
}

func Close() {
	for _, f := range files {
		f.File.Close()
	}
}

func Name(filename string) *zerolog.Logger {

	// If its terminal output, then we don't need to
	// create any type of file loggers
	if initialized && isTerminal {
		return termLog
	}

	// Force lowercase file names
	filename = strings.ToLower(filename)

	// Add .log extension if it doesn't exist
	if filepath.Ext(filename) == "" {
		filename += ".log"
	}

	// Look it up
	flog, ok := files[filename]

	if !ok {
		// Path where this log must be created
		path := fig.StringOr("./logs", "log.folder")
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}

		// Try to create directory if it doesn't exist
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(path, os.ModePerm)
				if err != nil {
					log.Fatal().Err(err)
				}
			}
		}

		// Open file to start logging
		path += filename
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal().Err(err)
		}

		logger := log.Output(f)
		flog = fileLog{
			File:   f,
			Logger: &logger,
		}

		files[filename] = flog
	}

	return flog.Logger
}

func init() {
	Initialize()
}
