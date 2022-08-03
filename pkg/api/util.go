package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GeneralResponse defines general response struct.
type GeneralResponse struct {
	Message string `json:"message"`
}

func writeOKResponse(w http.ResponseWriter, resp interface{}) {
	WriteResponse(w, http.StatusOK, resp)
}

func writeSuccessResponse(w http.ResponseWriter) {
	WriteResponse(w, http.StatusOK, GeneralResponse{Message: "success"})
}

func writeErrorResponse(w http.ResponseWriter, err error) {
	WriteResponse(w, http.StatusInternalServerError, GeneralResponse{Message: err.Error()})
}

func writeBadRequestResponse(w http.ResponseWriter, err error) {
	WriteResponse(w, http.StatusBadRequest, GeneralResponse{Message: err.Error()})
}

// WriteResponse writes code and resonse to http writer
func WriteResponse(w http.ResponseWriter, code int, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("{message: failed to encode resonse: %v}", err), http.StatusInternalServerError)
	}
}
