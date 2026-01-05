# fg-lib

`fg-lib` is a collection of robust, lightweight utility libraries for Go applications.

## Modules

### [Log](log/README.md)
A structured logging wrapper based on `log/slog` and `zap` with support for file rotation, dynamic levels, and context injection.

### [Config](config/README.md)
A struct-based configuration loader supporting environment variables, files (YAML/JSON), defaults, and auto-refresh.

### [DI](di/README.md)
A simple dependency injection container wrapper based on `uber-go/dig` for managing application components.

### [Graceful](graceful/README.md)
A library for managing application lifecycle, providing graceful shutdown handling with signals and timeouts.

### [Utils](utils/README.md)
A collection of terminal utilities including JSON diff visualization, table rendering, and interactive text editing.

## Installation

```bash
go get github.com/fuguiw/fg-lib
```

## License

See [LICENSE](LICENSE) file.
