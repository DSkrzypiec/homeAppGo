package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"homeApp/front"
)

const SessCookieName = "sessioncookie"

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.Handle("/counters", checkAuth(http.HandlerFunc(countersHandler)))

	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_, cookieErr := r.Cookie(SessCookieName)
	var tmpl *template.Template

	if cookieErr != nil { // no session cookie
		tmpl = front.Login()
	} else {
		tmpl = front.Home()
	}
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.FormValue("login")
	pass := r.FormValue("pass")

	// TODO: Make actual authentication
	if strings.ToLower(name) == "damian" && strings.ToLower(pass) == "crap" {
		fmt.Printf("Login from %v\n", r.RemoteAddr)
		// Cookie must be set before anything is written onto ResponseWriter!
		sessCookie := &http.Cookie{
			Name:     SessCookieName,
			Value:    "xxx-1234",
			Expires:  time.Now().Add(10 * time.Minute),
			HttpOnly: true,
		}
		http.SetCookie(w, sessCookie)
	} else {
		fmt.Println("I don't know you!")
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func countersHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := front.Counters()
	tmpl.Execute(w, nil)
}

func checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessCookie, cookieErr := r.Cookie(SessCookieName)
		if cookieErr != nil {
			fmt.Println("There is no cookie, redirecting to login...")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if sessCookie.Value != "xxx-1234" {
			fmt.Println("Icorrect credentials, redirecting to login...")
			sessCookie := &http.Cookie{
				Name:     SessCookieName,
				Value:    "",
				Expires:  time.Unix(0, 0),
				HttpOnly: true,
			}
			http.SetCookie(w, sessCookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
