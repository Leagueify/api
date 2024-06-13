package util

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func SendStatus(code int, context echo.Context, detail string) error {
	if detail != "" {
		return context.JSON(code,
			map[string]string{
				"status": strings.ToLower(http.StatusText(code)),
				"detail": detail,
			},
		)
	}
	return context.JSON(code,
		map[string]string{
			"status": strings.ToLower(http.StatusText(code)),
		},
	)
}
