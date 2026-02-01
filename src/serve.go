package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var requestCount int = 0
var component string

func init() {
	component = os.Getenv("COMPONENT")
	if component == "" {
		component = "component0"
	}
}

func name() string {
	name, err := os.Hostname()
	if err != nil {
		log.Printf("Error getting hostname: %v", err)
		return "unknown"
	}
	return name
}

func ip() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Error getting interface addresses: %v", err)
		return "unknown"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	log.Printf("No valid IPv4 address found")
	return "unknown"
}

func n(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintln(w, "HostName: ", name()); err != nil {
		log.Printf("Error writing hostname response: %v", err)
		return
	}
}

func i(w http.ResponseWriter, r *http.Request) {
	i := ip()
	if _, err := fmt.Fprintln(w, "IP Address: ", i); err != nil {
		log.Printf("Error writing IP address response: %v", err)
		return
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	hostname := name()
	ipAddr := ip()

	s := "HostName: " + hostname + " IP Address: " + ipAddr
	var text string = os.Getenv("RETURN_TEXT")
	if text != "" {
		s = s + " " + text
	}

	log.Printf("Hello request from %s", r.RemoteAddr)
	if _, err := fmt.Fprintln(w, s); err != nil {
		log.Printf("Error writing hello response: %v", err)
		return
	}
}

func e(w http.ResponseWriter, r *http.Request) {
	for _, e := range os.Environ() {
		if _, err := fmt.Fprintln(w, e); err != nil {
			log.Printf("Error writing environment variable: %v", err)
			return
		}
	}
}

func h(w http.ResponseWriter, r *http.Request) {
	for name, values := range r.Header {
		for _, value := range values {
			if _, err := fmt.Fprintln(w, name, value); err != nil {
				log.Printf("Error writing header: %v", err)
				return
			}
		}
	}
}
func r(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintln(w, "-------------------"); err != nil {
		log.Printf("Error writing separator: %v", err)
		return
	}
	if _, err := fmt.Fprintln(w, r.RemoteAddr); err != nil {
		log.Printf("Error writing remote address: %v", err)
		return
	}
}

func err2(w http.ResponseWriter, r *http.Request) {
	requestCount++
	if requestCount%2 == 0 {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		if _, err := fmt.Fprintln(w, "OK"); err != nil {
			log.Printf("Error writing OK response: %v", err)
			return
		}
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
		log.Printf("Path traversal attempt detected: %s -> %v", dir, err)
		http.Error(w, fmt.Sprintf("Invalid path: %v", err), http.StatusBadRequest)
		return
	}

	// Check if directory exists and is readable before writing any response
	files, err := os.ReadDir(safePath)
	if err != nil {
		log.Printf("Error reading directory %s: %v", safePath, err)
		http.Error(w, fmt.Sprintf("Error reading directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Now write the response
	if _, err := fmt.Fprintln(w, "Directory:", safePath); err != nil {
		log.Printf("Error writing directory path: %v", err)
		return
	}

	for _, f := range files {
		if _, err := fmt.Fprintln(w, f.Name()); err != nil {
			log.Printf("Error writing file name: %v", err)
			return
		}
	}

	log.Printf("Directory listing successful for: %s (%d files)", safePath, len(files))
}

func req(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query()
	var urlStr string
	if len(d["url"]) > 0 {
		urlStr = d["url"][0]
	}
	if len(urlStr) == 0 {
		urlStr = "http://google.com"
	}

	// Parse and validate the URL
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		log.Printf("Invalid URL provided: %s, error: %v", urlStr, err)
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Optionally: Only allow http/https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		log.Printf("Unsupported URL scheme: %s", parsedURL.Scheme)
		http.Error(w, "Only http and https URLs are allowed", http.StatusBadRequest)
		return
	}

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := netClient.Get(urlStr)
	if err != nil {
		log.Printf("Error making HTTP request to %s: %v", urlStr, err)
		http.Error(w, fmt.Sprintf("Failed to make request: %v", err), http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body from %s: %v", urlStr, err)
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}

	// Set appropriate content type based on response
	if resp.Header.Get("Content-Type") != "" {
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	// Copy status code from the response
	w.WriteHeader(resp.StatusCode)

	fmt.Fprint(w, string(body))
}

func main() {
	log.Printf("Starting Multitool Server on port 8080")

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

	log.Printf("Server ready to accept connections")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
