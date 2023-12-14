package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type traceIDKeyType int

const (
	traceIDKey    traceIDKeyType = iota
	TraceIDHeader                = "X-B3-TraceId"

	bufferLimit int = 512
)

type RequestLog struct{}

// RequestLogger returns a logger handler using a custom LogFormatter.
//
// Example:
//
//	if err := log.Setup("warn"); err != nil {
//	    slog.Error("unable to err)
//	}
//
//	r := chi.NewRouter()
//	r.Use(middleware.RequestLogger)
func RequestLogger(next http.Handler) http.Handler {
	var l *RequestLog

	fn := func(w http.ResponseWriter, r *http.Request) {
		var (
			entry = l.NewLogEntry(r)
			buf   = newLimitBuffer(bufferLimit)

			ww = middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		)

		ww.Tee(buf)

		t1 := time.Now()
		defer func() {
			var body []byte
			if ww.Status() >= http.StatusBadRequest {
				body, _ = io.ReadAll(buf)
			}

			entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t1), body)
		}()

		next.ServeHTTP(ww, middleware.WithLogEntry(r, entry))
	}

	return http.HandlerFunc(fn)
}

func (*RequestLog) NewLogEntry(r *http.Request) middleware.LogEntry {
	var (
		traceID   = sanitize(r.Header.Get(TraceIDHeader))
		referer   = sanitize(r.Referer())
		userAgent = sanitize(r.UserAgent())
	)

	if traceID == "" {
		traceID = uuid.NewString()
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(slog.With("trace_id", traceID))

	// adds the trace id into the request context.
	ctx := context.WithValue(r.Context(), traceIDKey, traceID)
	r = r.WithContext(ctx)

	logger.LogAttrs(r.Context(), slog.LevelInfo, "request started",
		slog.String("method", r.Method),
		slog.String("host", r.Host),
		slog.String("request", r.RequestURI),
		slog.String("remote-addr", r.RemoteAddr),
		slog.String("referer", referer),
		slog.String("user-agent", userAgent),
	)

	return &RequestLogEntry{}
}

type RequestLogEntry struct{}

func (*RequestLogEntry) Write(
	status int,
	bts int,
	header http.Header,
	elapsed time.Duration,
	extra any,
) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if status >= http.StatusBadRequest && logger.Enabled(context.Background(), slog.LevelDebug) {
		var attrs []any

		if body, ok := extra.([]byte); ok {
			attrs = append(attrs, slog.Attr{
				Key:   "body",
				Value: slog.StringValue(string(body)),
			})
		}

		if len(header) > 0 {
			attrs = append(attrs, slog.Group("header", attrsToAnys(headerLogField(header))...))
		}

		logger = logger.With(slog.Group("http_response", attrs...))
	}

	logger.Info(
		"request completed",
		slog.Int("status", status),
		slog.Int("bytes", bts),
		slog.Int64("elapsed_ms", elapsed.Milliseconds()),
		slog.Int64("elapsed_ns", elapsed.Nanoseconds()),
	)
}

func (*RequestLogEntry) Panic(v any, stack []byte) {
	slog.Error("something went wrong",
		slog.String("stack", string(stack)),
		slog.String("panic", sanitize(v.(string))),
	)
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, "\n", "")

	return strings.ReplaceAll(s, "\r", "")
}

func headerLogField(header http.Header) []slog.Attr {
	headerField := make([]slog.Attr, 0, len(header))

	for k, v := range header {
		k = strings.ToLower(k)

		switch {
		case len(v) == 0:
			continue
		case len(v) == 1:
			headerField = append(headerField, slog.Attr{
				Key:   k,
				Value: slog.StringValue(v[0]),
			})
		default:
			headerField = append(headerField, slog.Attr{
				Key:   k,
				Value: slog.StringValue(fmt.Sprintf("[%s]", strings.Join(v, "], ["))),
			})
		}

		if k == "authorization" || k == "cookie" || k == "set-cookie" {
			headerField[len(headerField)] = slog.Attr{
				Key:   k,
				Value: slog.StringValue("***"),
			}
		}
	}

	return headerField
}

func attrsToAnys(attr []slog.Attr) []any {
	attrs := make([]any, len(attr))
	for i, a := range attr {
		attrs[i] = a
	}
	return attrs
}

// limitBuffer is used to pipe response body information from the
// response writer to a certain limit amount. The idea is to read
// a portion of the response body such as an error response so we
// may log it.
type limitBuffer struct {
	*bytes.Buffer
	limit int
}

func newLimitBuffer(size int) io.ReadWriter {
	return limitBuffer{
		Buffer: bytes.NewBuffer(make([]byte, 0, size)),
		limit:  size,
	}
}

func (b limitBuffer) Write(p []byte) (n int, err error) {
	if b.Buffer.Len() >= b.limit {
		return len(p), nil
	}
	limit := b.limit
	if len(p) < limit {
		limit = len(p)
	}
	return b.Buffer.Write(p[:limit])
}

func (b limitBuffer) Read(p []byte) (n int, err error) {
	return b.Buffer.Read(p)
}
