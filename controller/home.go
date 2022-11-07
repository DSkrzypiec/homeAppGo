package controller

import (
	"homeApp/front"
	"net/http"
)

func Home(w http.ResponseWriter, _ *http.Request) {
	tmpl := front.Home()
	tmpl.Execute(w, nil)
}
