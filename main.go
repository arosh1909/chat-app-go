package main

import (
	"log"
	"net/http"
	"sync"
	"text/template"
	"path/filepath"
	"flag"
	"os"
	"github.com/arosh1909/Chat/trace"
	"github.com/stretchr/objx"
	"github.com/stretchr/signature"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	
)

type templateHandler struct {
	once sync.Once
	filename string
	templ *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "the addr of the application.")
	flag.Parse()
	
	gomniauth.SetSecurityKey(signature.RandomKey(64))
	gomniauth.WithProviders(google.New("997502716232-gq1k71gpkl99jtdp0lehkasuute1n88g.apps.googleusercontent.com",
							"gbDTgYICeW7PBigIiIFyo6CZ",
							"http://localhost:8080/auth/callback/google" ))
	r:=newRoom()
	r.tracer = trace.New(os.Stdout)
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request){
		http.SetCookie(w, &http.Cookie{
			Name: "auth",
			Value: "",
			Path: "/",
			MaxAge: -1,
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	go r.run()
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:",err)
	}
}