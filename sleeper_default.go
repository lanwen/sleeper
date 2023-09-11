//go:build !darwin

package sleeper

// startNotifier starts the internal notifier
func (s *sleeper) startNotifier() error {
	// Other platforms theoretically could be supported as well
	s.notifications <- Activity{
		Type: notificationListening,
	}
	return nil
}

// stopNotifier stops the internal notifier
func (s *sleeper) stopNotifier() {
}
