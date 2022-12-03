package session

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github/hxia043/session/internal/driver"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Manager struct {
	driver      driver.Driver
	lock        sync.Mutex
	cookieName  string
	maxLifeTime int64
}

func (manager *Manager) generateSessionId() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	manager.driver.SessionGC(manager.maxLifeTime)
	time.AfterFunc(time.Duration(manager.maxLifeTime)*time.Second, func() { manager.GC() })
}

func (manager *Manager) SeesionUpdate(w http.ResponseWriter, r *http.Request) (driver.Session, error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	cookie, _ := r.Cookie(manager.cookieName)
	sid, _ := url.QueryUnescape(cookie.Value)
	session, err := manager.driver.SessionUpdate(sid)

	return session, err
}

func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	cookie, _ := r.Cookie(manager.cookieName)
	sid, _ := url.QueryUnescape(cookie.Value)
	manager.driver.SessionDestroy(sid)

	expiration := time.Now()
	cookie = &http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
	http.SetCookie(w, cookie)
}

func (manager *Manager) SessionRead(w http.ResponseWriter, r *http.Request) (driver.Session, error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	cookie, _ := r.Cookie(manager.cookieName)
	sid, _ := url.QueryUnescape(cookie.Value)
	session, _ := manager.driver.SessionRead(sid)

	return session, nil
}

func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session driver.Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		sid := manager.generateSessionId()
		session, _ = manager.driver.SessionInit(sid)
		cookie := &http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(manager.maxLifeTime)}
		http.SetCookie(w, cookie)
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.driver.SessionRead(sid)
	}

	return
}

func (manager *Manager) CreateToken() string {
	hashMd5 := md5.New()
	salt := "hxia043&*%520"
	io.WriteString(hashMd5, salt+time.Now().String())
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

func NewManager(dirverName, cookieName string, maxLifeTime int64) (*Manager, error) {
	driver, ok := drivers[dirverName]
	if !ok {
		return nil, fmt.Errorf("session: unknown driver %q (forgotten import?)", dirverName)
	}

	return &Manager{driver: driver, cookieName: cookieName, maxLifeTime: maxLifeTime}, nil
}
