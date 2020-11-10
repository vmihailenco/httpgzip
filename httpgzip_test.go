package httpgzip

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/require"
)

const (
	smallTestBody = "aaabbcaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbc"
	testBody      = "aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc"
)

func TestParseEncodings(t *testing.T) {
	examples := map[string]codings{

		// Examples from RFC 2616
		"compress, gzip":                     {"compress": 1.0, "gzip": 1.0},
		"":                                   {},
		"*":                                  {"*": 1.0},
		"compress;q=0.5, gzip;q=1.0":         {"compress": 0.5, "gzip": 1.0},
		"gzip;q=1.0, identity; q=0.5, *;q=0": {"gzip": 1.0, "identity": 0.5, "*": 0.0},

		// More random stuff
		"AAA;q=1":     {"aaa": 1.0},
		"BBB ; q = 2": {"bbb": 1.0},
	}

	for eg, exp := range examples {
		act, _ := parseEncodings(eg)
		require.Equal(t, exp, act)
	}
}

func TestGzipHandler(t *testing.T) {
	// This just exists to provide something for GzipHandler to wrap.
	handler := newTestHandler(testBody)

	// requests without accept-encoding are passed along as-is

	req1, _ := http.NewRequest("GET", "/whatever", nil)
	resp1 := httptest.NewRecorder()
	handler.ServeHTTP(resp1, req1)
	res1 := resp1.Result()

	require.Equal(t, 200, res1.StatusCode)
	require.Equal(t, "", res1.Header.Get("Content-Encoding"))
	require.Equal(t, "Accept-Encoding", res1.Header.Get("Vary"))
	require.Equal(t, testBody, resp1.Body.String())

	// but requests with accept-encoding:gzip are compressed if possible

	req2, _ := http.NewRequest("GET", "/whatever", nil)
	req2.Header.Set("Accept-Encoding", "gzip")
	resp2 := httptest.NewRecorder()
	handler.ServeHTTP(resp2, req2)
	res2 := resp2.Result()

	require.Equal(t, 200, res2.StatusCode)
	require.Equal(t, "gzip", res2.Header.Get("Content-Encoding"))
	require.Equal(t, "Accept-Encoding", res2.Header.Get("Vary"))
	require.Equal(t, gzipStrLevel(testBody, gzip.DefaultCompression), resp2.Body.Bytes())

	// content-type header is correctly set based on uncompressed body

	req3, _ := http.NewRequest("GET", "/whatever", nil)
	req3.Header.Set("Accept-Encoding", "gzip")
	res3 := httptest.NewRecorder()
	handler.ServeHTTP(res3, req3)

	require.Equal(t, http.DetectContentType([]byte(testBody)), res3.Header().Get("Content-Type"))
}

func TestGzipHandlerSmallBodyNoCompression(t *testing.T) {
	handler := newTestHandler(smallTestBody)

	req, _ := http.NewRequest("GET", "/whatever", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	res := resp.Result()

	// with less than 1400 bytes the response should not be gzipped

	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, "", res.Header.Get("Content-Encoding"))
	require.Equal(t, "Accept-Encoding", res.Header.Get("Vary"))
	require.Equal(t, smallTestBody, resp.Body.String())
}

func TestGzipHandlerAlreadyCompressed(t *testing.T) {
	handler := newTestHandler(testBody)

	req, _ := http.NewRequest("GET", "/gzipped", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	require.Equal(t, testBody, res.Body.String())
}

func TestNewGzipLevelHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, testBody)
	})

	for lvl := gzip.BestSpeed; lvl <= gzip.BestCompression; lvl++ {
		wrapper, err := NewGzipLevelHandler(lvl)
		require.Nil(t, err, "NewGzipLevleHandler returned error for level:", lvl)

		req, _ := http.NewRequest("GET", "/whatever", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		resp := httptest.NewRecorder()
		wrapper(handler).ServeHTTP(resp, req)
		res := resp.Result()

		require.Equal(t, 200, res.StatusCode)
		require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
		require.Equal(t, "Accept-Encoding", res.Header.Get("Vary"))
		require.Equal(t, gzipStrLevel(testBody, lvl), resp.Body.Bytes())
	}
}

