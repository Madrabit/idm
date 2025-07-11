package web

import (
	"bytes"
	"github.com/gofiber/fiber/v3"
	"idm/inner/common"
	"net/http"
	"runtime/debug"
	"strings"
)

func HTTPHandler(h http.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
			}
		}()
		bodyReader := bytes.NewReader(c.Body())
		req, err := http.NewRequest(string(c.Method()), c.OriginalURL(), bodyReader)
		if err != nil {
			return err
		}
		req.RequestURI = c.OriginalURL()
		req.Header = make(http.Header)
		for k, values := range c.GetReqHeaders() {
			for _, v := range values {
				req.Header.Add(k, v)
			}
		}
		rw := &responseWriter{ctx: c}
		h.ServeHTTP(rw, req)
		return nil
	}
}

type responseWriter struct {
	ctx    fiber.Ctx
	logger common.Logger
}

func (rw *responseWriter) Header() http.Header {
	h, ok := rw.ctx.Locals("headers").(http.Header)
	if !ok {
		h = make(http.Header)
		rw.ctx.Locals("headers", h)
	}
	return h
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	headers := rw.Header()
	ct := parseMimeOnly(headers.Get("Content-Type"))
	if ct != "" {
		rw.ctx.Set("Content-Type", ct)
	} else {
		extType := guessMimeFromPath(rw.ctx.OriginalURL())
		rw.ctx.Set("Content-Type", extType)
	}
	return rw.ctx.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	headers := rw.Header()
	for k, vv := range headers {
		for _, v := range vv {
			rw.ctx.Append(k, v)
		}
	}
	if ct := parseMimeOnly(headers.Get("Content-Type")); ct != "" {
		rw.ctx.Set("Content-Type", ct)
	}
	rw.ctx.Status(statusCode)
}

// отрезает параметры типа charset
func parseMimeOnly(ct string) string {
	if idx := strings.Index(ct, ";"); idx != -1 {
		return strings.TrimSpace(ct[:idx])
	}
	return strings.TrimSpace(ct)
}

func guessMimeFromPath(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript"
	case strings.HasSuffix(path, ".css"):
		return "text/css"
	case strings.HasSuffix(path, ".json"):
		return "application/json"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	default:
		return "text/plain"
	}
}
