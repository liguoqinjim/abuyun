package main

import (
	"encoding/json"
	"github.com/liguoqinjim/abuyun"
	"github.com/liguoqinjim/ruokuai"
	"io/ioutil"
	"log"
)

type User struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	RuokuaiUsername string `json:"ruokuai_username"`
	RuokuaiPassword string `json:"ruokuai_password"`
}

func main() {
	data, err := ioutil.ReadFile("user.json")
	if err != nil {
		log.Fatalf("")
	}

	u := &User{}
	err = json.Unmarshal(data, u)
	if err != nil {
		log.Fatalf("json.Unmarshal error:%v", err)
	}

	ruokuaiApp := ruokuai.Default(u.RuokuaiUsername, u.RuokuaiPassword)

	app := abuyun.New(u.Username, u.Password).SetRuokuaiApp(ruokuaiApp)
	defer app.Close()

	app.Login()

	app.GetHTTPTunnelList(abuyun.HTTP_PROXY_TYPE_DYNAMIC, 1)
}
