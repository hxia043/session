# 0. Session Introduction
session is a tool example to provide the cookie-session interaction of web programming.  

The design of session is to simulate the design of [database/sql/driver](https://pkg.go.dev/database/sql/driver), which use the interface-oriented programming.  

There are several object need to be abstract to interface:
- session: session as an interface will provide the add/delete/update/get of data by the globally unique session id. The session object can be implemented by different driver entity.
- driver: driver is the interface will provide the add/delete/update/get of session, the session has been store and index by the object of driver.

With the two interfaces, developer can define the driver and session accordingly.  

For how to introduce the driver, here use global session manager to provide the service of session, which means the driver object need to register to Driver interface before manager startup.

Let's see the struct of object:
```
// global session manager
type Manager struct {
	driver      driver.Driver
	lock        sync.Mutex
	cookieName  string
	maxLifeTime int64
}

// memory driver has implement the interface of Driver
// and will store the session in memory
type Memory struct {
	lock     sync.Mutex
	sessions map[string]*list.Element
	list     list.List
}

type MemorySession struct {
	sid          string
	timeAccessed time.Time
	value        map[interface{}]interface{}
}
```

The relationship is clearly, startup from manager to memory and then the session, different object has implement the different function.

And the program struct is to simulate the database/sql/driver as well, which to define the interface with unify file `driver.go`, and to implement them in different package.  

The relation is same as database/sql, database/sql/driver, database/sql/sqlite3.

# 1. Session Implementation
## 1.1 session gc
A goroutine has provide the session of GC with a `time.AfterFunc` function, the function can implement the recurrence skillfully.

And there is no waiting defined to wait the close of goroutine cause the restful service doesn't exit, so that the goroutine has the same lifecycle with the main goroutine.

Compare the time is intersting, the point is [gc should remove the session which expire already](https://segmentfault.com/q/1010000011706064), based on the assume, can set the gclifetime as same as maxlifetime of cookie. When cookie expire in client, the session will be deleted by gc at the same time.

Only equal is not safety it should has gap between the expire time of cookie and gclifetime of session, for example delete the session in server, but the cookie haven't expire in client, and an request with cookie has been send to server at that time, the server will raise error to client even the cookie haven't expire.

Based on the proposal, when cookie request, the session will update the `timeAccessed` field with the current time, gc will remove the session according to the `timeAccessed`, then it can prevent the conflict.

Keep in mind the time of remove session by gc is decide as:
```
if (session.timeAccessed.Unix() + maxLifeTime) < time.Now().Unix() {
    fmt.Println("remove sid: ", session.sid)
    delete(memory.sessions, session.sid)
    memory.list.Remove(element)
} else {
    break
}
```

How to remove the session is implemented by the structure of list.  
The list has stored the session according to the time sequence.
When remove the session, gc only need to check the session `time.Accessed` in the end of list.

## 1.2 session hjjacking
To prevent the [session hijacking](https://github.com/astaxie/build-web-application-with-golang/blob/master/zh/06.4.md), here use a hidden field to import the token to the form of client.

When client sent request to server, server will verify the token and cookie(token saved in session of server) to check whether the cookie and request is valid.

Here implement a verify function which to verify the token and cookie, all route use the token and cookie will call the verify first. The function of verify is same as middleware.

## 1.3 use session
1) login without username
```
$ curl http://127.0.0.1:9091/login
sessionid=x1YBpj90F-qsMtX_pf8B8_XbwYSDsaGAhugeiBKaRMI%3D; Path=/; Max-Age=3600; HttpOnly
be49dad8705b97642d30b07644d04428
...
```

2) post username with session and token
```
$ curl -X POST http://localhost:9091/login -b 'sessionid=x1YBpj90F-qsMtX_pf8B8_XbwYSDsaGAhugeiBKaRMI%3D' -d 'username=hxia' -d 'token=be49dad8705b97642d30b07644d04428'
```

3) login with username
```
curl -b 'sessionid=x1YBpj90F-qsMtX_pf8B8_XbwYSDsaGAhugeiBKaRMI%3D' -d 'token=be49dad8705b97642d30b07644d0442
```

# 1. Next
- [ ] To learn the list for deep check the implement of container/list.  
- [ ] To learn the lock for check wich lock should be used specific.
