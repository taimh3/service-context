package slogc

import (
	"flag"
	"log/slog"
	"os"
	"strings"

	sctx "github.com/taimaifika/service-context"
)

type config struct {
	logLevel  string
	logFormat string
}

type slogComponent struct {
	id string

	*config

	handler slog.Handler
	opts    *slog.HandlerOptions
}

func NewSlogComponent() *slogComponent {
	return &slogComponent{
		id:     "slog",
		config: new(config),
		opts:   &slog.HandlerOptions{},
	}
}

func (s slogComponent) SetLogLevel(l string) {
	switch strings.ToUpper(l) {
	case slog.LevelInfo.String():
		s.opts.Level = slog.LevelInfo
	case slog.LevelWarn.String():
		s.opts.Level = slog.LevelWarn
	case slog.LevelError.String():
		s.opts.Level = slog.LevelError
	default:
		s.opts.Level = slog.LevelDebug
	}
}

func (s *slogComponent) SetLogFormat(f string) {
	if strings.EqualFold(s.logFormat, "text") {
		s.handler = slog.NewTextHandler(os.Stdout, s.opts)
	} else {
		s.handler = slog.NewJSONHandler(os.Stdout, s.opts)
	}
}

func (s *slogComponent) ID() string {
	return s.id
}

func (s *slogComponent) InitFlags() {
	flag.StringVar(&s.logLevel, s.id+"-log-level", "debug", "Log level: debug | info | warm | error  . Default: debug")
	flag.StringVar(&s.logFormat, s.id+"-log-format", "text", "Log format: json | text . Default: text")
}

func (s *slogComponent) Activate(_ sctx.ServiceContext) error {
	// set log level
	s.SetLogLevel(s.logLevel)

	// set log format
	s.SetLogFormat(s.logFormat)

	// create slog logger
	logger := slog.New(s.handler)

	// set slog default logger
	slog.SetDefault(logger)
	return nil
}

func (s *slogComponent) Stop() error {
	return nil
}
