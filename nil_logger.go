package process

type nilLogger struct{}

var NilLogger Logger = nilLogger{}

func (l nilLogger) WithFields(LogFields) Logger    { return l }
func (l nilLogger) Info(string, ...interface{})    {}
func (l nilLogger) Warning(string, ...interface{}) {}
func (l nilLogger) Error(string, ...interface{})   {}
