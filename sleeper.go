package sleeper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

var Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

// SlowSubscriberTimeout is the timeout after which the subscriber is considered slow and processing with it is skipped
var SlowSubscriberTimeout = 5 * time.Second

// sleeper notifies about the sleep/wake events
type sleeper struct {
	notifications chan Activity
	subscriptions map[chan Activity]struct{}
	new           chan chan Activity
	unsub         chan chan Activity
	mux           sync.RWMutex
	active        sync.WaitGroup
	starting      bool
}

// Type determines if it is a sleep or an awake type activity
type Type string

// Enum for sleep related activities
const (
	notificationListening Type = "listening" // system notification, not shared with subscribers

	NotificationSleep Type = "sleep"
	NotificationAwake Type = "awake"
)

// Activity with a specific type, indicating the sleep and awake activity.
type Activity struct {
	Type Type
}

var s = &sleeper{
	subscriptions: make(map[chan Activity]struct{}),
	new:           make(chan chan Activity),
	unsub:         make(chan chan Activity),
	notifications: make(chan Activity),
}

// Subscribe returns a channel to listen to sleep/wake activities. Multiple subscribers are supported.
// The channel is closed when the context is cancelled. Slow subscribers will be skipped after SlowSubscriberTimeout.
func Subscribe(ctx context.Context) <-chan Activity {
	return s.subscribe(ctx)
}

func (s *sleeper) subscribe(ctx context.Context) <-chan Activity {
	s.mux.Lock()
	if len(s.subscriptions) == 0 && !s.starting {
		s.run(context.Background()) // independent context
	}
	s.mux.Unlock()

	sub := make(chan Activity)
	s.new <- sub
	go func() {
		<-ctx.Done()
		s.unsub <- sub
	}()

	return sub
}

// run runs the loop and starts the notifier once we've got any subscribers to listen to Activity channel
// with machine sleep/wake activities.
func (s *sleeper) run(ctx context.Context) {
	s.active.Add(1)
	s.starting = true
	s.notifications = make(chan Activity)

	ctx, cancel := context.WithCancelCause(ctx)
	started := make(chan struct{})

	go func(s *sleeper) {
		for {
			select {
			case sub := <-s.new:
				s.mux.Lock()
				if len(s.subscriptions) == 0 {
					go s.start(cancel)
					go s.await(ctx, started)
				}
				s.subscriptions[sub] = struct{}{}
				Logger.Debug("Sleep notifier subscribed", "chan", fmt.Sprintf("%v", sub), "len", len(s.subscriptions))
				s.mux.Unlock()
			case unsub := <-s.unsub:
				s.mux.Lock()
				delete(s.subscriptions, unsub)
				close(unsub)

				if len(s.subscriptions) == 0 {
					cancel(nil)
				}
				Logger.Debug("Sleep notifier unsubscribed", "chan", fmt.Sprintf("%v", unsub), "len", len(s.subscriptions))
				s.mux.Unlock()
			case notification, ok := <-s.notifications:
				if !ok {
					s.active.Done()
					Logger.Debug("Sleep notifier stopped")
					return
				}

				if notification.Type == notificationListening {
					close(started)
					s.starting = false
					Logger.Debug("Sleep notifier started")
					continue
				}

				s.mux.RLock()
				for sub := range s.subscriptions {
					select {
					case sub <- notification:
					case <-time.After(SlowSubscriberTimeout):
						Logger.Warn("Sleep notifier slow subscriber skipped", "chan", fmt.Sprintf("%v", sub))
					}
				}
				s.mux.RUnlock()
			}
		}
	}(s)
}

// start starts the sleep notifier
func (s *sleeper) start(cancel context.CancelCauseFunc) {
	Logger.Debug("Sleep notifier starting...")
	if err := s.startNotifier(); err != nil {
		cancel(err)
		return
	}
}

// await waits for the signal to stop the notifier
func (s *sleeper) await(ctx context.Context, started <-chan struct{}) {
	<-ctx.Done()
	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		Logger.Info("Sleep notifier error, stopping...", "err", err)
		close(s.notifications)
		return
	}
	<-started
	s.stopNotifier()
	close(s.notifications)
}

// Await waits for the notifier to stop, so that we clean the resources used
func Await() {
	s.active.Wait()
}
