package mocks

//go:generate go-mockgen -f github.com/go-nacelle/process -i Finalizer -o finalizer.go
//go:generate go-mockgen -f github.com/go-nacelle/process -i Initializer -o initializer.go
//go:generate go-mockgen -f github.com/go-nacelle/process -i Process -o process.go
//go:generate go-mockgen -f github.com/go-nacelle/process -i ProcessContainer -o process_container.go
//go:generate go-mockgen -f github.com/go-nacelle/process -i Runner -o runner.go
