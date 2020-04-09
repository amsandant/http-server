package main

import (
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
)

const configFile = "conf.json"

func main() {
	readConfig()
	lPort := ":" + strconv.Itoa(GConfig.Port)
	dir, _ := os.Getwd()
	GConfig.Static.Dir = strings.TrimSpace(GConfig.Static.Dir)
	if GConfig.Static.Dir != "" {
		dir = GConfig.Static.Dir
	}
	mux := http.NewServeMux()
	fileHandler := http.FileServer(http.Dir(dir))

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		proxyIndex := isProxy(request.URL.Path)
		if proxyIndex >= 0 {
			if LimitCheck(writer, request) {
				doProxy(writer, request, proxyIndex)
			}
			return
		}
		if GConfig.Static.History && request.URL.Path != "/" {
			filePath := strings.ReplaceAll(dir+request.URL.Path, "/", string(filepath.Separator))
			if !isExist(filePath) {
				request.URL.Path = "/"
				//http.Redirect(writer, request, "/", http.StatusFound)
				//return
			}
		}
		fileHandler.ServeHTTP(writer, request)
	})

	absDir, _ := filepath.Abs(dir)
	log.Println("ListenAndServe " + lPort + " -> " + absDir)

	for i, proxy := range GConfig.Proxies {
		log.Println("ListenAndProxy[" + strconv.Itoa(i) + "]: " + proxy.Uri + " -> " + proxy.Target)
	}

	if runtime.GOOS == "windows" {
		_ = exec.Command("cmd", "/c", "start", "http://localhost"+lPort).Start()
	}

	err := http.ListenAndServe(lPort, mux)
	if err != nil {
		log.Fatal("ListenAndServe " + lPort + " -> " + err.Error())
	}
}

func readConfig() {
	GConfig = &Config{
		Port: 18080,
		Static: StaticItem{
			Dir:     "",
			History: false,
		},
		Proxies: nil,
		Limit: LimitItem{
			Period:     10000,
			Times:      20,
			Enable:     false,
			StatusCode: 403,
			Message:    "403 Forbidden",
		},
	}
	fromFile := false
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		configBytes, err = ioutil.ReadFile(getExecPath() + configFile)
	}
	if err == nil {
		err = json.Unmarshal(configBytes, GConfig)
		if err == nil {
			fromFile = true
		}
	}

	if !fromFile {
		GConfig.Proxies = []ProxyItem{{
			Uri:     "/proxy",
			Target:  "",
			Forward: false,
		}}
		flag.IntVar(&GConfig.Port, "port", GConfig.Port, "Port")
		flag.BoolVar(&GConfig.Debug, "debug", GConfig.Debug, "Debug")
		flag.StringVar(&GConfig.Static.Dir, "static.dir", GConfig.Static.Dir, "StaticItem Directory")
		flag.BoolVar(&GConfig.Static.History, "static.history", GConfig.Static.History, "History router")
		flag.StringVar(&GConfig.Proxies[0].Uri, "proxy.uri", GConfig.Proxies[0].Uri, "ProxyItem uri")
		flag.StringVar(&GConfig.Proxies[0].Target, "proxy.target", GConfig.Proxies[0].Target, "ProxyItem target")
		flag.BoolVar(&GConfig.Proxies[0].Forward, "proxy.forward", GConfig.Proxies[0].Forward, "ProxyItem forward")
		flag.Parse()
	}
	proxies := make([]ProxyItem, 0)
	for _, proxy := range GConfig.Proxies {
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
	GConfig.Proxies = proxies
}

func isProxy(url string) int {
	for index, proxy := range GConfig.Proxies {
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
