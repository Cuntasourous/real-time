package forum

import (
	"html/template"
	"net/http"
)

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {

	w.WriteHeader(status)

	var err error
	var t *template.Template

	if status == http.StatusInternalServerError {
		t, err = template.ParseFiles("templates/error.html")
	} else if status == http.StatusNotFound {
		t, err = template.ParseFiles("templates/notfound.html")
	} else if status == http.StatusBadRequest {
		t, err = template.ParseFiles("templates/badrequest.html")
	}

	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}


func NotFound(w http.ResponseWriter, r *http.Request) {
	ErrorHandler(w, r, http.StatusNotFound)
  }