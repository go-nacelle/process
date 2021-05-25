package process

type Logger interface {
	WithFields(LogFields) Logger
	Info(string, ...interface{})
	Warning(string, ...interface{})
	Error(string, ...interface{})
}

type LogFields map[string]interface{}
