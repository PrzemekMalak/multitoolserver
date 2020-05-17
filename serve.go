package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
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

func ls(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query()
	var dir string
	if len(d["path"]) > 0 {
		dir = d["path"][0]
	}
	if len(dir) == 0 {
		dir = "/"
	}
	fmt.Fprintln(w, dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(w, "")
	}

	for _, f := range files {
		fmt.Fprintln(w, f.Name())
	}
}

func main() {
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
