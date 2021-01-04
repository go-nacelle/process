package process

type (
	Logger interface {
		WithFields(LogFields) Logger
		Info(string, ...interface{})
		Warning(string, ...interface{})
		Error(string, ...interface{})
	}

	LogFields map[string]interface{}
	nilLogger struct{}
)

var NilLogger Logger = &nilLogger{}

func (n *nilLogger) WithFields(LogFields) Logger    { return n }
func (n *nilLogger) Info(string, ...interface{})    {}
func (n *nilLogger) Warning(string, ...interface{}) {}
func (n *nilLogger) Error(string, ...interface{})   {}
