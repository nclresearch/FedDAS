package log

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var logFile *os.File

//Use log package as they are thread safe
var (
	Info        *log.Logger
	Warn        *log.Logger
	Error       *log.Logger
	Fatal       *log.Logger
	infoWriter  io.Writer
	warnWriter  io.Writer
	errorWriter io.Writer
	fatalWriter io.Writer
)

//Logging functions before logging services are available
func preLogInfo(message string) {
	logMessage := "[INFO] " + time.Now().Format("2006/01/02 15:04:05 PreInit: ") + message
	fmt.Println(logMessage)
}

func preLogFatal(message string, err error) {
	logMessage := "[FATAL] " + time.Now().Format("2006/01/02 15:04:05 PreInit: ") + message
	fmt.Println(logMessage)
	panic(err)
}

func preLogError(message string, err error) {
	logMessage := "[ERROR] " + time.Now().Format("2006/01/02 15:04:05 PreInit: ") + message
	fmt.Println(logMessage)
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	preLogInfo("Checking if current directory is writable")
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if unix.Access(pwd, unix.W_OK) != nil {
		preLogFatal("no write access to current directory", errors.New("permission denied"))
	}
	logsDir := pwd + "/logs"
	if _, err = os.Stat(logsDir); os.IsNotExist(err) {
		err = os.Mkdir(logsDir, 0755)
	}
	if err != nil {
		preLogFatal("failed writing to current directory", err)
	}
	var oldLogs = make([]string, 0)
	err = filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".log") {
			oldLogs = append(oldLogs, path)
		}
		return nil
	})
	if err != nil {
		preLogError("Failed compressing logs", err)
	} else {
		//Compress old logs with gzip
		for _, oldLog := range oldLogs {
			oldLogFile, err := os.Open(oldLog)
			if err != nil {
				preLogError("Failed compressing logs", err)
				continue
			}
			logBytes, err := ioutil.ReadAll(oldLogFile)
			if err != nil {
				preLogError("Failed compressing logs", err)
				continue
			}
			oldLogFile.Close()
			var buf bytes.Buffer
			writer := gzip.NewWriter(&buf)
			_, err = writer.Write(logBytes)
			if err != nil {
				preLogError("Failed compressing logs", err)
				continue
			}
			err = writer.Close()
			if err != nil {
				preLogError("Failed compressing logs", err)
				continue
			}
			err = ioutil.WriteFile(oldLog+".gz", buf.Bytes(), 0644)
			if err != nil {
				preLogError("Failed compressing logs", err)
				continue
			}
			_ = os.Remove(oldLog)
		}
	}
	filename := time.Now().Format("2006-01-02T15:04:05-0700") + ".log"
	logFile, err = os.Create(logsDir + "/" + filename)
	if err != nil {
		preLogError("Failed creating log. Logs will only be written to stdout", err)
	}
	infoWriter = io.MultiWriter(logFile, os.Stdout)
	Info = log.New(infoWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	warnWriter = io.MultiWriter(logFile, os.Stdout)
	Warn = log.New(warnWriter, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorWriter = io.MultiWriter(logFile, os.Stderr)
	Error = log.New(errorWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	fatalWriter = io.MultiWriter(logFile, os.Stderr)
	Fatal = log.New(fatalWriter, "[FATAL] ", log.Ldate|log.Ltime|log.Lshortfile)
}
