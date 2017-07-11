package error_logger

var DefaultErrorLogger ErrorLogger

type Options map[string]interface{}

type ErrorLogger interface {
	Capture(error, *Options) error
	Close()
}

func Setup(logger ErrorLogger) {
	DefaultErrorLogger = logger
	if e := recover(); e != nil {
		if err, ok := e.(error); ok {
			logger.Capture(err, nil)
			logger.Close()
			panic(err.Error())
		}
	}
}
