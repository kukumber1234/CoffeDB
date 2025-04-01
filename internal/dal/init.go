package dal

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"frapo/config"
)

var once sync.Once

func Create() {
	once.Do(func() {
		logDir := *config.Dir
		logFile := logDir + "/report.log"

		if err := os.MkdirAll(logDir, 0o755); err != nil {
			fmt.Println("Error creating log directory:", err)
			return
		}

		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			file, err := os.Create(logFile)
			if err != nil {
				fmt.Println("Error creating report.log:", err)
				return
			}
			file.Close()
		}

		file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			fmt.Println("Error opening report.log:", err)
			return
		}
		
		config.Logger = slog.New(slog.NewTextHandler(file, nil))

		fmt.Println("Logger initialized successfully!")
	})
}

func Helper() {
	fmt.Println(
		`Coffee Shop Management System

Usage:
hot-coffee [--port <N>] [--dir <S>] 
hot-coffee --help

Options:
--help       Show this screen.
--port N     Port number.
--dir S      Path to the data directory.`)
}
