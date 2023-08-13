package log

import "go.uber.org/zap"

var zLog *zap.SugaredLogger

func init() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	zLog = logger.Sugar()
}

func Errorf(template string, arg ...interface{}) {
	zLog.Errorf(template, arg)
}

// Info uses fmt.Sprint to log a templated message.
func Info(template string) {
	zLog.Info(template)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, arg ...interface{}) {
	zLog.Infof(template, arg)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infow(msg string, arg ...interface{}) {
	zLog.Infow(msg, arg...)
}
