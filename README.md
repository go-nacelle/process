# Nacelle Process Runner [![GoDoc](https://godoc.org/github.com/go-nacelle/process?status.svg)](https://godoc.org/github.com/go-nacelle/process) [![CircleCI](https://circleci.com/gh/go-nacelle/process.svg?style=svg)](https://circleci.com/gh/go-nacelle/process) [![Coverage Status](https://coveralls.io/repos/github/go-nacelle/process/badge.svg?branch=master)](https://coveralls.io/github/go-nacelle/process?branch=master)

Process initializer and supervisor for [nacelle](https://github.com/go-nacelle/nacelle).

---

Applications written with nacelle can be composed into distinct categories. An **initializer** is something that runs once on application startup. These usually instantiate a [service](https://nacelle.dev/docs/core/service) or set up shared state required by other parts of the application. A **process** is something that starts and runs continually throughout the life of the application. This may be a server, something that reacts to external events, or something that operates on internal timers. There is a special class of processes that are allowed to exit once running, but these are the exception.

This project also provides a set of abstract base processes for common process types: an [AWS Lambda event listener](https://nacelle.dev/docs/libraries/lambdabase), a [gRPC server](https://nacelle.dev/docs/libraries/grpcbase), an [HTTP server](https://nacelle.dev/docs/libraries/httpbase), and a [generic worker process](https://nacelle.dev/docs/libraries/workerbase)

### Usage

An **initializer** is a struct with an `Init` method that takes a [config](https://nacelle.dev/docs/core/config) object. Initializers should **not** perform long-running computation unless it is necessary for the startup of an application as they will block additional application startup. The following example creates a connection to a Postres database and stores it in a service container for subsequent initializers and processes to use.

```go
import (
    "github.com/go-nacelle/nacelle
    "github.com/jmoiron/sqlx
    _ "github.com/lib/pq"
)

type DatabaseInitializer struct {
    Services nacelle.ServiceContainer `service:"services"`
}

type Config struct {
    ConnectionString string `env:"connection_string" required:"true"`
}

func (i *DatabaseInitializer) Init(config nacelle.Config) error {
    dbConfig := &Config{}
    if err := config.Load(dbConfig); err != nil {
        return err
    }

    db, err := sqlx.Open("postgres", dbConfig.ConnectionString)
    if err != nil {
        return err
    }

    return i.Services.Set("db", db)
}
```

An initializer can also be a **finalizer** if it defines a `Finalize` method. This can be useful for initializers that need to do some cleanup action before the application shuts down such as closing a log or profile file, closing remote connections, or ensuring that certain buffers get flushed before the application ends.

A **process** is a struct with a `Start` and a `Stop` method. For long-running processes, such as servers, the start method should be blocking. The stop method may signal the process to gracefully shut-down (via a channel or synchronization primitive), but does not need to wait until the application exits. A process is also an initializer, so the above also applies. The following example uses the database connection created by the initializer above, injected by the service container, and pings it on a loop and logs its latency. The stop method closes a channel to inform the start method to unblock.

```go
type PingProcess struct {
    DB           *sqlx.DB       `service:"db"`
    Logger       nacelle.Logger `service:"logger"`
    halt         chan struct{}
    once         sync.Once
    tickInterval time.Duration
}

type Config struct {
    TickInterval int `env:"tick_interval"`
}

func (p *PingProcess) Init(config nacelle.Config) error {
    pingConfig := &Config{}
    if err := config.Load(pingConfig); err != nil {
        return err
    }

    p.tickInterval = time.Duration(pingConfig.tickInterval) * time.Second
    return nil
}

func (p *PingProcess) Start() error {
    for {
        select {
            case <-p.halt:
                return nil
            case <-time.After(p.tickInterval):
        }

        start := time.Now()
        err := p.DB.Ping()
        duration := time.Now().Sub(start)
        durationMs := float64(duration) / float64(time.Milliseconds)

        if err != nil {
            return err
        }

        p.Logger.Debug("Ping took %.2fms", durationMs)
    }

    return nil
}

func (p *PingProcess) Stop() error {
    p.once.Do(func() {
        close(p.halt)
    })

    return nil
}
```

### Registration

Initializers and processes are registered to a `ProcessContainer` which is subsequently used by a runner in order to ensure the correct initialization order. The bootstrapper [nacelle](https://nacelle.dev/docs/core) handles initialization and invocation of the runner.

Initializers are run in sequence, one at a time, in the order that they are registered to the process container. If any initializer fails, then any previously successful initializers are finalized and the application exits with an error. Finalizers are run in the reverse order that they are initialized.

Once all initializers have completed successfully, processes are started *by priority* (low to high). When starting a priority group, the init method of each process will run in parallel. Then, the start method fo each process will run in parallel. If any process fails to initialize, then all lower priority groups are stopped, all initializers are finalized, and the application exits with an error. This is also true if any process returns an error value from its start method. Once all processes have started, they are monitored for errors. Sending a signal to the application will attempt to gracefully shut down all processes in reverse priority.

The following options can be set on an initializer during registration.

- **WithInitializerName** sets the name of the initializer in log messages.
- **WithInitializerTimeout** sets the maximum time that the initializer can spend in its Init method. An error will be returned if this time is exceeded.
- **WithFinalizerTimeout** sets the maximum time that the initializer can spend on finalization. As the application is already shutting down, this will simply log and error and unblock the finalizer.

The following options can be set on a process during registration.

- **WithProcessName** sets the name of the process in log messages.
- **WithPriority** sets the priority group the process belongs to. Processes of the same - priority are initialized and started in parallel.
- **WithSilentExit** sets a flag that allows a nil error value to be returned without signaling an application shutdown. This can be useful for things like leader election on startup which should not stop hot standby processes from taking client requests.
- **WithProcessInitTimeout** sets the maximum time that the process can spend in its Init method.
- **WithProcessStartTimeout** sets the maximum time that the process can spend *unhealthy* after its Start method is called. See [health](#health) below.
- **WithProcessShutdownTimeout** sets the maximum time that the process can spend waiting for its Start method to unblock after its Stop method is called.

The following options can be set for the runner itself. Again, these configuration options be supplied through the nacelle bootstrapper.

- **WithStartTimeout** sets the maximum time that the application can spend in startup.
- **WithHealthCheckBackoff** sets the [backoff](https://github.com/efritz/backoff) instance used to check the health of processes during startup.
- **WithShutdownTimeout** sets the maximum time that the application can spend shutting down.

The following example registers a cache and a database initializer, then registers a health server, an HTTP server, and a gRPC server. The database initializer must complete within five seconds or the application fails (and, if orchestrated properly, will restart). The health server is registered at a lower priority than the other two. This ensures that the health server can take requests as early as the HTTP and gRPC servers, and also ensures that it will be shut down only after the higher-priority processes have been shut down.

```go
func setupProcesses(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
    processes.RegisterInitializer(
        NewCacheInitializer(),
        nacelle.WithInitializerName("cache"),
    )

    processes.RegisterInitializer(
        NewDatabaseInitializer(),
        nacelle.WithInitializerName("db"),
        nacelle.WithInitializerTimeout(time.Second * 5),
    )

    processes.RegisterProcess(
        NewHealthServer(),
        nacelle.WithProcessName("health-server"),
    )

    processes.RegisterProcess(
        NewHTTPServer(),
        nacelle.WithProcessName("http-server"),
        nacelle.WithPriority(1),
    )

    processes.RegisterProcess(
        NewGRPCServer(),
        nacelle.WithProcessName("grpc-server"),
        nacelle.WithPriority(1),
    )

    return nil
}
```

### Parallel Initializers

Initializers are run one at a time in the order in which they are registered. This guarantees that any service created by an initializer registered previously will be available to all later initializers and processes.

When initializers can be run independently (for example, creating multiple SDK client instances for a list of AWS services), it is unnecessary to run them in sequence. Groups of such services can be registered to a `ParallelInitializer`, which will run all of its initializers in parallel. The initializer will block its siblings until all of its initializers complete.

```go
awsGroup := nacelle.NewParallelInitializer()
awsGroup.RegisterInitializer(NewDynamoDBClientInitializer(), nacelle.WithInitializerName("dynamodb"))
awsGroup.RegisterInitializer(NewKinesisClientInitializer(), nacelle.WithInitializerName("kinesis"))
awsGroup.RegisterInitializer(NewLambdaClientInitializer(), nacelle.WithInitializerName("lambda"))
awsGroup.RegisterInitializer(NewSQSClientInitializer(), nacelle.WithInitializerName("sqs"))

processes.RegisterInitializer(awsGroup, nacelle.WithInitializerName("aws"))
```

### Health

A `Health` struct is provided that simply tracks a list of *reasons* that an application is not yet fully healthy. Once this list is empty, the application should be fully functional. The runner ensure that processes are both *live* (yet to fail) and *healthy*. The later property must be defined per process. A process that does not interact with the health instance is assumed to be healthy when it is live. Usage of a global health instance should be used as follows.

```go
type HealthConsciousProcess struct {
    Health nacelle.Health `service:"health"`
}

func (p *HealthConsciousProcess) Init(config nacelle.Config) error {
    // Before the process starts, add an "unhealthy" token unique to this process
    if err := p.Health.AddReason("health-conscious-process"); err != nil {
        return err
    }

    // ...
    return nil
}

func (p *HealthConsciousProcess) Start() error {
    // pre-healthy processing
    // ...

    // Once healthy, remove the reason registered above
    if err := p.Health.RemoveReason("health-conscious-process"); err != nil {
        return err
    }

    // post-healthy processing
    // ...

    return nil
}

func (p *HealthConsciousProcess) Stop() error {
    // ...
    return nil
}
```
