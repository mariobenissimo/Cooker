package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"gitub.com/mariobenissimo/apiGateway/internal/limiter"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	if limiter.GlobalLimiter.Allow() {
		// request accepted
		// check if it's possible forward request to server
		proxy := httputil.NewSingleHostReverseProxy(parseTarget("http://servizio1:8088/"))
		proxy.ServeHTTP(w, r)
	} else {
		// request limited, return HTTP 429 Too Many Requests
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
	}
}

// Parse destination url
func parseTarget(target string) *url.URL {
	url, err := url.Parse(target)
	if err != nil {
		panic("Impossibile analizzare l'URL di destinazione")
	}
	fmt.Println(url)
	return url
}