func TestNewGzipLevelHandlerReturnsErrorForInvalidLevels(t *testing.T) {
	var err error
	_, err = NewGzipLevelHandler(-42)
	require.NotNil(t, err)

	_, err = NewGzipLevelHandler(42)
	require.NotNil(t, err)
}

func TestMustNewGzipLevelHandlerWillPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("panic was not called")
		}
	}()

	_ = MustNewGzipLevelHandler(-42)
}

func TestGzipHandlerNoBody(t *testing.T) {
	tests := []struct {
		statusCode      int
		contentEncoding string
		emptyBody       bool
		body            []byte
	}{
		// Body must be empty.
		{http.StatusNoContent, "", true, nil},
		{http.StatusNotModified, "", true, nil},
		// Body is going to get gzip'd no matter what.
		{http.StatusOK, "", true, []byte{}},
		{http.StatusOK, "gzip", false, []byte(testBody)},
	}

	for num, test := range tests {
		handler := GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(test.statusCode)
			if test.body != nil {
				w.Write(test.body)
			}
		}))

		rec := httptest.NewRecorder()
		// TODO: in Go1.7 httptest.NewRequest was introduced this should be used
		// once 1.6 is not longer supported.
		req := &http.Request{
			Method:     "GET",
			URL:        &url.URL{Path: "/"},
			Proto:      "HTTP/1.1",
			ProtoMinor: 1,
			RemoteAddr: "192.0.2.1:1234",
			Header:     make(http.Header),
		}
		req.Header.Set("Accept-Encoding", "gzip")
		handler.ServeHTTP(rec, req)

		body, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("Unexpected error reading response body: %v", err)
		}

		header := rec.Header()
		require.Equal(t, test.contentEncoding, header.Get("Content-Encoding"), fmt.Sprintf("for test iteration %d", num))
		require.Equal(t, "Accept-Encoding", header.Get("Vary"), fmt.Sprintf("for test iteration %d", num))
		if test.emptyBody {
			require.Empty(t, body, fmt.Sprintf("for test iteration %d", num))
		} else {
			require.NotEmpty(t, body, fmt.Sprintf("for test iteration %d", num))
			require.NotEqual(t, test.body, body, fmt.Sprintf("for test iteration %d", num))
		}
	}
}

func TestGzipHandlerContentLength(t *testing.T) {
	testBodyBytes := []byte(testBody)
	tests := []struct {
		bodyLen   int
		bodies    [][]byte
		emptyBody bool
	}{
		{len(testBody), [][]byte{testBodyBytes}, false},
		// each of these writes is less than the DefaultMinSize
		{len(testBody), [][]byte{testBodyBytes[:200], testBodyBytes[200:]}, false},
		// without a defined Content-Length it should still gzip
		{0, [][]byte{testBodyBytes[:200], testBodyBytes[200:]}, false},
		// simulate a HEAD request with an empty write (to populate headers)
		{len(testBody), [][]byte{nil}, true},
	}

	// httptest.NewRecorder doesn't give you access to the Content-Length
	// header so instead, we create a server on a random port and make
	// a request to that instead
	ln, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("failed creating listen socket: %v", err)
	}
	defer ln.Close()
	srv := &http.Server{
		Handler: nil,
	}
	go srv.Serve(ln)

	for num, test := range tests {
		srv.Handler = GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if test.bodyLen > 0 {
				w.Header().Set("Content-Length", strconv.Itoa(test.bodyLen))
			}
			for _, b := range test.bodies {
				w.Write(b)
			}
		}))
		req := &http.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/", Scheme: "http", Host: ln.Addr().String()},
			Header: make(http.Header),
			Close:  true,
		}
		req.Header.Set("Accept-Encoding", "gzip")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Unexpected error making http request in test iteration %d: %v", num, err)
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Unexpected error reading response body in test iteration %d: %v", num, err)
		}

		l, err := strconv.Atoi(res.Header.Get("Content-Length"))
		if err != nil {
			t.Fatalf("Unexpected error parsing Content-Length in test iteration %d: %v", num, err)
		}
		if test.emptyBody {
			require.Empty(t, body, fmt.Sprintf("for test iteration %d", num))
			require.Equal(t, 0, l, fmt.Sprintf("for test iteration %d", num))
		} else {
			require.Len(t, body, l, fmt.Sprintf("for test iteration %d", num))
		}
		require.Equal(t, "gzip", res.Header.Get("Content-Encoding"), fmt.Sprintf("for test iteration %d", num))
		require.NotEqual(t, test.bodyLen, l, fmt.Sprintf("for test iteration %d", num))
	}
}

