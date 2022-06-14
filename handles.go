package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elazarl/goproxy"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
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
	ContentTypeSvg = "image/svg+xml"
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
	case ".svg":
		return ContentTypeSvg
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
			if ctx.Req.URL.Scheme=="https" { // 处理https的url多了443端口的BUG
				ctx.Req.URL.Host=strings.Replace(ctx.Req.URL.Host,":443","",1)
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
						refererUrl,err:=url.Parse(req.Header.Get("referer"))
						if err!=nil{
							refererUrl=nil
						}
						newResp := &http.Response{}
						newResp.Request = ctx.Req
						newResp.TransferEncoding = ctx.Req.TransferEncoding
						newResp.Header = make(http.Header)
						updateFlag:=updateResponse(newResp,r,false,refererUrl)
						if updateFlag==1{
							return req,newResp
						}
						return req,nil
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
			if ruleConf.UpdateRespRules==nil || len(ruleConf.UpdateRespRules)==0{
				return nil
			}
			for _,r:=range ruleConf.UpdateRespRules{
				if !r.Active{  // 规则不启用，下一个
					continue
				}
				regex:=regexp.MustCompile(r.UrlMatchRegexp)
				if regex.MatchString(ctx.Req.URL.String()) { // 请求的url满足匹配规则,如果处理了body则直接结束
					refererUrl,err:=url.Parse(ctx.Req.Header.Get("referer"))
					if err!=nil{
						refererUrl=nil
					}
					if resp == nil {
						if r.RespAction == nil || strings.TrimSpace(r.RespAction.BodyFile) == "" {
							return nil
						}

						newResp := &http.Response{}
						newResp.Request = ctx.Req
						newResp.TransferEncoding = ctx.Req.TransferEncoding
						newResp.Header = make(http.Header)
						if updateResponse(newResp, r,false,refererUrl)==0{
							continue
						}
						return newResp
					}
					if (resp.StatusCode<300 ||  resp.StatusCode>=400) && updateResponse(resp, r,true,refererUrl)==0{
						continue
					}
					return resp // 设置响应内容则此规则为最后的规则
				}
			}
			return resp
		})
}

func updateResponse(resp *http.Response, r *respRule,recordFlag bool,refererUrl *url.URL) int{// 更新响应内容,return 0-仅处理header，1-处理了body
	resp.StatusCode = 200
	resp.Header.Del("Location")
	bodyFile:=strings.TrimSpace(r.RespAction.BodyFile)
	setResponseFlag:=0
	if bodyFile!="" {
		bodyFile="./respFiles/"+bodyFile // 如果改造WebUI，此处需要防止“任意文件读取”
		respFile,err:=os.Open(bodyFile)
		if err!=nil{
			log.Printf("文件%s无法打开,异常为%s,如果服务端返回了对应json,会自动生成对应文件",bodyFile,err)
			if os.IsNotExist(err) && recordFlag && resp.Header.Get("Content-Type")!="" &&
				strings.Contains(resp.Header.Get("Content-Type"),"application/json") && resp.Body!=nil{
				rbody, err := ioutil.ReadAll(resp.Body) // 读取后resp.Body的内容就为空
				if err!=nil{
					log.Printf("读取相应的body失败,异常为%s\n",err)
					return 0 // 异常直接结束
				}
				resp.Body=ioutil.NopCloser(bytes.NewBuffer(rbody)) // 需要再次将内容写回去
				defer resp.Body.Close()
				go func() {
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
				setResponseFlag=1
			} else{
				return setResponseFlag
			}
		}else{
			defer respFile.Close()
			rBytes,err:=ioutil.ReadAll(respFile)
			if err!=nil{
				log.Printf("文件%s打开内容报错请检查\n",bodyFile)
				return 0 // 异常直接结束
			}
			buf:=bytes.NewBuffer(rBytes)
			resp.ContentLength = int64(buf.Len())
			resp.Body=ioutil.NopCloser(buf)
			suffix := path.Ext(bodyFile)
			resp.Header.Set("Content-Type", getContentTypeBySuffix(suffix))
			setResponseFlag=1
		}
	}
	if r.RespAction.SetHeaders!=nil { // 设置请求头
		for _,sh:=range r.RespAction.SetHeaders{
			resp.Header.Set(sh.Header,sh.Value)
		}
	}
	if strings.TrimSpace(r.RespAction.PassCORS)!=""{
		allOrigin:=r.RespAction.PassCORS
		if r.RespAction.PassCORS=="*" && refererUrl!=nil{
			allOrigin=refererUrl.Scheme+"://"+refererUrl.Host
		}
		resp.Header.Set("Access-Control-Allow-Origin",allOrigin)
		resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		resp.Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		resp.Header.Set("Access-Control-Allow-Credentials","true")
	}
	return setResponseFlag
}

// 下载证书
func certDownloadHandler(w http.ResponseWriter, r *http.Request)  {
	certFile,err:=os.Open("./z.x509.cer")
	if err!=nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer certFile.Close()
	fileHeader:=make([]byte,512)
	certFile.Read(fileHeader)
	fileStat,_:=certFile.Stat()
	w.Header().Set("Content-Disposition", "attachment; filename=" + "z.x509.cer")
	w.Header().Set("Content-Type", http.DetectContentType(fileHeader))
	w.Header().Set("Content-Length", strconv.FormatInt(fileStat.Size(), 10))
	certFile.Seek(0, 0)
	_,err=io.Copy(w, certFile)
	if err!=nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request,folder string ) {
	path := "./"+folder + r.URL.Path
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	defer f.Close()
	//d, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	ext := filepath.Ext(path)
	if contentType := getContentTypeBySuffix(ext); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	//w.Header().Set("Content-Length", strconv.FormatInt(d.Size(), 10))
	w.Write(data)
}

func saveConf(w http.ResponseWriter, r *http.Request){
	if r.Method!="POST"{
		w.WriteHeader(405)
		return
	}
	jsonStr:=r.PostFormValue("json")
	conFilePath:="./rules.json"
	var conFile *os.File
	if _,err:=os.Stat(conFilePath);err!=nil && os.IsNotExist(err){
		conFile,_=os.OpenFile(conFilePath,os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0766)
	}else{
		conFile,_=os.OpenFile(conFilePath,os.O_RDWR|os.O_TRUNC, 0766)
	}
	defer conFile.Close()
	conFile.Write([]byte(jsonStr))
	w.Header().Set("Content-Type", ContentTypeJson)
	w.Write([]byte(`{"code":0}`))
}