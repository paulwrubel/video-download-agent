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
	switch strings.ToLower(viper.GetString("logging.level")) {
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

	globalFlags := []string{}
	if viper.IsSet("global_flags") {
		globalFlags = viper.GetStringSlice("global_flags")
	}

	globalOpts := map[string]string{}
	if viper.IsSet("global_options") {
		globalOpts = viper.GetStringMapString("global_options")
	}

	var sets []Set
	viper.UnmarshalKey("sets", &sets)

	// starting metrics ticker
	interval := viper.GetDuration("polling_interval")

	ytdlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		log.WithError(err).Fatalln("could not find yt-dlp path")
	}

	log.WithField("interval", interval.String()).Infoln("starting metrics ticker")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	go func() {
		for ; true; <-ticker.C {
			tickTime := time.Now()
			for _, set := range sets {
				logEntry := log.WithTime(tickTime).WithField("set", set.Name)
				logEntry.Infoln("beginning set")

				// setting flags
				logEntry.Infoln("setting flags")
				var args []string
				for option, data := range set.Options {
					args = append(args, option, data)
				}
				for option, data := range globalOpts {
					args = append(args, option, data)
				}
				args = append(args, set.Flags...)
				args = append(args, globalFlags...)
				args = append(args, set.URL)

				// creating command
				ytdlpCmd := exec.Command(ytdlpPath, args...)
				logEntry.Infof("final command: \"%s\"\n", ytdlpCmd.String())

				// creating pipes
				stdout, err := ytdlpCmd.StdoutPipe()
				if err != nil {
					logEntry.WithError(err).Errorln("error creating stdout pipe, skipping set...")
					continue
				}
				stderr, err := ytdlpCmd.StderrPipe()
				if err != nil {
					logEntry.WithError(err).Errorln("error creating stderr pipe, skipping set...")
					continue
				}

				// redirect pipe output to logger
				go redirectToLogger(logEntry, logEntry.Debugf, stdout)
				go redirectToLogger(logEntry, logEntry.Errorf, stderr)

				// start youtube-dl command
				logEntry.Infoln("starting command")
				err = ytdlpCmd.Start()
				if err != nil {
					logEntry.WithError(err).Errorln("error starting command, continuing...")
					continue
				}

				// wait on command (which will close pipes)
				logEntry.Infoln("waiting on cmd...")
				err = ytdlpCmd.Wait()
				if err != nil {
					logEntry.WithError(err).Errorln("error waiting on cmd, continuing...")
					continue
				}

				logEntry.Infoln("set complete")
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
