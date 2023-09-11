# Sleeper

A MacOS sleep notifier for Go. Provides a simple way to detect when a machine goes to sleep.

Heavily inspired by https://github.com/prashantgupta24/mac-sleep-notifier, however:

- has a safe way to subscribe multiple listeners, 
- a way to unsubscribe,
- supports context to stop listening

## Usage

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
defer stop()

go func() {
	for n := range sleeper.Subscribe(ctx) {
		slog.Info("New notification", "state", string(n.Type))
	}
}()

sleeper.Await()
```

## Example

Working example in a `example/main.go` file.

```bash
go run example/main.go
```