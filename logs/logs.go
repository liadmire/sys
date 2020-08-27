package logs

type LevelType int

const (
	DEBUG LevelType = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelPrefix = [FATAL + 1]string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]", "[FATAL]"}

var logger = newLogger()

func Debug(format string, v ...interface{}) {
	logger.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	logger.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	logger.Warn(format, v...)
}
func Error(format string, v ...interface{}) {
	logger.Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	logger.Fatal(format, v...)
}

func SetLevel(lt LevelType) {
	logger.SetLevel(lt)
}
