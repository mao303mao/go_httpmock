package main

import (
	"bytes"
	"fmt"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

const (
	ContentTypeJson = "application/json;charset=UTF-8"
	ContentTypeText = "text/plain;charset=UTF-8"
	ContentTypeHtml = "text/html;charset=UTF-8"
	ContentTypeJpeg = "image/jpeg"
)

func doResponseRules(proxy *goproxy.ProxyHttpServer){ // response add cors headers
	proxy.OnResponse().DoFunc(
		func (resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response{
			ruleConf,err:=readRuleFile();if err!=nil{
				fmt.Println(err)
				return resp
			}
			if ruleConf.RespRules!=nil{
				for _,r:=range ruleConf.RespRules{
					if !r.Active{
						continue
					}
					regex:=regexp.MustCompile(r.UrlMatchRegexp)
					if regex.MatchString(resp.Request.URL.String()){
						if r.RespAction.SetHeaders!=nil { // 设置请求头
							for _,sh:=range r.RespAction.SetHeaders{
								resp.Header.Set(sh.Header,sh.Value)
							}
						}
						if setBd:=r.RespAction.SetBody; setBd!= (setBody{}) {
							resp.StatusCode = 200
							resp.Header.Del("Location")
							switch setBd.BodyType {
							case 0:
								resp.Header.Set("Content-Type", ContentTypeJson)
							case 1:
								resp.Header.Set("Content-Type", ContentTypeText)
							case 2:
								resp.Header.Set("Content-Type", ContentTypeHtml)
							case 3:
								resp.Header.Set("Content-Type", ContentTypeJpeg)
							default:
								resp.Header.Set("Content-Type", ContentTypeText)
							}
							respFile,err:=os.Open(setBd.BodyFile)
							if err!=nil{
								fmt.Printf("文件%s无法打开",setBd.BodyFile)
								return resp
							}
							defer respFile.Close()
							rBytes,err:=ioutil.ReadAll(respFile)
							if err!=nil{
								fmt.Printf("文件%s打开内容报错请检查",setBd.BodyFile)
								return resp
							}
							buf:=bytes.NewBuffer(rBytes)
							resp.ContentLength = int64(buf.Len())
							resp.Body=ioutil.NopCloser(buf)
							return resp // 设置响应内容则此规则为最后的规则
						}
					}
				}
			}
			return resp
		})

}