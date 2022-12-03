package session

import (
	"github/hxia043/session/internal/driver"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]driver.Driver)
)

func Register(name string, driver driver.Driver) {
	driversMu.RLock()
	defer driversMu.RUnlock()

	if driver == nil {
		panic("session: Register driver is nil")
	}

	if _, dup := drivers[name]; dup {
		panic("session: Register called twice for driver " + name)
	}

	drivers[name] = driver
}
