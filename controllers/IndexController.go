package controllers

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"shorturl/models"
	"shorturl/services"
	"strconv"
	"strings"
	"time"

	"github.com/mileusna/useragent"

	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
)

var Index = &IndexController{}

type IndexController struct {
	Controller
}

type CreateRequest struct {
	Urls []string
}

type result struct {
	url  string
	code string
}

//单个生成短网址
func (i *IndexController) Create(c *gin.Context) {
	lUrl := c.PostForm("url")
	expireDay, _ := strconv.Atoi(c.PostForm("expireday"))
	logs.Info("incoming create url request, url: " + lUrl)
	if lUrl == "" {
		logs.Info("url is empty, url: " + lUrl)
		i.failed(c, models.ParamsError, "参数错误")
		return
	}

	if ok := govalidator.IsURL(lUrl); !ok {
		logs.Info("url is invalid, url: " + lUrl)
		i.failed(c, models.ParamsError, "无效的url")
		return
	}
	//ip
	ip := ClientIP(c.Request)

	shortUrl, err := services.UrlService{}.GenShortUrl(lUrl, ip, expireDay)
	if err != nil {
		logs.Error("gen shortUrl failed, error: " + err.Error())
		i.failed(c, models.Failed, "请求出错")
		return
	} else {
		createlogs.Info("[create]: " + lUrl + " => " + shortUrl)
		i.success(c, gin.H{
			"url": shortUrl,
		})
		return
	}
}

//批量生成短网址
func (i *IndexController) MultiCreate(c *gin.Context) {
	var request CreateRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		i.failed(c, models.ParamsError, "参数错误")
		return
	}
	if len(request.Urls) == 0 {
		i.failed(c, models.ParamsError, "url不能为空")
		return
	}
	if len(request.Urls) > 50 {
		i.failed(c, models.ParamsError, "最多可同时生成50个")
		return
	}

	str, _ := json.Marshal(request.Urls)
	logs.Info("incoming multicreate url request, url: " + string(str))

	//ip
	ip := ClientIP(c.Request)
	expireDay, _ := strconv.Atoi(c.PostForm("expireday"))
	var cCode = make(chan result)
	for _, v := range request.Urls {
		go func(lUrl string) {
			if ok := govalidator.IsURL(lUrl); !ok {
				logs.Info("url is invalid, url: " + lUrl)
				cCode <- result{lUrl, "url is not valid"}
				return
			}
			shortUrl, err := services.UrlService{}.GenShortUrl(lUrl, ip, expireDay)
			if err != nil {
				logs.Error("gen shortUrl failed, error: " + err.Error())
				cCode <- result{lUrl, err.Error()}
			} else {
				createlogs.Info("[create]: " + lUrl + " => " + shortUrl)
				cCode <- result{lUrl, shortUrl}
			}
		}(v)
	}

	var results = make(map[string]interface{})
	for {
		res := <-cCode
		results[res.url] = res.code
		if len(results) == len(request.Urls) {
			close(cCode)
			i.success(c, gin.H{"urls": results})
			return
		}
	}
}

func (i *IndexController) Query(c *gin.Context) {
	sUrl := c.PostForm("url")

	parse, err := url.Parse(sUrl)
	if err != nil {
		i.failed(c, models.ParamsError, err.Error())
		return
	}
	code := strings.Trim(parse.Path, "/")
	if len(code) < 3 || len(code) > 6 {
		i.failed(c, models.ParamsError, "参数错误")
		return
	}
	lUrl, err := services.UrlService{}.RestoreUrl(code)
	if err != nil {
		i.failed(c, models.NotFound, err.Error())
		return
	} else {
		i.success(c, gin.H{
			"url": lUrl,
		})
		return
	}
}

func (i *IndexController) Path(c *gin.Context) {
	code := c.Param("code")
	logs.Info("incoming query, code: " + code)
	if len(code) < 3 || len(code) > 6 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	lUrl, err := services.UrlService{}.RestoreUrl(code)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	ip := ClientIP(c.Request)
	uaStr := c.Request.UserAgent()
	userAgent := ua.Parse(uaStr)
	os := userAgent.OS //ua.OSVersion
	browser := userAgent.Name
	tnow := strconv.FormatInt(int64(time.Now().Unix()), 10)
	jumplogs.Info(tnow + "@#" + code + "@#" + lUrl + "@#" + ip + "@#" + os + "@#" + browser + "@#" + uaStr)
	c.Header("Location", lUrl)
	c.AbortWithStatus(302)
	return
}

// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

var createlogs = logs.NewLogger(10000)
var jumplogs = logs.NewLogger(10000)

func init() {
	createlogs.SetLogger(logs.AdapterFile, `{"filename":"storage/logs/create_`+`.log","Hourly":true}`)
	jumplogs.SetLogger(logs.AdapterFile, `{"filename":"storage/logs/jump_`+`.log","Hourly":true}`)
}
