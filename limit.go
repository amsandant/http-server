package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type LimitInfo struct {
	LastTime int64
	Times    int
}

func (l *LimitInfo) Rest() {
	l.Times = 1
	l.LastTime = getCurrentTime()
}
func (l *LimitInfo) Increment() {
	l.Times++
}

var GLimitMap = make(map[string]*LimitInfo)

func LimitCheck(w http.ResponseWriter, r *http.Request) bool {
	if !GConfig.Limit.Enable {
		return true
	}
	ips := getIps(r)
	ip := pickIp(ips)
	info := &LimitInfo{}
	if _, ok := GLimitMap[ip]; !ok {
		GLimitMap[ip] = info
		GLimitMap[ip].Rest()
	} else {
		info = GLimitMap[ip]
	}
	if getCurrentTime()-info.LastTime > GConfig.Limit.Period {
		info.Rest()
	} else {
		info.Increment()
	}
	if !isWhiteIp(ip) && info.Times > GConfig.Limit.Times {
		w.WriteHeader(GConfig.Limit.StatusCode)
		_, _ = w.Write([]byte(GConfig.Limit.Message))
		log.Println(r.Method + ": " + r.URL.Path + " -> [" + ip + "]: " + GConfig.Limit.Message)
		return false
	}
	if GConfig.Limit.Delay > 0 {
		time.Sleep(time.Duration(GConfig.Limit.Delay) * time.Millisecond)
	}
	return true
}

func isWhiteIp(ip string) bool {
	if GConfig.Limit.WhiteIps == nil {
		return false
	}
	for _, item := range GConfig.Limit.WhiteIps {
		if strings.TrimSpace(item) == strings.TrimSpace(ip) {
			return true
		}
	}
	return false
}

func getCurrentTime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func getIps(r *http.Request) string {
	ip := r.Header.Get("x-forwarded-for")
	if isValidIp(ip) {
		return ip
	}
	ip = r.Header.Get("Proxy-Client-IP")
	if isValidIp(ip) {
		return ip
	}
	ip = r.Header.Get("WL-Proxy-Client-IP")
	if isValidIp(ip) {
		return ip
	}
	ip = r.Header.Get("HTTP_CLIENT_IP")
	if isValidIp(ip) {
		return ip
	}
	ip = r.Header.Get("HTTP_X_FORWARDED_FOR")
	if isValidIp(ip) {
		return ip
	}
	ip = r.Header.Get("HTTP_X_FORWARDED_FOR")
	if isValidIp(ip) {
		return ip
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return ip
	}
	return ""
}

func pickIp(ips string) string {
	if ips == "" {
		return ""
	}
	arr := strings.Split(ips, ",")
	ip := ""
	for _, str := range arr {
		if isValidIp(str) {
			if "0.0.0.0.0.0.0.1" == str || "0.0.0.0.0.0.0.1%0" == str || "0:0:0:0:0:0:0:1" == str || "::1" == str {
				ip = "127.0.0.1"
			} else {
				ip = strings.TrimSpace(str)
			}
		}
	}
	return ip
}

func isValidIp(ip string) bool {
	return ip != "" && strings.TrimSpace(strings.ToLower(ip)) != "unknown"
}
