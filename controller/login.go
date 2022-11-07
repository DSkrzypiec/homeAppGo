package controller

import (
	"homeApp/front"
	"net/http"
)

func LoginForm(w http.ResponseWriter, _ *http.Request) {
	tmpl := front.Login()
	tmpl.Execute(w, nil)
}
