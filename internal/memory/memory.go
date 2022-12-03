package memory

import (
	"container/list"
	"errors"
	"fmt"
	"github/hxia043/session/internal/driver"
	"github/hxia043/session/internal/session"
	"sync"
	"time"
)

var driverName = "memory"
var memory = &Memory{}

type MemorySession struct {
	sid          string
	timeAccessed time.Time
	value        map[interface{}]interface{}
}

func (session *MemorySession) Show() {
	fmt.Println("session id: ", session.sid)
	fmt.Println("time accessed: ", session.timeAccessed)
	fmt.Println("value: ", session.value)
}

func (session *MemorySession) Get(key interface{}) interface{} {
	if value, ok := session.value[key]; ok {
		return value
	} else {
		return nil
	}
}

func (session *MemorySession) Set(key, value interface{}) error {
	session.value[key] = value
	return nil
}

type Memory struct {
	lock     sync.Mutex
	sessions map[string]*list.Element
	list     list.List
}

func (memory *Memory) SessionGC(maxLifeTime int64) {
	memory.lock.Lock()
	defer memory.lock.Unlock()

	for {
		element := memory.list.Back()
		if element == nil {
			break
		}
		if session, ok := element.Value.(*MemorySession); ok {
			fmt.Println("time accessed: ", session.timeAccessed.Unix())
			fmt.Println("max life time: ", maxLifeTime)
			fmt.Println("time now: ", time.Now().Unix())

			if (session.timeAccessed.Unix() + maxLifeTime) < time.Now().Unix() {
				fmt.Println("remove sid: ", session.sid)
				delete(memory.sessions, session.sid)
				memory.list.Remove(element)
			} else {
				break
			}
		}
	}
}

func (memory *Memory) SessionUpdate(sid string) (driver.Session, error) {
	if element, ok := memory.sessions[sid]; ok {
		if sessionMemory, ok := element.Value.(*MemorySession); ok {
			sessionMemory.timeAccessed = time.Now()
			memory.list.MoveToFront(element)

			return sessionMemory, nil
		}
	}

	return nil, errors.New("no session find")
}

func (memory *Memory) SessionDestroy(sid string) {
	if element, ok := memory.sessions[sid]; ok {
		memory.list.Remove(element)
		delete(memory.sessions, sid)
	}
}

func (memory *Memory) SessionRead(sid string) (driver.Session, error) {
	if element, ok := memory.sessions[sid]; ok {
		return element.Value.(*MemorySession), nil
	} else {
		session, err := memory.SessionInit(sid)
		return session, err
	}
}

func (memory *Memory) SessionInit(sid string) (driver.Session, error) {
	memory.lock.Lock()
	defer memory.lock.Unlock()

	v := make(map[interface{}]interface{}, 0)
	newSession := &MemorySession{sid: sid, timeAccessed: time.Now(), value: v}
	element := memory.list.PushFront(newSession)
	memory.sessions[sid] = element

	return newSession, nil
}

func init() {
	if driverName != "" {
		memory.sessions = make(map[string]*list.Element, 0)
		session.Register(driverName, memory)
	}
}
