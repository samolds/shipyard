package httperror

import (
	"net/http"

	"github.com/zeebo/errs"
)

var (
	BadRequest      = errs.Class("bad request")     // 400
	Unauthenticated = errs.Class("unauthenticated") // 401
	Unauthorized    = errs.Class("unauthorized")    // 403
	NotFound        = errs.Class("not found")       // 404
	Unexpected      = errs.Class("internal")        // 500
)

func StatusCodeByError(err error) int {
	switch {
	case BadRequest.Has(err):
		return http.StatusBadRequest
	case Unauthenticated.Has(err):
		return http.StatusUnauthorized // not a typo
	case Unauthorized.Has(err):
		return http.StatusForbidden
	case NotFound.Has(err):
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
