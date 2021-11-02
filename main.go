package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"time"
)

var ruleConf=&rule{} // 全局初始化
var fileUP int64=0

func autoUpdateConf(){
	for{
		if err:=readRuleFile();err!=nil{
			log.Printf("获取配置文件的信息异常:%s\n",err.Error())
		}
		time.Sleep(10 * time.Second)
	}
}

func main() {
	go autoUpdateConf()
	setCA(caCert, caKey)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	doResponseRules(proxy)
    fmt.Println("启动在本机的8088端口的http/https代理；请将z.x509.cer安装为windows的根信任证书")
	log.Fatal(http.ListenAndServe(":8088", proxy))
}
