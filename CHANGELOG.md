# Changelog

## [Unreleased]

## [v2.1.0] - 2023-04-30

### Added

- Added `State.Errors`. [#20](https://github.com/go-nacelle/process/pull/20)

## [v2.0.1] - 2022-10-10

### Added

- Added `ContextWithHealth` and `HealthFromContext`. [#18](https://github.com/go-nacelle/process/pull/18)

### Fixed

- Fixed wait group race condition. [#17](https://github.com/go-nacelle/process/pull/17)

## [v2.0.0] - 2021-05-31

### Added

- The `Finalize` methods on process instances are now invoked when defined. [#5](https://github.com/go-nacelle/process/pull/5)
- Added `WithInjecter` and `WithHealth`. [#14](https://github.com/go-nacelle/process/pull/14), [#16](https://github.com/go-nacelle/process/pull/16)
- Added `Logger` interface, `LogFields` type, and `NilLogger` variable, and `WithMetaLogger` method. [#15](https://github.com/go-nacelle/process/pull/15), [#16](https://github.com/go-nacelle/process/pull/16)

### Changed

- Added context parameters to `Init`, `Start`, `Stop`, and `Finalize` methods. [#5](https://github.com/go-nacelle/process/pull/5)
- Removed config parameters from `Init` methods. [#7](https://github.com/go-nacelle/process/pull/7)
- The `Init` methods of initializers and processors registered at the same priority initializer or process priority are now called concurrently. [#9](https://github.com/go-nacelle/process/pull/9)
- Initializers are now invoked before the processes of the same priority, but after the processes of the previous priority. [#16](https://github.com/go-nacelle/process/pull/16)
- Renamed `Process` to `Runner` and its `Start` method to `Run`. [#16](https://github.com/go-nacelle/process/pull/16)
- Extracted the `Stop` method from the `Runner` into a `Stopper` interface. [#16](https://github.com/go-nacelle/process/pull/16)
- The `Runner` interface was replaced with a `Run` function returning a `State` value that abstracts application shutdown. [#16](https://github.com/go-nacelle/process/pull/16)
- The `ProcessContainer`, `ProcessMeta`, and `InitializerMeta` interfaces were replaced with `Container`, `ContainerBuilder`, and `Meta` structs. This localizes the differences between a process and an interface to registration (and not execution). [#16](https://github.com/go-nacelle/process/pull/16)
- The `Health` interface was replaced with `Health` and `HealthComponentStatus` structs. [#16](https://github.com/go-nacelle/process/pull/16)
- Renamed `With{Initializer,Process}{Option}` to `WithMeta{Option}`, `WithProcessLogFields` to `WithMetadata`, `InjectHook` to `Injecter`, and `WithSilentExit` to `WithEarlyExit`. [#16](https://github.com/go-nacelle/process/pull/16)

### Removed

- Removed `ParallelInitializer`. [#9](https://github.com/go-nacelle/process/pull/9)
- Removed mocks package. [#11](https://github.com/go-nacelle/process/pull/11)
- Removed dependency on [go-nacelle/service](https://github.com/go-nacelle/service). [#14](https://github.com/go-nacelle/process/pull/14)
- Removed dependency on [go-nacelle/log](https://github.com/go-nacelle/log). [#15](https://github.com/go-nacelle/process/pull/15)
- Removed now irrelevant options `WithStartTimeout`, `WithHealthCheckInterval`, and `WithShutdownTimeout`. [#16](https://github.com/go-nacelle/process/pull/16)

## [v1.1.0] - 2020-10-03

### Added

- Added `WithInitializerLogFields` and `WithProcessLogFields`. [#2](https://github.com/go-nacelle/process/pull/2)
- [go-nacelle/config@v1.0.0] -> [go-nacelle/config@v1.2.1]
  - Added `FlagSourcer` that reads configuration values from the command line. [#3](https://github.com/go-nacelle/config/pull/3)
  - Added `Init` method to `Config` and `Sourcer`. [#4](https://github.com/go-nacelle/config/pull/4)
  - Added options to supply a filesystem adapter to glob, file, and directory sourcers. [#2](https://github.com/go-nacelle/config/pull/2)
- [go-nacelle/log@v1.0.0] -> [go-nacelle/log@v1.1.2]
  - Added `WithIndirectCaller` to control the number of stack frames to omit. [#2](https://github.com/go-nacelle/log/pull/2)
  - Added mocks package. [d24aad2](https://github.com/go-nacelle/log/commit/d24aad20df4c5b24dbdff3860c348af82abed169)
- [go-nacelle/service@v1.0.0] -> [go-nacelle/service@v1.0.2]
  - Added overlay container. [#1](https://github.com/go-nacelle/service/pull/1)

### Removed

- Removed dependency on [efritz/backoff](https://github.com/efritz/backoff). [bd4092d](https://github.com/go-nacelle/process/commit/bd4092d39078bba1e9cdce0e3187560fbfb172bc)
- Removed dependency on [efritz/watchdog](https://github.com/efritz/watchdog). [4121898](https://github.com/go-nacelle/process/commit/41218985f4849dc0e89c26e0fe2b274a31af61fb)
- Removed dependency on [aphistic/sweet](https://github.com/aphistic/sweet) by rewriting tests to use [testify](https://github.com/stretchr/testify). [#3](https://github.com/go-nacelle/process/pull/3)
- [go-nacelle/config@v1.0.0] -> [go-nacelle/config@v1.2.1]
  - Removed dependency on [aphistic/sweet](https://github.com/aphistic/sweet) by rewriting tests to use [testify](https://github.com/stretchr/testify). [#5](https://github.com/go-nacelle/config/pull/5)
- [go-nacelle/log@v1.0.0] -> [go-nacelle/log@v1.1.2]
  - Removed dependency on [aphistic/sweet](https://github.com/aphistic/sweet) by rewriting tests to use [testify](https://github.com/stretchr/testify). [#3](https://github.com/go-nacelle/log/pull/3)
  - Removed dependency on [aphistic/gomol](https://github.com/aphistic/gomol) by rewriting base logger internally. [4e537aa](https://github.com/go-nacelle/log/commit/4e537aa0e5a08638bfb45f5153e8deccf6e1d00d)
- [go-nacelle/service@v1.0.0] -> [go-nacelle/service@v1.0.2]
  - Removed dependency on [aphistic/sweet](https://github.com/aphistic/sweet) by rewriting tests to use [testify](https://github.com/stretchr/testify). [#2](https://github.com/go-nacelle/service/pull/2)

### Fixed

- [go-nacelle/log@v1.0.0] -> [go-nacelle/log@v1.1.2]
  - Fixed bad console output. [db6e246](https://github.com/go-nacelle/log/commit/db6e24657334615a099e39bae0359179778016e4), [45875f1](https://github.com/go-nacelle/log/commit/45875f173a0db48fc3f615d96a4f83e015cdf130)

### Changed

- [go-nacelle/log@v1.0.0] -> [go-nacelle/log@v1.1.2]
  - Changed log field blacklist from a comma-separated list to a json-encoded array. [96b9d53](https://github.com/go-nacelle/log/commit/96b9d53baff25f7c0436799f520c3d4a5970941e)

## [v1.0.1] - 2019-06-07

### Added

- Added `HasReason` to `Health`. [#1](https://github.com/go-nacelle/process/pull/1)

## [v1.0.0] - 2019-06-17

### Changed

- Migrated from [efritz/nacelle](https://github.com/efritz/nacelle).

[Unreleased]: https://github.com/go-nacelle/process/compare/v2.1.0...HEAD
[go-nacelle/config@v1.0.0]: https://github.com/go-nacelle/config/releases/tag/v1.0.0
[go-nacelle/config@v1.2.1]: https://github.com/go-nacelle/config/releases/tag/v1.2.1
[go-nacelle/log@v1.0.0]: https://github.com/go-nacelle/log/releases/tag/v1.0.0
[go-nacelle/log@v1.1.2]: https://github.com/go-nacelle/log/releases/tag/v1.1.2
[go-nacelle/service@v1.0.0]: https://github.com/go-nacelle/service/releases/tag/v1.0.0
[go-nacelle/service@v1.0.2]: https://github.com/go-nacelle/service/releases/tag/v1.0.2
[v1.0.0]: https://github.com/go-nacelle/process/releases/tag/v1.0.0
[v1.0.1]: https://github.com/go-nacelle/process/compare/v1.0.0...v1.0.1
[v1.1.0]: https://github.com/go-nacelle/process/compare/v1.0.1...v1.1.0
[v2.0.0]: https://github.com/go-nacelle/process/compare/v1.1.0...v2.0.0
[v2.0.1]: https://github.com/go-nacelle/process/compare/v2.0.0...v2.0.1
[v2.1.0]: https://github.com/go-nacelle/process/compare/v2.0.1...v2.1.0
