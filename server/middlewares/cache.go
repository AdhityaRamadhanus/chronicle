package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AdhityaRamadhanus/chronicle"
	"github.com/spf13/viper"
)

type cachedResponseWriter struct {
	http.ResponseWriter
	status       int
	CacheService chronicle.CacheService
	Key          string
	Exp          time.Duration
}

func newcachedResponseWriter(res http.ResponseWriter) *cachedResponseWriter {
	return &cachedResponseWriter{ResponseWriter: res}
}

func (crw *cachedResponseWriter) Status() int {
	return crw.status
}

func (crw *cachedResponseWriter) Write(resBody []byte) (n int, err error) {
	if crw.Status() == 200 {
		crw.CacheService.SetEx(crw.Key, resBody, crw.Exp)
	}

	return crw.ResponseWriter.Write(resBody)
}

func (crw *cachedResponseWriter) WriteHeader(code int) {
	crw.ResponseWriter.WriteHeader(code)
	crw.status = code
}

func buildCacheKeyFromURI(req *http.Request) string {
	cacheKeyParts := []string{
		"chronicle",
		"http-cache",
	}

	paths := strings.Split(req.URL.Path, "/")
	for _, path := range paths {
		cacheKeyParts = append(cacheKeyParts, path)
	}

	// prevent redis key attack
	knownQueryStrings := []string{
		"page",
		"limit",
		"sort-by",
		"order",
		//filter
		"status",
		"topic",
	}

	querystring := req.URL.Query()
	for _, knownQueryString := range knownQueryStrings {
		keyPart := fmt.Sprintf("%s=%s", knownQueryString, querystring.Get(knownQueryString))
		cacheKeyParts = append(cacheKeyParts, keyPart)
	}

	return strings.Join(cacheKeyParts, ":")
}

func Cache(cacheService chronicle.CacheService) func(string, http.HandlerFunc) http.HandlerFunc {
	return func(duration string, next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if !viper.GetBool("cache_response") {
				next(res, req)
				return
			}

			existingCache, err := cacheService.Get(req.RequestURI)
			if err != nil {
				crw := newcachedResponseWriter(res)
				crw.CacheService = cacheService
				crw.Key = buildCacheKeyFromURI(req)
				crw.Exp, _ = time.ParseDuration(duration)

				next(crw, req)
				return
			}

			//only cache json
			res.Header().Set("Content-Type", "application/json; charset=utf-8")
			res.WriteHeader(http.StatusOK)
			res.Write(existingCache)
			return
		})
	}
}