func TestGzipHandlerMinSizeMustBePositive(t *testing.T) {
	_, err := NewGzipLevelAndMinSize(gzip.DefaultCompression, -1)
	require.Error(t, err)
}

func TestGzipHandlerMinSize(t *testing.T) {
	responseLength := 0
	b := []byte{'x'}

	wrapper, _ := NewGzipLevelAndMinSize(gzip.DefaultCompression, 128)
	handler := wrapper(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Write responses one byte at a time to ensure that the flush
			// mechanism, if used, is working properly.
			for i := 0; i < responseLength; i++ {
				n, err := w.Write(b)
				require.Equal(t, 1, n)
				require.Nil(t, err)
			}
		},
	))

	r, _ := http.NewRequest("GET", "/whatever", &bytes.Buffer{})
	r.Header.Add("Accept-Encoding", "gzip")

	// Short response is not compressed
	responseLength = 127
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Result().Header.Get(contentEncoding) == "gzip" {
		t.Error("Expected uncompressed response, got compressed")
	}

	// Long response is not compressed
	responseLength = 128
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Result().Header.Get(contentEncoding) != "gzip" {
		t.Error("Expected compressed response, got uncompressed")
	}
}

func TestGzipDoubleClose(t *testing.T) {
	c, err := New()
	require.Nil(t, err)

	handler := c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// call close here and it'll get called again interally by
		// NewGzipLevelHandler's handler defer
		w.Write([]byte("test"))
		w.(io.Closer).Close()
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	// the second close shouldn't have added the same writer
	// so we pull out 2 writers from the pool and make sure they're different
	w1 := c.pool.Get()
	w2 := c.pool.Get()
	// require.NotEqual looks at the value and not the address, so we use regular ==
	require.False(t, w1 == w2)
}

type panicOnSecondWriteHeaderWriter struct {
	http.ResponseWriter
	headerWritten bool
}

func (w *panicOnSecondWriteHeaderWriter) WriteHeader(s int) {
	if w.headerWritten {
		panic("header already written")
	}
	w.headerWritten = true
	w.ResponseWriter.WriteHeader(s)
}

func TestGzipHandlerDoubleWriteHeader(t *testing.T) {
	handler := GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "15000")
		// Specifically write the header here
		w.WriteHeader(304)
		// Ensure that after a Write the header isn't triggered again on close
		w.Write(nil)
	}))
	wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w = &panicOnSecondWriteHeaderWriter{
			ResponseWriter: w,
		}
		handler.ServeHTTP(w, r)
	})

	rec := httptest.NewRecorder()
	// TODO: in Go1.7 httptest.NewRequest was introduced this should be used
	// once 1.6 is not longer supported.
	req := &http.Request{
		Method:     "GET",
		URL:        &url.URL{Path: "/"},
		Proto:      "HTTP/1.1",
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
		Header:     make(http.Header),
	}
	req.Header.Set("Accept-Encoding", "gzip")
	wrapper.ServeHTTP(rec, req)
	body, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Unexpected error reading response body: %v", err)
	}
	require.Empty(t, body)
	header := rec.Header()
	require.Equal(t, "gzip", header.Get("Content-Encoding"))
	require.Equal(t, "Accept-Encoding", header.Get("Vary"))
	require.Equal(t, 304, rec.Code)
}

