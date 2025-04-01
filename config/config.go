package config

import (
	"flag"
	"log/slog"
)

var (
	Help = flag.Bool("help", false, "Show help message")
	Dir  = flag.String("dir", "Logger", "Path to the data directory")

	Logger *slog.Logger
)
