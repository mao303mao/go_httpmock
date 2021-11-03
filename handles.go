package main

import (
	"bytes"
	"fmt"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const (
	ContentTypeJson = "application/json;charset=UTF-8"
	ContentTypeText = "text/plain;charset=UTF-8"
	ContentTypeHtml = "text/html;charset=UTF-8"
	ContentTypeJpeg = "image/jpeg"
)

func doResponseRules(proxy *goproxy.ProxyHttpServer){ // response add cors headers
	proxy.OnRequest().DoFunc( // 构造响应
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			if ruleConf.isEmpty() {
				log.Printf("规则文件内为空\n")
				return req,nil
			}
			if ruleConf.NewRespRules==nil || len(ruleConf.NewRespRules)==0 {
				return req,nil
			}
			for _,r:=range ruleConf.NewRespRules {
				if !r.Active {
					continue
				}
				regex:=regexp.MustCompile(r.UrlMatchRegexp)
				if regex.MatchString(ctx.Req.URL.String()){
					rewriteUrl:=strings.TrimSpace(r.ReWriteUrl)
					if r.RespAction==nil && rewriteUrl=="" {
						return req,nil
					}
					if rewriteUrl!=""{
						subMatchs:=regex.FindStringSubmatch(ctx.Req.URL.String())
						for i,sm:=range subMatchs{
							rewriteUrl=strings.ReplaceAll(rewriteUrl,fmt.Sprintf("${%d}",i),sm)
						}
						newURL,err:=url.Parse(rewriteUrl)
						if err!=nil{
							log.Printf("重写的url(%s)格式有误\n",rewriteUrl)
							return req,nil
						}
						ctx.Req.URL=newURL
						return req,nil
					}
					if r.RespAction!=nil && *r.RespAction.SetBody== (setBody{}){
						newResp := &http.Response{}
						newResp.Request = ctx.Req
						newResp.TransferEncoding = ctx.Req.TransferEncoding
						newResp.Header = make(http.Header)
						if !updateResponse(newResp,r){
							continue
						}
						return req,newResp
					}
				}
			}
			return req,nil
		})

	proxy.OnResponse().DoFunc( // 更新响应
		func (resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response{
			if ruleConf.isEmpty() {
				log.Printf("规则文件内为空\n")
				return resp
			}
			if resp==nil{
				log.Println("服务器响应为空，构造新的响应")
				if ruleConf.UpdateRespRules==nil || len(ruleConf.UpdateRespRules)==0{
					return nil
				}
				for _,r:=range ruleConf.UpdateRespRules {
					if !r.Active {
						continue
					}
					regex:=regexp.MustCompile(r.UrlMatchRegexp)
					if regex.MatchString(ctx.Req.URL.String()){
						if r.RespAction==nil || *r.RespAction.SetBody== (setBody{}) {
							return nil
						}
						newResp := &http.Response{}
						newResp.Request = ctx.Req
						newResp.TransferEncoding = ctx.Req.TransferEncoding
						newResp.Header = make(http.Header)
						if !updateResponse(newResp,r){
							continue
						}
						return newResp
					}
				}
				return nil
			}
			if ruleConf.UpdateRespRules!=nil && len(ruleConf.UpdateRespRules)>0{
				for _,r:=range ruleConf.UpdateRespRules{
					if !r.Active{
						continue
					}
					regex:=regexp.MustCompile(r.UrlMatchRegexp)
					if regex.MatchString(ctx.Req.URL.String()){
						if !updateResponse(resp,r){
							continue
						}
						return resp // 设置响应内容则此规则为最后的规则
						}
					}
			}
			return resp
		})

}

func updateResponse(resp *http.Response, r *respRule) bool{
	resp.StatusCode = 200
	resp.Header.Del("Location")
	if r.RespAction.SetHeaders!=nil { // 设置请求头
		for _,sh:=range r.RespAction.SetHeaders{
			resp.Header.Set(sh.Header,sh.Value)
		}
	}
	if *(r.RespAction.SetBody)!= (setBody{}) {
		switch r.RespAction.SetBody.BodyType {
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
		respFile,err:=os.Open(r.RespAction.SetBody.BodyFile)
		if err!=nil{
			log.Printf("文件%s无法打开\n",r.RespAction.SetBody.BodyFile)
			return false
		}
		defer respFile.Close()
		rBytes,err:=ioutil.ReadAll(respFile)
		if err!=nil{
			log.Printf("文件%s打开内容报错请检查\n",r.RespAction.SetBody.BodyFile)
			return false
		}
		buf:=bytes.NewBuffer(rBytes)
		resp.ContentLength = int64(buf.Len())
		resp.Body=ioutil.NopCloser(buf)
		return true
	}
	return true
}