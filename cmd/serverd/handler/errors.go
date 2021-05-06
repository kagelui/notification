package handler

import (
	"net/http"

	"github.com/kagelui/notification/internal/pkg/web"
)

var (
	errParsingRequest = &web.Error{
		Status: http.StatusBadRequest,
		Code:   "400 Bad Request",
		Desc:   "cannot parse request",
	}
)
