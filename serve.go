package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

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
		return " "
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

func ls(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query()
	dir := d["path"]
	fmt.Println(dir)
	files, err := ioutil.ReadDir(dir[0])
	if err != nil {
		fmt.Println(err)
	}

	for _, f := range files {
		fmt.Fprintln(w, f.Name())
	}
}

func main() {
	http.HandleFunc("/host", n)
	http.HandleFunc("/ip", i)
	http.HandleFunc("/env", e)
	http.HandleFunc("/headers", h)
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/ls", ls)
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8080", nil)
}
