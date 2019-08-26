# Nacelle Process Runner [![GoDoc](https://godoc.org/github.com/go-nacelle/process?status.svg)](https://godoc.org/github.com/go-nacelle/process) [![CircleCI](https://circleci.com/gh/go-nacelle/process.svg?style=svg)](https://circleci.com/gh/go-nacelle/process) [![Coverage Status](https://coveralls.io/repos/github/go-nacelle/process/badge.svg?branch=master)](https://coveralls.io/github/go-nacelle/process?branch=master)

Process initializer and supervisor for [nacelle](https://nacelle.dev).

---

Nacelle applications can be coarsely decomposed into several behavioral categories.

<dl>
  <dt>Process</dt>
  <dd>A <a href="https://godoc.org/github.com/go-nacelle/process#Process">process</a> is a long-running component of an application such as a server or an event processor. Processes will usually run for the life of the application (e.g. until a shutdown signal is received or an error occurs). There is a special class of processes that are allowed to exit once running, but these are the exception.</dd>
</dl>

<dl>
  <dt>Initializer</dt>
  <dd>An <a href="https://godoc.org/github.com/go-nacelle/process#Initializer">initializer</a> is a component that is invoked once on application startup. An initializer usually instantiates a service or set up shared state required by other parts of the application</dd>
</dl>

<dl>
  <dt>Service</dt>
  <dd>A <a href="https://nacelle.dev/docs/core/service">service</a> is an object that encapsulates some data, state, or behavior, but does not have a rigid initialization. A service is generally instantiated by an initializer and inserted into a shared service container.</dd>
</dl>

You can see additional examples of initializer and process definition and registration in the [example repository](https://github.com/go-nacelle/example). Specifically, there is an [initializer](https://github.com/go-nacelle/example/blob/843979aaa86786784a1ca3646e8d0d1f69e29c65/internal/redis_initializer.go#L80) to create a shared Redis connection and its [registration](https://github.com/go-nacelle/example/blob/843979aaa86786784a1ca3646e8d0d1f69e29c65/cmd/worker/main.go#L9) in one of the program entrypoints. This project also provides a set of abstract base processes for common process types: an [AWS Lambda event listener](https://nacelle.dev/docs/libraries/lambdabase), a [gRPC server](https://nacelle.dev/docs/libraries/grpcbase), an [HTTP server](https://nacelle.dev/docs/libraries/httpbase), and a [generic worker process](https://nacelle.dev/docs/libraries/workerbase), which are a good and up-to-date source for best-practices.

### Process Definition

A **process** is a struct with an `Init`, a `Start`, and a `Stop` method. The initialization method that takes a [config](https://nacelle.dev/docs/core/config) object as a parameter. Each method may return an error value. For long-running processes, such as servers, the start method should be blocking. The stop method may signal the process to gracefully shut-down (via a channel or synchronization primitive), but does not need to wait until the application exits. A process is also an initializer, so the above also applies. The following example uses the database connection created by the initializer defined [below](https://nacelle.dev/docs/core/process#initializer-definition), injected by the service container, and pings it on a loop to logs its latency. The stop method closes a channel to inform the start method to unblock.

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

#### Tracking Process Health

Processes can dynamically report their own health via a shared **health tracker** object available in the [service container](https://nacelle.dev/docs/core/service). The tracker maintains a list of *reasons* that an application is not fully healthy. When this list is empty, the application should be fully functional.

A process that does not interact with the health instance is assumed to be healthy when it is live. Usage of a global health instance should be used as follows.

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

### Initializer Definition

An **initializer** is a struct with an `Init` method that takes a [config](https://nacelle.dev/docs/core/config) object and may return an error value. Initializers should **not** perform long-running computation unless it is necessary for the startup of an application as they will block additional application startup. The following example creates a connection to a Postgres database and stores it in a service container for subsequent initializers and processes to use.

```go
import (
    "github.com/go-nacelle/nacelle"
    "github.com/jmoiron/sqlx"
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

### Application Lifecycle

Concrete instances of initializers and processes are registered to a **process container** at application startup (as in the example shown below). The nacelle [bootstrapper](https://nacelle.dev/docs/core) then handles initialization and invocation of the process container and runner that will control the initialization and supervision of registered components. The following sequence of events occur on application boot.

- **Initializer boot stage:** Each initializer is booted sequentially in the order that it is registered. First, [services](https://nacelle.dev/docs/core/service) are injected into the initializer instance via the shared service container. If a previous initializer had registered a service into the container, it will be available at this time. Next, the initializer's Init method is invoked. If an error occurs in either stage, the remainder of the boot process is abandoned. Any initializer that had successfully completed is **unwound**: each initializer that implements the finalizer interface will be have their finalization method invoked. Initializers are finalized sequentially in the reverse order of their initialization.

- **Process boot stage:** After all initializers have completed successfully, the processes will begin to boot. First, services are injected into **all** processes, regardless of their priority order. If this fails, then the remainder of the boot process is abandoned and the initializers are unwound. Then, processes continue to boot in **batches** based on their (ascending) priority order. Within each batch, the following sequence of events occur.

  - **Batch initialization:** The Init method of each process is invoked sequentially in the order that it is registered. If this fails, then the remainder of the boot process is abandoned the initializers are unwound.

  - **Batch launch:** The Start method of each process is invoked. Each invocation is made concurrently and in a different goroutine. The remainder of the boot process is suspended until all processes within this priority become [healthy](#tracking-process-health). If a process returns from its Start method or does not become healthy within the given timeout period, the boot process is abandoned and the process is unwound, as described in the next stage.

- **Supervisory stage:** Once all process batches have been started and have become healthy after initialization, the supervisory stage begins. This stage listens for one of the following events and begins to unwind the process.

  - The user sends the process a signal
  - A process's Start method returns with an error
  - A process's Start method returns without an error, but is not marked for silent exit

The process is unwound by stopping each process for which a Start goroutine was created. The Stop method of each process is called concurrently with all processes within its priority batch, and each priority batch is stopped by (descending) priority order. Finally, initializers are unwound as described above.

### Registration

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

#### Process Registration Options

The following options can be set on a process during registration.

<dl>
    <dt>WithPriority</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithPriority">WithPriority</a> sets the priority group the process belongs to. Processes of the same - priority are initialized and started in parallel.</dd>

  <dt>WithProcessInitTimeout</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithProcessInitTimeout">WithProcessInitTimeout</a> sets the maximum time that the process can spend in its Init method.</dd>

  <dt>WithProcessName</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithProcessName">WithProcessName</a> sets the name of the process in log messages.</dd>

  <dt>WithProcessShutdownTimeout</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithProcessShutdownTimeout">WithProcessShutdownTimeout</a> sets the maximum time that the process can spend waiting for its Start method to unblock after its Stop method is called.</dd>

  <dt>WithProcessStartTimeout</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithProcessStartTimeout">WithProcessStartTimeout</a> sets the maximum time that the process can spend <strong>unhealthy</strong> after its Start method is called. See <a href="#health">health</a> below.</dd>

  <dt>WithSilentExit</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithSilentExit">WithSilentExit</a> sets a flag that allows a nil error value to be returned without signaling an application shutdown. This can be useful for things like leader election on startup which should not stop hot standby processes from taking client requests.</dd>
</dl>

#### Initializer Registration Options

The following options can be set on an initializer during registration.

<dl>
  <dt>WithFinalizerTimeout</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithFinalizerTimeout">WithFinalizerTimeout</a> sets the maximum time that the initializer can spend on finalization. As the application is already shutting down, this will simply log and error and unblock the finalizer.</dd>

  <dt>WithInitializerName</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithInitializerName">WithInitializerName</a> sets the name of the initializer in log messages.</dd>

  <dt>WithInitializerTimeout</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/process#WithInitializerTimeout">WithInitializerTimeout</a> sets the maximum time that the initializer can spend in its Init method. An error will be returned if this time is exceeded.</dd>
</dl>

#### Parallel Initializers

Initializers run sequentially and non-concurrently so that a previously registered initializer can provide a service required by a subsequently registered initializer. However, this default sequencing behavior is not always necessary and can increase startup time.

When initializers can be run independently (for example, creating multiple SDK client instances for a list of AWS services), it is unnecessary to run them in sequence. Groups of such services can be registered to a `ParallelInitializer`. This is a bag of initializers that run the initializers registered to it in parallel. The initializer will block its siblings until all of its children complete successfully.

```go
awsGroup := nacelle.NewParallelInitializer()
awsGroup.RegisterInitializer(NewDynamoDBClientInitializer())
awsGroup.RegisterInitializer(NewKinesisClientInitializer())
awsGroup.RegisterInitializer(NewLambdaClientInitializer())
awsGroup.RegisterInitializer(NewSQSClientInitializer())

// Register parallel group
processes.RegisterInitializer(awsGroup)
```
