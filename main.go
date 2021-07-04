package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Set struct {
	Name    string            `mapstructure:"name"`
	URL     string            `mapstructure:"url"` // url of video to download
	Options map[string]string `mapstructure:"options"`
	Flags   []string          `mapstructure:"flags"`
}

func main() {

	log.SetOutput(os.Stdout)

	log.Infoln("starting program")

	if len(os.Args) < 2 {
		log.WithError(errors.New("not enough arguments")).Fatalln("usage: ./vd_agent [CONFIG_FILE].yaml")
	}

	viper.SetConfigFile(os.Args[1])
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		log.WithError(err).Fatalln("specified config file not found")
	} else if err != nil {
		log.WithError(err).Fatalln("error reading in config file")
	}

	log.Infoln("validating configuration")

	// setting log level based on debug flag
	switch viper.GetString("logging.level") {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	log.Debugln("logging detected config below:")
	log.Debugln("--------")
	log.Debugln(spew.Sdump(viper.AllSettings()))
	log.Debugln("--------")

	// checking for missing keys
	missingKeys := []string{}
	if !viper.IsSet("polling_interval") {
		missingKeys = append(missingKeys, "polling_interval")
	}
	if !viper.IsSet("sets") {
		missingKeys = append(missingKeys, "sets")
	}
	if len(missingKeys) > 0 {
		log.WithField("missing_keys", strings.Join(missingKeys, ",")).Fatalln("missing keys in config")
	}

	var sets []Set
	viper.UnmarshalKey("sets", &sets)

	// starting metrics ticker
	interval := viper.GetDuration("polling_interval")

	ytdlPath, err := exec.LookPath("youtube-dl")
	if err != nil {
		log.WithError(err).Fatalln("could not find youtube-dl path")
	}

	log.WithField("interval", interval.String()).Infoln("starting metrics ticker")
	metricsTicker := time.NewTicker(interval)
	defer metricsTicker.Stop()
	go func() {
		for ; true; <-metricsTicker.C {
			for _, set := range sets {
				tickTime := time.Now()
				logEntry := log.WithTime(tickTime).WithField("set", set.Name)
				logEntry.Debugln("beginning set")

				// setting flags
				logEntry.Debugln("setting flags")
				var args []string
				for option, data := range set.Options {
					args = append(args, option, data)
				}
				args = append(args, set.Flags...)
				args = append(args, set.URL)

				// creating command
				ytdlCmd := exec.Command(ytdlPath, args...)
				logEntry.Debugf("final command: \"%s\"\n", ytdlCmd.String())

				// creating pipes
				stdout, err := ytdlCmd.StdoutPipe()
				if err != nil {
					logEntry.WithError(err).Errorln("error creating stdout pipe, skipping set...")
					continue
				}
				stderr, err := ytdlCmd.StderrPipe()
				if err != nil {
					logEntry.WithError(err).Errorln("error creating stderr pipe, skipping set...")
					continue
				}

				// redirect pipe output to logger
				go redirectToLogger(logEntry, logEntry.Debugf, stdout)
				go redirectToLogger(logEntry, logEntry.Errorf, stderr)

				// start youtube-dl command
				logEntry.Debugln("starting command")
				err = ytdlCmd.Start()
				if err != nil {
					logEntry.WithError(err).Errorln("error starting command, continuing...")
					continue
				}

				// wait on command (which will close pipes)
				logEntry.Debugln("waiting on cmd...")
				err = ytdlCmd.Wait()
				if err != nil {
					logEntry.WithError(err).Errorln("error waiting on cmd, continuing...")
					continue
				}

				logEntry.Debugln("set complete")
			}
			log.Infoln("waiting for next interval...")
		}
	}()

	// establishing shutdown procedure
	log.Infoln("waiting for shutdown signal...")
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt)

	<-shutdownChan

	log.Infoln("shutting down")
}

func redirectToLogger(logEntry *log.Entry, logFunc func(string, ...interface{}), file io.ReadCloser) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		logFunc("\t%s\n", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		logEntry.WithError(err).Errorln("error from stderr scanner")
	}
}
