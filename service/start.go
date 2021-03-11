package service

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const configFile = "conf.json"
const IdleTimeout = 10 * time.Second

func Start() {
	readConfig()
	if cacheConfig.Debug {
		infoBytes, _ := json.Marshal(cacheConfig)
		var str bytes.Buffer
		_ = json.Indent(&str, infoBytes, "", "  ")
		log.Printf("%s\n%s", "conf.json", str.String())
	}

	lPort := ":" + strconv.Itoa(cacheConfig.Port)
	dir, _ := os.Getwd()
	cacheConfig.Static.Dir = strings.TrimSpace(cacheConfig.Static.Dir)
	if cacheConfig.Static.Dir != "" {
		dir = cacheConfig.Static.Dir
	}
	mux := http.NewServeMux()
	fileHandler := http.FileServer(http.Dir(dir))

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		proxyIndex := isProxy(request.URL.Path)
		if proxyIndex >= 0 {
			if LimitCheck(writer, request) {
				DoProxy(writer, request, proxyIndex)
			}
			return
		}
		if cacheConfig.Static.History && request.URL.Path != "/" {
			filePath := strings.ReplaceAll(dir+request.URL.Path, "/", string(filepath.Separator))
			if !isExist(filePath) {
				//request.URL.Path = "/"
				writeHistory(dir, writer)
				return
			}
		}
		fileHandler.ServeHTTP(writer, request)
	})

	absDir, _ := filepath.Abs(dir)
	log.Println("ListenAndServe " + lPort + " -> " + absDir)

	for i, proxy := range cacheConfig.Proxies {
		log.Println("ListenAndProxy[" + strconv.Itoa(i) + "]: " + proxy.Uri + " -> " + proxy.Target)
	}
	if cacheConfig.CrtFile != "" && cacheConfig.KeyFile != "" {
		if runtime.GOOS == "windows" {
			_ = exec.Command("cmd", "/c", "start", "https://localhost"+lPort).Start()
		}
		server := http.Server{
			Addr:        lPort,
			Handler:     mux,
			IdleTimeout: IdleTimeout,
		}
		err := server.ListenAndServeTLS(cacheConfig.CrtFile, cacheConfig.KeyFile)
		if err != nil {
			log.Fatal("ListenAndServe " + lPort + " -> " + err.Error())
		}
	} else {
		if runtime.GOOS == "windows" {
			_ = exec.Command("cmd", "/c", "start", "http://localhost"+lPort).Start()
		}
		server := http.Server{
			Addr:        lPort,
			Handler:     mux,
			IdleTimeout: IdleTimeout,
		}
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe " + lPort + " -> " + err.Error())
		}
	}

}

func writeHistory(dir string, writer http.ResponseWriter) {
	indexPath := dir + "/index.html"
	indexFile, err := os.Open(indexPath)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}
	indexContent, err := ioutil.ReadAll(indexFile)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}
	writer.WriteHeader(http.StatusNotFound)
	_, _ = writer.Write(indexContent)
	_ = indexFile.Close()
}

func readConfig() {
	config := NewDefaultConfig()
	cacheConfig = &config
	flag.IntVar(&cacheConfig.Port, "port", cacheConfig.Port, "Port")
	flag.BoolVar(&cacheConfig.Debug, "debug", cacheConfig.Debug, "Debug")
	flag.StringVar(&cacheConfig.Static.Dir, "static.dir", cacheConfig.Static.Dir, "StaticItem Directory")
	flag.BoolVar(&cacheConfig.Static.History, "static.history", cacheConfig.Static.History, "History router")
	flag.StringVar(&cacheConfig.Proxies[0].Uri, "proxy.uri", cacheConfig.Proxies[0].Uri, "ProxyItem uri")
	flag.StringVar(&cacheConfig.Proxies[0].Target, "proxy.target", cacheConfig.Proxies[0].Target, "ProxyItem target")
	flag.BoolVar(&cacheConfig.Proxies[0].Forward, "proxy.forward", cacheConfig.Proxies[0].Forward, "ProxyItem forward")
	flag.IntVar(&cacheConfig.Proxies[0].Timeout, "proxy.timeout", cacheConfig.Proxies[0].Timeout, "代理超时时间,默认0永不超时")

	flag.Parse()

	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		configBytes, err = ioutil.ReadFile(getExecPath() + configFile)
	}
	if err == nil {
		_ = json.Unmarshal(configBytes, cacheConfig)
	}

	proxies := make([]ProxyItem, 0)
	for _, proxy := range cacheConfig.Proxies {
		if strings.TrimSpace(proxy.Uri) == "" || strings.TrimSpace(proxy.Target) == "" {
			continue
		}
		if proxy.Target[len(proxy.Target)-1:] == "/" {
			proxy.Target = proxy.Target[:len(proxy.Target)-1]
		}
		if proxy.Uri[len(proxy.Uri)-1:] == "/" {
			proxy.Uri = proxy.Uri[:len(proxy.Uri)-1]
		}
		proxies = append(proxies, proxy)
	}
	cacheConfig.Proxies = proxies
}

func isProxy(url string) int {
	for index, proxy := range cacheConfig.Proxies {
		if url == proxy.Uri {
			return index
		}
		reg, _ := regexp.Compile(proxy.Uri + "/.*")
		if reg.Match([]byte(url)) {
			return index
		}
	}
	return -1
}

func isExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func getExecPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[0 : index+1]
}
