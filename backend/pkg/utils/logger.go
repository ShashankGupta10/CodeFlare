package utils

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorYellow    = "\033[33m"
	colorBlue      = "\033[34m"
	colorPurple    = "\033[35m"
	colorCyan      = "\033[36m"
	colorGray      = "\033[37m"
	colorReset     = "\033[0m"
	colorLightCyan = "\033[96m"
	colorMagenta   = "\033[35m"
)

// Returns color ASNII for the specified http status code
func statusColor(code int) string {
	switch {
	case code >= 100 && code < 200:
		return colorYellow
	case code >= 200 && code < 300:
		return colorGreen
	case code >= 300 && code < 400:
		return colorBlue
	case code >= 400 && code < 500:
		return colorRed
	case code >= 500:
		return colorPurple
	default:
		return colorReset
	}
}

// Custom Middleware function for Pretty logging :).
func CustomLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			logMessage := fmt.Sprintf("%s[%s]%s %s%s%s%s%s %s%s%s %s%s%d%s%s %s%v%s %s",
				colorLightCyan, time.Now().Format("2006-01-02 15:04:05"), colorReset,
				"\033[1m", colorGray, req.Method, colorReset, "\033[0m",
				colorCyan, req.URL.Path, colorReset,
				"\033[1m", statusColor(res.Status), res.Status, colorReset, "\033[0m",
				colorGray, time.Since(start), colorReset,
				id,
			)

			fmt.Println(logMessage)

			return nil
		}
	}
}
