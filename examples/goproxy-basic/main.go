package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/elazarl/goproxy"
)

func main() {
	localIP := os.Getenv("LOCAL_IP")
	if localIP == "" {
		fmt.Println("export LOCAL_IP=your ip")
		return
	}
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	if os.Getenv("SHOW_LOG") == "true" {
		proxy.Verbose = true
	}
	log.Fatal(http.ListenAndServe(localIP, proxy))
}
