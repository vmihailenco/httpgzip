package httpgzip

import (
	"fmt"
	"mime"
	"net/http"
	"sync"

	"github.com/klauspost/compress/gzip"
)

type Config struct {
	minSize      int
	level        int
	contentTypes []parsedContentType

	pool sync.Pool
}

func New(opts ...ConfigOption) (*Config, error) {
	c := &Config{
		level:   gzip.DefaultCompression,
		minSize: DefaultMinSize,
	}

	for _, o := range opts {
		o(c)
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	c.pool = sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(nil, c.level)
			return w
		},
	}

	return c, nil
}

func (c *Config) AcceptsGzip(r *http.Request) bool {
	return acceptsGzip(r)
}

func (c *Config) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(vary, acceptEncoding)

		if !c.AcceptsGzip(r) {
			h.ServeHTTP(w, r)
			return
		}

		gw := c.ResponseWriter(w)
		defer gw.Close()

		h.ServeHTTP(gw, r)
	})
}

func (c *Config) ResponseWriter(w http.ResponseWriter) ResponseWriter {
	gw := &gzipResponseWriter{
		ResponseWriter: w,
		cfg:            c,
	}
	if _, ok := w.(http.CloseNotifier); ok {
		return &gzipResponseWriterWithCloseNotify{gw}
	}
	return gw
}

func (c *Config) validate() error {
	if c.level != gzip.DefaultCompression &&
		(c.level < gzip.BestSpeed || c.level > gzip.BestCompression) {
		return fmt.Errorf("invalid compression level requested: %d", c.level)
	}

	if c.minSize < 0 {
		return fmt.Errorf("minimum size must be more than zero")
	}

	return nil
}

type ConfigOption func(c *Config)

func MinSize(size int) ConfigOption {
	return func(c *Config) {
		c.minSize = size
	}
}

func CompressionLevel(level int) ConfigOption {
	return func(c *Config) {
		c.level = level
	}
}

// ContentTypes specifies a list of content types to compare
// the Content-Type header to before compressing. If none
// match, the response will be returned as-is.
//
// Content types are compared in a case-insensitive, whitespace-ignored
// manner.
//
// A MIME type without any other directive will match a content type
// that has the same MIME type, regardless of that content type's other
// directives. I.e., "text/html" will match both "text/html" and
// "text/html; charset=utf-8".
//
// A MIME type with any other directive will only match a content type
// that has the same MIME type and other directives. I.e.,
// "text/html; charset=utf-8" will only match "text/html; charset=utf-8".
//
// By default, responses are gzipped regardless of
// Content-Type.
func ContentTypes(types []string) ConfigOption {
	return func(c *Config) {
		c.contentTypes = nil
		for _, v := range types {
			mediaType, params, err := mime.ParseMediaType(v)
			if err == nil {
				c.contentTypes = append(c.contentTypes, parsedContentType{mediaType, params})
			}
		}
	}
}
