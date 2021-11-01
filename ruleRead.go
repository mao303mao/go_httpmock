package main

import (
	"encoding/json"
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
	RespRules []respRule `json:"respRules"`
}

func readRuleFile() (*rule ,error){
	jsonfile,err:=os.Open("./rules.json")
	if err!=nil{
		return nil, err
	}
	defer jsonfile.Close()
	ruleConf:=&rule{}
	if err=json.NewDecoder(jsonfile).Decode(ruleConf);err!=nil{
		return nil, err
	}
	return ruleConf, nil
}
