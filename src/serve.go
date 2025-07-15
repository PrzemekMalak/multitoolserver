package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var requestCount int = 0

func name() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

func ip() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func n(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "HostName: ", name())
}

func i(w http.ResponseWriter, r *http.Request) {
	i := ip()
	fmt.Fprintln(w, "IP Address: ", i)
}

func hello(w http.ResponseWriter, r *http.Request) {
	s := "HostName: " + name() + " IP Address: " + ip()
	var text string = os.Getenv("RETURN_TEXT")
	if text != "" {
		s = s + " " + text
	}
	fmt.Fprintln(w, s)
}

func e(w http.ResponseWriter, r *http.Request) {
	for _, e := range os.Environ() {
		fmt.Fprintln(w, e)
	}
}

func h(w http.ResponseWriter, r *http.Request) {
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Fprintln(w, name, value)
		}
	}
}
func r(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "-------------------")
	fmt.Fprintln(w, r.RemoteAddr)
}

func err2(w http.ResponseWriter, r *http.Request) {
	requestCount++
	if requestCount%2 == 0 {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		fmt.Fprintln(w, "OK")
	}
}

func err(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// sanitizePath validates and sanitizes the input path to prevent path traversal
func sanitizePath(inputPath string) (string, error) {
	// Clean the path to resolve any ".." or "." sequences
	cleanPath := filepath.Clean(inputPath)

	// Check for any remaining ".." sequences that might have been encoded
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}

	// Convert to absolute path for final validation
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	return absPath, nil
}

func ls(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query()
	var dir string
	if len(d["path"]) > 0 {
		dir = d["path"][0]
	}
	if len(dir) == 0 {
		dir = "/"
	}

	// Sanitize the path to prevent path traversal
	safePath, err := sanitizePath(dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid path: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "Directory:", safePath)
	files, err := os.ReadDir(safePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading directory: %v", err), http.StatusInternalServerError)
		return
	}

	for _, f := range files {
		fmt.Fprintln(w, f.Name())
	}
}

func req(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query()
	var url string
	if len(d["url"]) > 0 {
		url = d["url"][0]
	}
	if len(url) == 0 {
		url = "http://google.com"
	}

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := netClient.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(body))

}

func main() {
	http.HandleFunc("/req", req)     //sends request to external service and retuns its response; use ?url=
	http.HandleFunc("/source", r)    //returns source ip
	http.HandleFunc("/error", err)   //returns error 500
	http.HandleFunc("/error2", err2) //returns error 500 every second request
	http.HandleFunc("/host", n)      //returns hostname
	http.HandleFunc("/ip", i)        //returns ip address of the host
	http.HandleFunc("/env", e)       //returns env variables
	http.HandleFunc("/headers", h)   //returns headers
	http.HandleFunc("/hello", hello) //returns hostname, ip address and vslue of RETURN_TEXT env variable if available
	http.HandleFunc("/ls", ls)       //returns directory contents; use ?path=PATH to select a directory
	http.HandleFunc("/", hello)      //returns hostname, ip address and vslue of RETURN_TEXT env variable if available
	http.ListenAndServe(":8080", nil)
}
