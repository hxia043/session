package main

import (
	"fmt"
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
	if verify(w, r) {
		globalSessionManager.SessionDestroy(w, r)
		w.Write([]byte(w.Header().Values("Set-Cookie")[0]))
	} else {
		w.Write([]byte("verify token failed, please register again"))
	}
}

func update(w http.ResponseWriter, r *http.Request) {
	if verify(w, r) {
		session, err := globalSessionManager.SeesionUpdate(w, r)
		if err != nil {
			panic(err)
		}

		session.Show()
	} else {
		w.Write([]byte("verify token failed, please register again"))
	}
}

func verify(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	if token := r.Form.Get("token"); token != "" {
		session, _ := globalSessionManager.SessionRead(w, r)
		if session.Get("token") != token {
			fmt.Println("verify token failed")
			return false
		}
		return true
	}
	return false
}

func login(w http.ResponseWriter, r *http.Request) {
	session := globalSessionManager.SessionStart(w, r)
	r.ParseForm()
	if r.Method == "GET" {
		template, _ := template.ParseFiles("../html/login.go.tpl")
		w.Header().Set("Content-Type", "text/html")

		if value, ok := session.Get("username").([]string); ok {
			template.Execute(w, value[0])
		} else {
			template.Execute(w, session.Get("username"))
		}

		token := globalSessionManager.CreateToken()
		session.Set("token", token)
		template.Execute(w, token)
	} else {
		if verify(w, r) {
			session.Set("username", r.Form["username"])
			w.Write([]byte(w.Header().Values("Set-Cookie")[0]))
		} else {
			w.Write([]byte("verify token failed, please register again"))
		}
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
