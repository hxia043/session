package main

import (
	_ "github/hxia043/session/internal/memory"
	"github/hxia043/session/internal/session"
	"net/http"
	"text/template"
)

var (
	globalSessionManager *session.Manager
	err                  error
)

func logout(w http.ResponseWriter, r *http.Request) {
	globalSessionManager.SessionDestroy(w, r)
	w.Write([]byte(w.Header().Values("Set-Cookie")[0]))
}

func update(w http.ResponseWriter, r *http.Request) {
	session, err := globalSessionManager.SeesionUpdate(w, r)
	if err != nil {
		panic(err)
	}

	session.Show()
}

func login(w http.ResponseWriter, r *http.Request) {
	session := globalSessionManager.SessionStart(w, r)
	r.ParseForm()
	if r.Method == "GET" {
		template, _ := template.ParseFiles("../html/login.go.tpl")
		w.Header().Set("Content-Type", "text/html")

		session.Show()
		if value, ok := session.Get("username").([]string); ok {
			template.Execute(w, value[0])
		} else {
			template.Execute(w, session.Get("username"))
		}
	} else {
		session.Set("username", r.Form["username"])
		w.Write([]byte(w.Header().Values("Set-Cookie")[0]))
	}
}

func main() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/update", update)
	http.HandleFunc("/logout", logout)

	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		panic(err)
	}
}

func init() {
	globalSessionManager, err = session.NewManager("memory", "sessionid", 30)
	if err != nil {
		panic(err.Error())
	}

	go globalSessionManager.GC()
}
