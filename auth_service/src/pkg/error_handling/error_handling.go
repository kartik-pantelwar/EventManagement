package errorhandling

import (
	pkgresponse "authservice/src/pkg/response"
	"net/http"
)

func HandleError(w http.ResponseWriter, msg string, statusCode int) {
	response := pkgresponse.StandardResponse{
		Status:  "FAILURE",
		Message: msg,
	}
	pkgresponse.WriteResponse(w, statusCode, response)

}
