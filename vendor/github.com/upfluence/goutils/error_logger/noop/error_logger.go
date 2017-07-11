package noop

import (
	"github.com/upfluence/goutils/error_logger"
	"github.com/upfluence/goutils/log"
)

func init() {
	error_logger.DefaultErrorLogger = &Logger{}
}

type Logger struct{}

func NewErrorLogger() *Logger { return &Logger{} }

func (l *Logger) Capture(err error, opts *error_logger.Options) error {
	log.Error(err.Error())
	return nil

}
func (l *Logger) Close() {}
