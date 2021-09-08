package middleware

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webpserver/app"
	"webpserver/config"
	"webpserver/util"
)

var (
	httpReqs      *prometheus.CounterVec
	httpDurations *prometheus.SummaryVec
)

func init() {

	httpReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "How many HTTP requests served, labeled by status code, HTTP method, cache state and content type.",
		},
		[]string{"code", "method", "cache", "contenttype"},
	)
	prometheus.MustRegister(httpReqs)

	httpDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_durations_seconds",
			Help:       "HTTP latency distributions.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"code", "method", "cache", "contenttype"},
	)
	prometheus.MustRegister(httpDurations)
}

// LogMiddleware Logs incoming requests, including response status.
func LogMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		o := &responseObserver{ResponseWriter: w}
		h.ServeHTTP(o, r)
		addr := r.RemoteAddr
		if i := strings.LastIndex(addr, ":"); i != -1 {
			addr = addr[:i]
		}

		ip := util.GetRemoteIPAdress(r)
		duration := time.Since(start)

		// metrics
		httpReqs.WithLabelValues(strconv.Itoa(o.status), r.Method, o.ResponseWriter.Header().Get("Cache"),
			strings.Split(o.ResponseWriter.Header().Get("Content-type"), ";")[0]).Inc()
		httpDurations.WithLabelValues(strconv.Itoa(o.status), r.Method, o.ResponseWriter.Header().Get("Cache"),
			strings.Split(o.ResponseWriter.Header().Get("Content-type"), ";")[0]).Observe(float64(duration) / 1e9)

		// logs
		switch o.status / 100 {
		case 1, 2, 3, 4:
			app.Log.Info("HTTP response", "code", o.status, "proto", r.Proto, "method", r.Method, "url", r.URL,
				"time", strconv.FormatFloat(float64(duration)/1e9, 'f', -1, 64), "time_human", duration.Round(10*time.Microsecond), "size", o.written,
				"ua", r.UserAgent(), "ip", ip)
		default:
			app.Log.Warn("HTTP response", "code", o.status, "proto", r.Proto, "method", r.Method, "url", r.URL,
				"time", strconv.FormatFloat(float64(duration)/1e9, 'f', -1, 64), "time_human", duration.Round(10*time.Microsecond), "size", o.written,
				"ua", r.UserAgent(), "ip", ip)
		}
	})
}

type responseObserver struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (o *responseObserver) Write(p []byte) (n int, err error) {
	//if !o.wroteHeader {
	//	o.WriteHeader(http.StatusOK)
	//}
	n, err = o.ResponseWriter.Write(p)
	o.written += int64(n)
	return
}

func (o *responseObserver) WriteHeader(code int) {
	if o.ResponseWriter.Header().Get("Server") == "" {
		o.ResponseWriter.Header().Add("Server", config.Cfg.ServerName)
	}
	if strings.Contains(o.ResponseWriter.Header().Get("Content-type"), "image/webp") {
		o.ResponseWriter.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", config.Cfg.WebPMaxAge))
	}

	o.ResponseWriter.WriteHeader(code)
	if o.wroteHeader {
		return
	}
	o.wroteHeader = true
	o.status = code
}
