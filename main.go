package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var ruleConf = &rule{} // 全局初始化
var fileUP int64=0

// 上游代理
type upstreamProxy struct {
	ProxyActive bool `json:"proxyActive"`
	ProxyUrl string `json:"proxyUrl"`
	ProxyUser string `json:"proxyUser"`
	ProxyPassword string `json:"proxyPassword"`
}
// 上游代理的代理认证
const ProxyAuthHeader="Proxy-Authorization"
func basicAuth(username string,password string) string{
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}
func setBasicAuth(username string,password string, req *http.Request){
	req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
}

// 定时更新规则
func autoUpdateConf(){
	tc:=time.NewTicker(10*time.Second)
	for{
		if err:=readRuleFile();err!=nil{
			log.Printf("获取配置文件的信息异常:%s\n",err.Error())
		}
		<-tc.C
	}
}

//设置上游http代理
func setUpstreamProxy(server * goproxy.ProxyHttpServer){
	upProxy :=&upstreamProxy{}
	confFile,err:=os.Open("./upstreamProxyConfig.json")
	if err!=nil{
		log.Printf("读取上行代理配置文件失败(%s)，采用无代理方式\n",err.Error())
		return
	}
	defer confFile.Close()
	if err=json.NewDecoder(confFile).Decode(upProxy);err!=nil{
		log.Printf("读取上行代理配置内容失败(%s)，采用无代理方式\n",err.Error())
		return
	}
	if !upProxy.ProxyActive{
		return
	}
	server.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		return url.Parse(upProxy.ProxyUrl)
	}
	log.Printf("已使用上行代理(%s)\n",upProxy.ProxyUrl)
	if strings.TrimSpace(upProxy.ProxyUser)=="" || strings.TrimSpace(upProxy.ProxyPassword)==""{
		server.ConnectDial = server.NewConnectDialToProxy(upProxy.ProxyUrl)
	}else{
		server.ConnectDial = server.NewConnectDialToProxyWithHandler(upProxy.ProxyUrl,func(req *http.Request) {
			setBasicAuth(upProxy.ProxyUser,upProxy.ProxyPassword,req)
		})
		server.OnRequest().Do(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			setBasicAuth(upProxy.ProxyUser,upProxy.ProxyPassword,req)
			return req, nil
		}))
	}
}
// func allowh2c(next http.Handler) http.Handler {
// 	h2server := &http2.Server{IdleTimeout: time.Second * 60}
// 	return h2c.NewHandler(next, h2server)
// }

func main() {
	go autoUpdateConf()
	setCA(caCert, caKey)
	proxy := goproxy.NewProxyHttpServer()
	nonproxyHandler:=http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			infoHandler(w, r)
		case "/cert":
			certDownloadHandler(w, r)
		default:
			http.Error(w, "Unsupported path ", http.StatusNotFound)
		}
	})
	proxy.NonproxyHandler=nonproxyHandler //覆盖原有的非代理处理
	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	doResponseRules(proxy) // mock响应处理
	setUpstreamProxy(proxy) // 设置上行代理处理
    log.Println("启动在本机的8088端口的http/https代理")
	log.Println("可浏览器访问8088的网页端，下载z.x509.cer并安装为windows的根信任证书")
	log.Fatal(http.ListenAndServe(":8088",proxy))
}
