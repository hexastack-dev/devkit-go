package requestlog

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hexastack-dev/devkit-go/log"
)

var DefaultFilteredHeaders = []string{"Authorization", "Cookie"}

type ReqLogger struct {
	log             log.Logger
	filteredHeaders map[string]bool
}

func NewReqLogger(logger log.Logger, filteredHeaders []string) *ReqLogger {
	if len(filteredHeaders) == 0 {
		filteredHeaders = DefaultFilteredHeaders
	}
	filter := make(map[string]bool)
	for _, v := range filteredHeaders {
		filter[v] = true
	}
	return &ReqLogger{log: logger, filteredHeaders: filter}
}

func (s *ReqLogger) Log(ent *Entry) {
	msg := fmt.Sprintf("%d %s %s://%s%s %s",
		ent.Status,
		ent.Request.Method,
		getScheme(ent.Request),
		ent.Request.Host,
		ent.Request.URL.RequestURI(),
		ent.Request.Proto,
	)
	var r struct {
		Request struct {
			Timestamp time.Time         `json:"timestamp"`
			Size      int64             `json:"size"`
			Method    string            `json:"method"`
			URL       string            `json:"url"`
			Query     map[string]string `json:"query"`
			Header    map[string]string `json:"header"`
		} `json:"request"`
		Response struct {
			StatusCode int   `json:"statusCode"`
			Size       int64 `json:"size"`
		} `json:"response"`
	}

	r.Request.Timestamp = ent.ReceivedTime
	r.Request.Method = ent.Request.Method
	r.Request.URL = ent.Request.URL.RequestURI()
	r.Request.Size = ent.RequestBodySize + headerSize(ent.Request.Header)
	r.Request.Query = joinMultiValuesMap(ent.Request.URL.Query())
	r.Request.Header = joinMultiValuesMap(filterHeaders(ent.Request.Header, s.filteredHeaders))

	r.Response.StatusCode = ent.Status
	r.Response.Size = ent.ResponseHeaderSize + ent.ResponseBodySize

	s.log.WithContext(ent.Request.Context()).Info(msg,
		log.Field("status", ent.Status),
		log.Field("elapsedTime", ent.Latency),
		log.Field("http", r),
	)
}

func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func joinMultiValuesMap(m map[string][]string) map[string]string {
	m1 := make(map[string]string)
	for k, v := range m {
		m1[k] = strings.Join(v, "; ")
	}

	return m1
}

func filterHeaders(h http.Header, filter map[string]bool) http.Header {
	h1 := h.Clone()
	for k, v := range h1 {
		if _, ok := filter[k]; ok {
			for i := range v {
				h1[k][i] = "**REDACTED**"
			}
		}
	}

	return h1
}
