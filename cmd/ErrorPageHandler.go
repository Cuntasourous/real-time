package forum

import (
	"html/template"
	"net/http"
)

type ErrorData struct {
	Message    string
	StatusCode int
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	var data ErrorData

	switch status {
	case http.StatusNotFound:
		data = ErrorData{
			Message: "The page you are looking for does not exist.",
		}
	case http.StatusBadRequest:
		data = ErrorData{
			Message: "The request could not be understood by the server.",
		}
	case http.StatusInternalServerError:
		data = ErrorData{
			Message: "The server encountered an internal error.",
		}
	case http.StatusMethodNotAllowed:
		data = ErrorData{
			Message: "The method is not allowed for the requested URL.",
		}
	default:
		data = ErrorData{
			Message: "An unexpected error occurred.",
		}
	}
	parsedTemplate, err := template.ParseFiles("error/error.html")
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	data.StatusCode = status
	parsedTemplate.Execute(w, data)
}
