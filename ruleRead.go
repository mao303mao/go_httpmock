package main

import (
	"encoding/json"
	"log"
	"os"
)

type setHeader struct {
	Header string `json:"header"`
	Value string `json:"value"`
}

type respAction struct {
	SetHeaders []*setHeader `json:"setHeaders"`
	BodyFile string `json:"bodyFile"`
}
type respRule struct {
	Active bool `json:"active"`
	UrlMatchRegexp string `json:"urlMatchRegexp"`
	RespAction *respAction `json:"respAction"`
	ReWriteUrl string `json:"reWriteUrl"`
}
type rule struct {
	Author string `json:"author"`
	CreateTime string `json:"createTime"`
	UpdateDate string `json:"updateDate"`
	UpdateRespRules []*respRule `json:"updateRespRules"`
	NewRespRules []*respRule `json:"newRespRules"`
}

func (this *rule)isEmpty() bool{
	if  (this.UpdateRespRules==nil || len(this.UpdateRespRules)==0) && (this.NewRespRules==nil || len(this.NewRespRules)==0){
		return true
	}

	return false
}

func readRuleFile() error{
	jsonfile,err:=os.Open("./rules.json")
	if err!=nil{
		return err
	}
	defer jsonfile.Close()
	fs,err:=jsonfile.Stat()
	if err!=nil{
		return err
	}
	if fs.ModTime().Unix()>fileUP{
		if err=json.NewDecoder(jsonfile).Decode(ruleConf);err!=nil{
			return err
		}
		fileUP=fs.ModTime().Unix()
		log.Println("...配置文件已经更新(",fs.ModTime().Format("2006-01-02 15:04:05"),")")
	}
	return nil
}