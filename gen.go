package process

//go:generate go-mockgen -f github.com/go-nacelle/process -i maximumProcess -o mock_types_test.go

type maximumProcess interface {
	Initializer
	Process
	Stoppable
	Finalizer
}
