package main

import (
	"errors"
	"fmt"
	"github.com/pipiBRH/rate"
	"net"
	"net/http"
	"strings"
	"time"
)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := FromRequest(r)

		fmt.Println(ip)
		if _, ok := Clients[ip]; !ok {
			Clients[ip] = rate.NewLimiter(60, time.Minute)
		}
		if Clients[ip].Allow() == false {
			http.Error(w, "Error", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var Clients map[string]*rate.LimitRate
var cidrs []*net.IPNet

func main() {
	Clients = make(map[string]*rate.LimitRate)
	mux := http.NewServeMux()
	mux.HandleFunc("/", okHandler)
	mux.HandleFunc("/hc", hcHandler)

	http.ListenAndServe(":4000", limit(mux))
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	ip := FromRequest(r)
	w.Write([]byte(fmt.Sprintln(Clients[ip].GetCount())))
}

func hcHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func FromRequest(r *http.Request) string {
	xRealIP := r.Header.Get("X-Real-Ip")
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	if xRealIP == "" && xForwardedFor == "" {
		var remoteIP string
		if strings.ContainsRune(r.RemoteAddr, ':') {
			remoteIP, _, _ = net.SplitHostPort(r.RemoteAddr)
		} else {
			remoteIP = r.RemoteAddr
		}

		return remoteIP
	}

	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		isPrivate, err := isPrivateAddress(address)
		if !isPrivate && err == nil {
			return address
		}
	}

	return xRealIP
}

func isPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, errors.New("address is not valid")
	}

	for i := range cidrs {
		if cidrs[i].Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}
