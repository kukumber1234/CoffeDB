package handler

import (
	"encoding/json"
	"frapo/config"
	flag "frapo/models"
	"net/http"
)

func SendResponse(message string, err error, status int, w http.ResponseWriter) {
	config.Logger.Error(message, err)

	Response := flag.Error{
		Message: message,
		Status:  int64(status),
	}

	response, err := json.Marshal(Response)
	if err != nil {
		config.Logger.Error("Failed to send response", err)
		http.Error(w, `{"Error": "Internal Server Error", "Status": 500}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}
