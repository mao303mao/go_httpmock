package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	ContentTypeJson = "application/json;charset=UTF-8"
	ContentTypeText = "text/plain;charset=UTF-8"
	ContentTypeHtml = "text/html;charset=UTF-8"
	ContentTypeJpeg = "image/jpeg"
	ContentTypePng = "image/png"
	ContentTypeGif = "image/gif"
	ContentTypeJs   = "application/javascript"
	ContentTypeCss  = "text/css"
	ContentXml = "text/xml"
)

func getContentTypeBySuffix(suffix string) string{
	switch strings.ToLower(suffix) {
	case ".jpg",".jpeg":
		return ContentTypeJpeg
	case ".png":
		return ContentTypePng
	case ".gif":
		return ContentTypeGif
	case ".htm",".html",".jsp":
		return ContentTypeHtml
	case ".css":
		return ContentTypeCss
	case ".js":
		return ContentTypeJs
	case ".json":
		return ContentTypeJson
	case ".xml":
		return ContentXml
	default:
		return ContentTypeText
	}
}



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
						ctx.Req.Host=newURL.Host
						return req,nil
					}
					if r.RespAction!=nil && strings.TrimSpace(r.RespAction.BodyFile)!=""{
						newResp := &http.Response{}
						newResp.Request = ctx.Req
						newResp.TransferEncoding = ctx.Req.TransferEncoding
						newResp.Header = make(http.Header)
						updateResponse(newResp,r,false)
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
			//--------新代码-------
			if ruleConf.UpdateRespRules==nil || len(ruleConf.UpdateRespRules)==0{
				return nil
			}
			for _,r:=range ruleConf.UpdateRespRules{
				if !r.Active{  // 规则不启用，下一个
					continue
				}
				regex:=regexp.MustCompile(r.UrlMatchRegexp)
				if regex.MatchString(ctx.Req.URL.String()) { // 请求的url满足匹配规则,且只会处理第1个规则
					if resp == nil {
						if r.RespAction == nil || strings.TrimSpace(r.RespAction.BodyFile) == "" {
							return nil
						}
						newResp := &http.Response{}
						newResp.Request = ctx.Req
						newResp.TransferEncoding = ctx.Req.TransferEncoding
						newResp.Header = make(http.Header)
						updateResponse(newResp, r,false)
						return newResp
					}
					updateResponse(resp, r,true)
					return resp // 设置响应内容则此规则为最后的规则
				}
			}
			return resp
		})
}

func updateResponse(resp *http.Response, r *respRule,recordFlag bool) {// 更新响应内容
	resp.StatusCode = 200
	resp.Header.Del("Location")
	bodyFile:=strings.TrimSpace(r.RespAction.BodyFile)
	if bodyFile!="" {
		respFile,err:=os.Open(bodyFile)
		if err!=nil{
			if os.IsNotExist(err) && recordFlag && resp.Header.Get("Content-Type")!="" &&
				strings.Contains(resp.Header.Get("Content-Type"),"application/json") && resp.Body!=nil{
				go func() {
					rbody, err := ioutil.ReadAll(resp.Body) // 读取后resp.Body的内容就为空
					if err!=nil{
						log.Printf("读取相应的body失败,异常为%s\n",err)
						return
					}
					defer resp.Body.Close()
					respFile,err=os.OpenFile(bodyFile,os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0766)
					if err!=nil{
						log.Printf("创建文件%s失败,异常为%s\n",bodyFile,err)
						return
					}
					defer respFile.Close()
					var jsonBeuti bytes.Buffer
					err = json.Indent(&jsonBeuti,rbody,"","  ")
					if err!=nil{
						log.Printf("服务端相应的body内容无法json格式化，请检查")
						return
					}
					respFile.Write(jsonBeuti.Bytes())
				}()
				return
			}
			log.Printf("文件%s无法打开,异常为%s\n,如果服务端返回了对应json,会自动生成对应文件",bodyFile,err)
			return
		}
		defer respFile.Close()
		rBytes,err:=ioutil.ReadAll(respFile)
		if err!=nil{
			log.Printf("文件%s打开内容报错请检查\n",bodyFile)
			return
		}
		buf:=bytes.NewBuffer(rBytes)
		resp.ContentLength = int64(buf.Len())
		resp.Body=ioutil.NopCloser(buf)
		suffix := path.Ext(bodyFile)
		resp.Header.Set("Content-Type", getContentTypeBySuffix(suffix))
		return
	}
	if r.RespAction.SetHeaders!=nil { // 设置请求头
		for _,sh:=range r.RespAction.SetHeaders{
			resp.Header.Set(sh.Header,sh.Value)
		}
	}
}