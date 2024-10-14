package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type ErrorData struct {
	Message    string
	StatusCode int
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
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
	case http.StatusConflict:
		data = ErrorData{
			Message: "Conflict",
		}
	default:
		data = ErrorData{
			Message: "An unexpected error occurred.",
		}
	}
	data.StatusCode = status

	parsedTemplate, err := template.ParseFiles("error/error.html")
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	err = parsedTemplate.Execute(w, data)
	if err != nil {
		// Log the error for debugging purposes
		log.Printf("Error executing error.html template")
		// Set the response status code
	}
	fmt.Fprintf(w, "Error: %d - Something went wrong.", status)
}
