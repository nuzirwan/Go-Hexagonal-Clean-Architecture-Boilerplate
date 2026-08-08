package httputil

import (
	"errors"
	"net/http"

	domainerror "github.com/nuzirwan/go-boilerplate/internal/domain/error"
)

func DomainError(w http.ResponseWriter, r *http.Request, err error) {
	var de *domainerror.DomainError
	if errors.As(err, &de) {
		status := mapDomainCodeToHTTP(de.Code)
		writeJSON(w, status, Response{
			Status: "error",
			Error:  &ErrorBody{Code: string(de.Code), Message: de.Message, Field: de.Field},
			Meta:   buildMeta(r),
		})
		return
	}
	Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
}

func mapDomainCodeToHTTP(code domainerror.Code) int {
	switch code {
	case domainerror.ErrValidation:
		return http.StatusBadRequest
	case domainerror.ErrNotFound:
		return http.StatusNotFound
	case domainerror.ErrConflict:
		return http.StatusConflict
	case domainerror.ErrBusinessRule:
		return http.StatusUnprocessableEntity
	case domainerror.ErrUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