func TestStatusCodes(t *testing.T) {
	handler := GzipHandler(http.NotFoundHandler())
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	result := w.Result()
	if result.StatusCode != 404 {
		t.Errorf("StatusCode should have been 404 but was %d", result.StatusCode)
	}
}

func TestFlushBeforeWrite(t *testing.T) {
	b := []byte(testBody)
	handler := GzipHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.(http.Flusher).Flush()
		rw.Write(b)
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	res := w.Result()
	require.Equal(t, http.StatusNotFound, res.StatusCode)
	require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
	require.NotEqual(t, b, w.Body.Bytes())
}

func TestImplementCloseNotifier(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(acceptEncoding, "gzip")
	GzipHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, ok := rw.(http.CloseNotifier)
		require.True(t, ok, "response writer must implement http.CloseNotifier")
	})).ServeHTTP(&mockRWCloseNotify{}, request)
}

func TestImplementFlusherAndCloseNotifier(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(acceptEncoding, "gzip")
	GzipHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, okCloseNotifier := rw.(http.CloseNotifier)
		require.True(t, okCloseNotifier, "response writer must implement http.CloseNotifier")
		_, okFlusher := rw.(http.Flusher)
		require.True(t, okFlusher, "response writer must implement http.Flusher")
	})).ServeHTTP(&mockRWCloseNotify{}, request)
}

func TestNotImplementCloseNotifier(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(acceptEncoding, "gzip")
	GzipHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, ok := rw.(http.CloseNotifier)
		require.False(t, ok, "response writer must not implement http.CloseNotifier")
	})).ServeHTTP(httptest.NewRecorder(), request)
}

type mockRWCloseNotify struct{}

func (m *mockRWCloseNotify) CloseNotify() <-chan bool {
	panic("implement me")
}

func (m *mockRWCloseNotify) Header() http.Header {
	return http.Header{}
}

func (m *mockRWCloseNotify) Write([]byte) (int, error) {
	panic("implement me")
}

func (m *mockRWCloseNotify) WriteHeader(int) {
	panic("implement me")
}

func TestIgnoreSubsequentWriteHeader(t *testing.T) {
	handler := GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.WriteHeader(404)
	}))
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	result := w.Result()
	if result.StatusCode != 500 {
		t.Errorf("StatusCode should have been 500 but was %d", result.StatusCode)
	}
}

func TestDontWriteWhenNotWrittenTo(t *testing.T) {
	// When using gzip as middleware without ANY writes in the handler,
	// ensure the gzip middleware doesn't touch the actual ResponseWriter
	// either.

	handler0 := GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler0.ServeHTTP(w, r)
		w.WriteHeader(404) // this only works if gzip didn't do a WriteHeader(200)
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	handler1.ServeHTTP(w, r)

	result := w.Result()
	if result.StatusCode != 404 {
		t.Errorf("StatusCode should have been 404 but was %d", result.StatusCode)
	}
}

