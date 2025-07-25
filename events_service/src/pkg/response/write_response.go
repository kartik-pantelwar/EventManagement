package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type StandardResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   error       `json:"error,omitempty"`
}

func WriteResponse(w http.ResponseWriter, statuscode int, resp StandardResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statuscode)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	} else {
		// ^Terminal Logging
		fmt.Print("\n")
		fmt.Println(resp.Message)
	}
}

// WriteSuccess writes a successful response
func WriteSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	resp := StandardResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
	WriteResponse(w, statusCode, resp)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	resp := StandardResponse{
		Status:  "error",
		Message: message,
	}
	WriteResponse(w, statusCode, resp)
}
