package sleeper

// #cgo LDFLAGS: -framework CoreFoundation -framework IOKit
// #include "sleeper.h"
import "C"
import (
	"errors"
)

// startNotifier starts the internal notifier function which communicates with the C library.
func (s *sleeper) startNotifier() error {
	if C.registerNotifications() != 0 {
		return errors.New("failed to register sleep notifications")
	}
	return nil
}

// stopNotifier stops the internal notifier function which communicates with the C library.
func (s *sleeper) stopNotifier() {
	C.unregisterNotifications()
}

//export Started
func Started() {
	s.notifications <- Activity{
		Type: notificationListening,
	}
}

//export WillWake
func WillWake() {
	s.notifications <- Activity{
		Type: NotificationAwake,
	}
}

//export WillSleep
func WillSleep() {
	s.notifications <- Activity{
		Type: NotificationSleep,
	}
}
