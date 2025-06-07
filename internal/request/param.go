package request

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func GetParamUint64(r *http.Request, key string) (uint64, error) {
	return strconv.ParseUint(chi.URLParam(r, key), 10, 64)
}
