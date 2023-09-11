package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"log/slog"

	"lanwen.dev/sleeper"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	ctx2, cancel2 := context.WithCancel(ctx)
	events := sleeper.Subscribe(ctx2)

	logger.Warn("This subscriber will be stopped to see how disabling works during runtime", "chan", fmt.Sprintf("%v", events))
	time.AfterFunc(200*time.Millisecond, cancel2)
	sleeper.Await()

	logger.Warn("Create another 2 subscriptions to see all of them receive events. SIGINT stops everything")
	go func() {
		for n := range sleeper.Subscribe(ctx) {
			logger.Info("New notification", "subscriber", 1, "state", string(n.Type))
		}
	}()

	for n := range sleeper.Subscribe(ctx) { // this blocks until SIGINT
		logger.Info("New notification", "subscriber", 2, "state", string(n.Type))
	}

	sleeper.Await()
}
