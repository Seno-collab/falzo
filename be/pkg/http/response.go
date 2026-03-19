package http

import "net/http"

type ResponseStatus struct {
	Data any `json:data`
}

func BuildResponseSuccess(data any, w http.ResponseWriter) {
	w.Write()
}