var contentTypeTests = []struct {
	name                 string
	contentType          string
	acceptedContentTypes []string
	expectedGzip         bool
}{
	{
		name:                 "Always gzip when content types are empty",
		contentType:          "",
		acceptedContentTypes: []string{},
		expectedGzip:         true,
	},
	{
		name:                 "MIME match",
		contentType:          "application/json",
		acceptedContentTypes: []string{"application/json"},
		expectedGzip:         true,
	},
	{
		name:                 "MIME no match",
		contentType:          "text/xml",
		acceptedContentTypes: []string{"application/json"},
		expectedGzip:         false,
	},
	{
		name:                 "MIME match with no other directive ignores non-MIME directives",
		contentType:          "application/json; charset=utf-8",
		acceptedContentTypes: []string{"application/json"},
		expectedGzip:         true,
	},
	{
		name:                 "MIME match with other directives requires all directives be equal, different charset",
		contentType:          "application/json; charset=ascii",
		acceptedContentTypes: []string{"application/json; charset=utf-8"},
		expectedGzip:         false,
	},
	{
		name:                 "MIME match with other directives requires all directives be equal, same charset",
		contentType:          "application/json; charset=utf-8",
		acceptedContentTypes: []string{"application/json; charset=utf-8"},
		expectedGzip:         true,
	},
	{
		name:                 "MIME match with other directives requires all directives be equal, missing charset",
		contentType:          "application/json",
		acceptedContentTypes: []string{"application/json; charset=ascii"},
		expectedGzip:         false,
	},
	{
		name:                 "MIME match case insensitive",
		contentType:          "Application/Json",
		acceptedContentTypes: []string{"application/json"},
		expectedGzip:         true,
	},
	{
		name:                 "MIME match ignore whitespace",
		contentType:          "application/json;charset=utf-8",
		acceptedContentTypes: []string{"application/json;            charset=utf-8"},
		expectedGzip:         true,
	},
}

func TestContentTypes(t *testing.T) {
	for _, tt := range contentTypeTests {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", tt.contentType)
			io.WriteString(w, testBody)
		})

		wrapper, err := GzipHandlerWithOpts(ContentTypes(tt.acceptedContentTypes))
		require.Nil(t, err, "NewGzipHandlerWithOpts returned error", tt.name)

		req, _ := http.NewRequest("GET", "/whatever", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		resp := httptest.NewRecorder()
		wrapper(handler).ServeHTTP(resp, req)
		res := resp.Result()

		require.Equal(t, 200, res.StatusCode)
		if tt.expectedGzip {
			require.Equal(t, "gzip", res.Header.Get("Content-Encoding"), tt.name)
		} else {
			require.NotEqual(t, "gzip", res.Header.Get("Content-Encoding"), tt.name)
		}
	}
}

// --------------------------------------------------------------------

func BenchmarkGzipHandler_S2k(b *testing.B)   { benchmark(b, false, 2048) }
func BenchmarkGzipHandler_S20k(b *testing.B)  { benchmark(b, false, 20480) }
func BenchmarkGzipHandler_S100k(b *testing.B) { benchmark(b, false, 102400) }
func BenchmarkGzipHandler_P2k(b *testing.B)   { benchmark(b, true, 2048) }
func BenchmarkGzipHandler_P20k(b *testing.B)  { benchmark(b, true, 20480) }
func BenchmarkGzipHandler_P100k(b *testing.B) { benchmark(b, true, 102400) }

// --------------------------------------------------------------------

func gzipStrLevel(s string, lvl int) []byte {
	var b bytes.Buffer
	w, _ := gzip.NewWriterLevel(&b, lvl)
	io.WriteString(w, s)
	w.Close()
	return b.Bytes()
}

func benchmark(b *testing.B, parallel bool, size int) {
	bin, err := ioutil.ReadFile("testdata/benchmark.json")
	if err != nil {
		b.Fatal(err)
	}

	req, _ := http.NewRequest("GET", "/whatever", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	handler := newTestHandler(string(bin[:size]))

	if parallel {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				runBenchmark(b, req, handler)
			}
		})
	} else {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			runBenchmark(b, req, handler)
		}
	}
}

func runBenchmark(b *testing.B, req *http.Request, handler http.Handler) {
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if code := res.Code; code != 200 {
		b.Fatalf("Expected 200 but got %d", code)
	} else if blen := res.Body.Len(); blen < 500 {
		b.Fatalf("Expected complete response body, but got %d bytes", blen)
	}
}

func newTestHandler(body string) http.Handler {
	return GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gzipped":
			w.Header().Set("Content-Encoding", "gzip")
			io.WriteString(w, body)
		default:
			io.WriteString(w, body)
		}
	}))
}