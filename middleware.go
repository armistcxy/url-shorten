package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

type Middleware func(http.Handler) http.Handler

func ApplyChain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

/*
Fields in `http.Request`: https://pkg.go.dev/net/http#Request
*/

func HTTPLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// only log if status code >= 400
		if metrics.Code < 400 {
			return
		}

		httpInfo := &HTTPInfo{
			method:          r.Method,
			proto:           r.Proto,
			userAgent:       r.UserAgent(),
			requestBodySize: r.ContentLength,
			referer:         r.Referer(),
			ipaddr:          realip.FromRequest(r), // https://github.com/tomasen/realip
			code:            metrics.Code,
			responseSize:    metrics.Written,
			duration:        metrics.Duration,
		}

		logHTTPInfo(httpInfo)
	})
}

type HTTPInfo struct {
	method          string
	proto           string
	userAgent       string
	referer         string
	requestBodySize int64
	ipaddr          string
	code            int
	responseSize    int64
	duration        time.Duration
}

func logHTTPInfo(info *HTTPInfo) {
	var strBuilder strings.Builder

	// Append all fields to the string builder
	strBuilder.WriteString(fmt.Sprintf("\nmethod: %s\n", info.method))
	strBuilder.WriteString(fmt.Sprintf("proto: %s\n", info.proto))
	strBuilder.WriteString(fmt.Sprintf("userAgent: %s\n", info.userAgent))
	strBuilder.WriteString(fmt.Sprintf("referer: %s\n", info.referer))
	strBuilder.WriteString(fmt.Sprintf("requestBodySize: %d bytes\n", info.requestBodySize))
	strBuilder.WriteString(fmt.Sprintf("ipaddr: %s\n", info.ipaddr))
	strBuilder.WriteString(fmt.Sprintf("code: %d\n", info.code))
	strBuilder.WriteString(fmt.Sprintf("responseSize: %d bytes\n", info.responseSize))
	strBuilder.WriteString(fmt.Sprintf("duration: %v\n", info.duration))

	log.Println(strBuilder.String())
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(1000), 1000)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, ErrReachLimitRequest.Error(), http.StatusTooManyRequests)
			return
		}
	})
}

var ErrReachLimitRequest = errors.New("reach the number of maximum requests")
