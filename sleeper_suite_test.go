package sleeper_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"lanwen.dev/sleeper"
)

func TestShouldStartAndStop(t *testing.T) {
	RegisterTestingT(t)

	ctx, cancel := context.WithCancel(context.Background())
	events := sleeper.Subscribe(ctx)
	_ = sleeper.Subscribe(ctx)
	cancel()

	Eventually(events).Should(BeClosed())
	Eventually(func() bool {
		sleeper.Await()
		return true
	}).Should(BeTrue())
}

func TestShouldStartAndStopWithDelay(t *testing.T) {
	RegisterTestingT(t)

	ctx, cancel := context.WithCancel(context.Background())
	events := sleeper.Subscribe(ctx)
	_ = sleeper.Subscribe(ctx)
	<-time.After(100 * time.Millisecond)
	cancel()

	Eventually(events).Should(BeClosed())
	Eventually(func() bool {
		sleeper.Await()
		return true
	}).Should(BeTrue())
}
