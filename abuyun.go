package abuyun

import (
	"bufio"
	"encoding/json"
	"github.com/liguoqinjim/ruokuai"
	"github.com/parnurzeal/gorequest"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	HTTP_PROXY_TYPE_PRO     = iota + 2 //专业版
	HTTP_PROXY_TYPE_DYNAMIC            //动态版
	HTTP_PROXY_TYPE_CLASSIC            //经典版
)

type AbuyunApp struct {
	Username string
	Password string

	PHPSESSID string //保存了整个cookie信息

	RuokuaiApp *ruokuai.RuoKuaiApp

	req *http.Request
}

func New(username, password string) *AbuyunApp {
	app := &AbuyunApp{
		Username: username,
		Password: password,
		req:      &http.Request{},
	}

	f, err := os.Open("cookies")
	if err == nil {
		header := http.Header{}
		app.req.Header = header

		//读取cookies
		buf := bufio.NewReaderSize(f, 0)
		for {
			line, err := buf.ReadString('\n')

			header.Add("Cookie", line)

			if err == io.EOF {
				break
			}
		}

		f.Close()
	}

	return app
}

func (app *AbuyunApp) Close() {
	if len(app.req.Cookies()) > 0 {
		f, err := os.Create("cookies")
		if err != nil {
			log.Fatalf("os.Create error:%v", err)
		}

		w := bufio.NewWriter(f)
		for n, c := range app.req.Cookies() {
			_, err := w.WriteString(c.String())
			log.Println("保存cookie=", c.String())
			if err != nil {
				log.Fatalf("w.WriteString error")
			}

			if n != len(app.req.Cookies())-1 {
				w.WriteString("\n")
			}
		}

		w.Flush()
		f.Close()
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
	app.req.Header = tmpHeader
	for k, v := range resp.Header {
		if k == "Set-Cookie" {
			for _, v2 := range v {
				tmpHeader.Add("Cookie", v2)
			}
		}
	}

	//log.Println("cookies:", tmpReq.Cookies())

	//得到验证码
	resp, _, errs = request.Get("https://center.abuyun.com/captcha").
		AddCookies(app.req.Cookies()).
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
		AddCookies(app.req.Cookies()).
		SendStruct(&UserInfo{Name: app.Username, Pass: app.Password, Code: cr.Result, Remember: false}).
		End()
	if errs != nil {
		log.Fatalf("request.Post errs:%v", errs)
	}
	log.Println("body=", body)
	//log.Println("resp=", resp)

	for k, v := range resp.Header {
		if k == "Set-Cookie" {
			for _, v2 := range v {
				tmpHeader.Add("Cookie", v2)
			}
		}
	}

	body = strings.Replace(body, ")]}',", "", -1)
	log.Println("body=", body)
	result := &LoginResult{}
	err = json.Unmarshal([]byte(body), result)
	if err != nil {
		log.Fatalf("json.Unmarshal error:%v", err)
	}
	log.Printf("loginResult:%v", result)

	//打印cookie
	log.Println("login cookies", app.req.Cookies())
	for _, c := range app.req.Cookies() {
		log.Println("login cookie", c)
	}
}

func (app *AbuyunApp) GetHTTPTunnelList(tunnelType, pageNum int) {
	request := gorequest.New()

	//https://center.abuyun.com/backend/cloud/http/tunnel/lists?level=2&p=1
	_, body, errs := request.Get("https://center.abuyun.com/backend/cloud/http/tunnel/lists").
		Set("Referer", "https://center.abuyun.com/").
		Param("level", strconv.Itoa(tunnelType)).
		Param("p", strconv.Itoa(pageNum)).
		AddCookies(app.req.Cookies()).
		End()
	if errs != nil {
		log.Fatalf("errs:%v", errs)
	}

	body = strings.Replace(body, ")]}',", "", -1)
	log.Println("body=", body)
	result := &HTTPTunnelResult{}
	err := json.Unmarshal([]byte(body), result)
	if err != nil {
		log.Fatalf("json.Unmarshal error:%v", err)
	}
	log.Printf("httpTunnelResult:%v", result)
}

//账号管理
func (app *AbuyunApp) AccountInfo() {
	request := gorequest.New()

	//https://center.abuyun.com/backend/passport/account/self/details
	_, body, errs := request.Get("https://center.abuyun.com/backend/passport/account/self/details").
		Set("Referer", "https://center.abuyun.com/").
		AddCookies(app.req.Cookies()).
		End()
	if errs != nil {
		log.Fatalf("errs:%v", errs)
	}

	body = strings.Replace(body, ")]}',", "", -1)
	log.Println("body=", body)
	result := &AccountInfoResult{}
	err := json.Unmarshal([]byte(body), result)
	if err != nil {
		log.Fatalf("json.Unmarshal error:%v", err)
	}
	log.Printf("accountInfoResult:%v", result)
}

func (app *AbuyunApp) WalletInfo() {
	request := gorequest.New()

	//https://center.abuyun.com/backend/passport/wallet/profile/details
	_, body, errs := request.Get("https://center.abuyun.com/backend/passport/wallet/profile/details").
		Set("Referer", "https://center.abuyun.com/").
		AddCookies(app.req.Cookies()).
		End()
	if errs != nil {
		log.Fatalf("errs:%v", errs)
	}

	body = strings.Replace(body, ")]}',", "", -1)
	log.Println("body=", body)
	result := &WalletInfoResult{}
	err := json.Unmarshal([]byte(body), result)
	if err != nil {
		log.Fatalf("json.Unmarshal error:%v", err)
	}
	log.Printf("wallInfoResult:%v", result)
}

type UserInfo struct {
	Name     string `json:"name"`
	Pass     string `json:"pass"`
	Code     string `json:"code"`
	Remember bool   `json:"remember"`
}

type LoginResult struct {
	Code   int `json:"code"`
	Result struct {
		Account struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"account"`
		Token string `json:"token"`
	} `json:"result"`
}

type HTTPTunnelResult struct {
	Code   int `json:"code"`
	Result struct {
		Time     string `json:"time"`
		Current  int    `json:"current"`
		Total    int    `json:"total"`
		Capacity int    `json:"capacity"`
		Lists    []struct {
			TunnelID     string `json:"TunnelId"`
			TunnelType   string `json:"TunnelType"`
			TunnelLevel  string `json:"TunnelLevel"`
			GroupID      string `json:"GroupId"`
			CityID       string `json:"CityId"`
			ProvID       string `json:"ProvId"`
			DefRequests  string `json:"DefRequests"`
			IsCustomized string `json:"IsCustomized"`
			AuthMode     string `json:"AuthMode"`
			BindingIP    string `json:"BindingIp"`
			Duration     string `json:"Duration"`
			Requests     string `json:"Requests"`
			License      string `json:"License"`
			SecretKey    string `json:"SecretKey"`
			Status       string `json:"Status"`
			Memo         string `json:"Memo"`
			ChargeTime   string `json:"ChargeTime"`
			ExpireTime   string `json:"ExpireTime"`
			ProvName     string `json:"ProvName"`
			CityName     string `json:"CityName"`
			IsExpired    bool   `json:"IsExpired"`
		} `json:"lists"`
		TipFeature bool   `json:"tipFeature"`
		SessionKey string `json:"sessionKey"`
	} `json:"result"`
}

type AccountInfoResult struct {
	Code   int `json:"code"`
	Result struct {
		Profile struct {
			UserID    string `json:"UserId"`
			LoginName string `json:"LoginName"`
			Mobile    string `json:"Mobile"`
			RegTime   string `json:"RegTime"`
			Subject   string `json:"Subject"`
		} `json:"profile"`
	} `json:"result"`
}

type WalletInfoResult struct {
	Code   int `json:"code"`
	Result struct {
		Wallet struct {
			CashBalance    string      `json:"CashBalance"`
			FreeBalance    string      `json:"FreeBalance"`
			TotalSpending  string      `json:"TotalSpending"`
			AlipayAccount  interface{} `json:"AlipayAccount"`
			AlipayRealName interface{} `json:"AlipayRealName"`
		} `json:"wallet"`
	} `json:"result"`
}
