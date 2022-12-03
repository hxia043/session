package driver

var Drivers = make(map[string]Driver)

type Session interface {
	Get(key interface{}) interface{}
	Set(key, value interface{}) error
	Show()
}

type Driver interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionUpdate(sid string) (Session, error)
	SessionDestroy(sid string)
	SessionGC(maxLifeTime int64)
}
