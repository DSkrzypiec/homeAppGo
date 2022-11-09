package controller

import (
	"homeApp/db"
	"homeApp/front"
	"net/http"
)

type Documents struct {
	DbClient *db.Client
}

type DocumentsList struct {
	Documents []db.DocumentInfo
}

func (d *Documents) DocumentsViewHandler(w http.ResponseWriter, r *http.Request) {
	documents, dErr := d.DbClient.Documents()
	if dErr != nil {
		// TODO
	}

	tmpl := front.Documents()
	execErr := tmpl.Execute(w, DocumentsList{Documents: documents})
	if execErr != nil {
		// TODO
	}
}
