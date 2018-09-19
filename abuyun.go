package abuyun

import (
	"github.com/liguoqinjim/ruokuai"
	"github.com/parnurzeal/gorequest"
	"io"
	"log"
	"net/http"
	"os"
)

type AbuyunApp struct {
	Username string
	Password string

	PHPSESSID string //保存了整个cookie信息

	RuokuaiApp *ruokuai.RuoKuaiApp
}

func New(username, password string) *AbuyunApp {
	return &AbuyunApp{
		Username: username,
		Password: password,
	}
}

func (app *AbuyunApp) SetRuokuaiApp(ruokuaiApp *ruokuai.RuoKuaiApp) *AbuyunApp {
	app.RuokuaiApp = ruokuaiApp
	return app
}

func (app *AbuyunApp) Login() {
	request := gorequest.New()

	resp, _, errs := request.Get("https://center.abuyun.com/login").
		Set("Referer", "https://center.abuyun.com/").
		End()
	if errs != nil {
		log.Fatalf("errs:%v", errs)
	}

	//log.Println(body)
	//log.Println(resp)

	tmpHeader := http.Header{}
	tmpReq := http.Request{}
	for k, v := range resp.Header {
		if k == "Set-Cookie" {
			for _, v2 := range v {
				tmpHeader.Add("Cookie", v2)
			}
			tmpReq.Header = tmpHeader
			break
		}
	}

	//log.Println("cookies:", tmpReq.Cookies())

	//得到验证码
	resp, _, errs = request.Get("https://center.abuyun.com/captcha").
		AddCookies(tmpReq.Cookies()).
		End()
	if errs != nil {
		log.Fatalf("errs:%v", errs)
	}

	fi, err := os.Create("captcha.png")
	if err != nil {
		log.Fatalf("os.Create error:%v", err)
	}
	defer fi.Close()

	_, err = io.Copy(fi, resp.Body)
	if err != nil {
		log.Fatalf("io.Copy error:%v", err)
	}
	defer resp.Body.Close()

	//解析验证码
	//code := "1234"
	if app.RuokuaiApp == nil {
		log.Fatalf("app.RuokuaiApp nil")
	}
	cr, er := app.RuokuaiApp.Create("3040", "captcha.png")
	if er != nil {
		if er.ErrorCode == "" {
			log.Fatalf("ruokuaierror:%s", er.Error)
		} else {
			log.Fatalf("ruokuaierror:%s,errorcode:%s", er.Error, er.ErrorCode)
		}
	} else {
		if cr.Result == "" {
			log.Fatalf("ir error:%s", cr.Result)
		} else {
			log.Println("验证码为:", cr.Result)
		}
	}

	//登录
	resp, body, errs := request.Post("https://center.abuyun.com/backend/passport/account/auth/verify").
		AddCookies(tmpReq.Cookies()).
		SendStruct(&UserInfo{Name: app.Username, Pass: app.Password, Code: cr.Result, Remember: false}).
		End()
	log.Println("body=", body)
	log.Println("resp=", resp)
}

type UserInfo struct {
	Name     string `json:"name"`
	Pass     string `json:"pass"`
	Code     string `json:"code"`
	Remember bool   `json:"remember"`
}
