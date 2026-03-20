package handler

import "sync"

// ConfigChangeFunc is called when domain/origin config changes.
type ConfigChangeFunc func()

var (
	notifierMu sync.RWMutex
	notifierFn ConfigChangeFunc
)

// SetConfigChangeNotifier sets the function to call when config changes.
func SetConfigChangeNotifier(fn ConfigChangeFunc) {
	notifierMu.Lock()
	notifierFn = fn
	notifierMu.Unlock()
}

func notifyConfigChange() {
	notifierMu.RLock()
	fn := notifierFn
	notifierMu.RUnlock()
	if fn != nil {
		go fn()
	}
}
