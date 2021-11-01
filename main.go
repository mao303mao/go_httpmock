package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
)

func main() {
	setCA(caCert, caKey)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	doResponseRules(proxy)
    fmt.Println("启动在本机的8088端口的http/https代理；请将z.x509.cer安装为windows的根信任证书")
	log.Fatal(http.ListenAndServe(":8088", proxy))

}
