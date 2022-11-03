package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

const SessCookieName = "sessioncookie"

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.Handle("/counters", checkAuth(http.HandlerFunc(countersHandler)))
	http.Handle("/secret", checkAuth(http.HandlerFunc(secretHandler)))

	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_, cookieErr := r.Cookie(SessCookieName)
	var tmpl *template.Template

	if cookieErr != nil { // no session cookie
		tmpl = template.Must(template.ParseFiles("html/login.html", "html/common/header.html"))
	} else {
		tmpl = template.Must(template.ParseFiles("html/home.html", "html/common/header.html", "html/common/menu.html"))
	}
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.FormValue("login")
	pass := r.FormValue("pass")

	// TODO: Make actual authentication
	if strings.ToLower(name) == "damian" && strings.ToLower(pass) == "crap" {
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

func secretHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Hello Damian! You have access to the secret!")
}

func countersHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFiles("html/counters.html", "html/common/header.html", "html/common/menu.html"))
	tmpl.Execute(w, nil)
}

func checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessCookie, cookieErr := r.Cookie(SessCookieName)
		if cookieErr != nil {
			fmt.Println("There is no cookie, redirecting to login...")
			http.Redirect(w, r, "http://localhost:8080/", http.StatusSeeOther)
			return
		}
		if sessCookie.Value != "xxx-1234" {
			fmt.Println("Icorrect credentials, redirecting to login...")
			http.Redirect(w, r, "http://localhost:8080/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
