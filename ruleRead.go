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
type setBody struct {
	BodyType int `json:"bodyType"`
	BodyFile string `json:"bodyFile"`
}
type respAction struct {
	SetHeaders []setHeader `json:"setHeaders"`
	SetBody setBody `json:"setBody"`
}
type respRule struct {
	Active bool `json:"active"`
	UrlMatchRegexp string `json:"urlMatchRegexp"`
	RespAction respAction `json:"respAction"`
}
type rule struct {
	Author string `json:"author"`
	CreateTime string `json:"createTime"`
	UpdateDate string `json:"updateDate"`
	UpdateRespRules []respRule `json:"updateRespRules"`
	NewRespRules []respRule `json:"newRespRules"`
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
		log.Println("...配置文件已经更新...",fileUP)
	}
	return nil
}
