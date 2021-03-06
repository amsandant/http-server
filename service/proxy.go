package service

import (
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func DoProxy(w http.ResponseWriter, r *http.Request, index int) {
	ip := pickIp(getIps(r))
	// Request remote
	client := &http.Client{
		Timeout: time.Duration(cacheConfig.Proxies[index].Timeout) * time.Second,
	}
	defer func(cli *http.Client) {
		cli.CloseIdleConnections()
	}(client)

	remoteUrl := cacheConfig.Proxies[index].Target + strings.Replace(r.RequestURI, cacheConfig.Proxies[index].Uri, "", 1)
	remoteRequest, _ := http.NewRequest(r.Method, remoteUrl, r.Body)
	remoteUrlPath := cacheConfig.Proxies[index].Target + remoteRequest.URL.Path

	// proxy header
	remoteRequest.Header = r.Header
	// forward
	handForward(r, remoteRequest, index)
	remoteResponse, err := client.Do(remoteRequest)
	if err == nil {
		defer func(res *http.Response) {
			_ = res.Body.Close()
		}(remoteResponse)
		// Set Headers
		remoteHeaders := remoteResponse.Header
		for key, value := range remoteHeaders {
			if len(value) == 1 {
				w.Header().Set(key, value[0])
			}
		}
		// Set status code
		remoteStatusCode := remoteResponse.StatusCode
		w.WriteHeader(remoteStatusCode)
		//Set body
		written, err := io.Copy(w, remoteResponse.Body)
		if err != nil {
			log.Println("[" + ip + "] " + r.Method + ": " + r.URL.Path + " -> " + remoteUrlPath + " : [" + strconv.FormatInt(written, 10) + "]" + err.Error())
		} else if cacheConfig.Debug {
			log.Println("[" + ip + "] " + r.Method + ": " + r.URL.Path + " -> " + remoteUrlPath + " : " + remoteResponse.Status)
		} else if remoteStatusCode != 200 {
			log.Println("[" + ip + "] " + r.Method + ": " + r.URL.Path + " -> " + remoteUrlPath + " : " + remoteResponse.Status)
		}
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		_, _ = w.Write([]byte(err.Error()))
		log.Println("[" + ip + "] " + r.Method + ": " + r.URL.Path + " -> " + remoteUrlPath + " : " + err.Error())
	}
}

func handForward(r *http.Request, tr *http.Request, index int) {
	if !cacheConfig.Proxies[index].Forward {
		return
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return
	}
	forward := r.Header.Get("x-forwarded-for")
	if isValidIp(forward) {
		forward += ", " + ip
	} else {
		forward = ip
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	tr.Header.Set("x-forwarded-for", forward)
	tr.Header.Set("X-Forwarded-Proto", scheme)
	tr.Header.Set("X-Forwarded-Scheme", scheme)
	tr.Header.Set("X-Forwarded-Host", host)
	tr.Header.Set("X-Forwarded-Port", port)
	tr.Host = host + ":" + port
}
