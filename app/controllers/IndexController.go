package controllers

import (
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
	"net/http"
	"shorturl/app/helps"
	"shorturl/app/services"
	"shorturl/models"
	"strings"
)

func Create(c *gin.Context) {
	url := c.PostForm("url")
	logs.Info("incoming create url request, url: " + url)
	if url == "" || !strings.HasPrefix(url, "http") {
		c.JSON(http.StatusOK, gin.H{
			"code": helps.ParamsError,
			"msg":  "参数错误",
			"data": "",
		})
		return
	}
	code, err := services.UrlService{}.GenCode(url)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": helps.Failed,
			"msg":  "请求出错",
			"data": "",
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": helps.Success,
			"msg":  "ok",
			"data": gin.H{
				"url": models.AppConfig.BaseUrl + code,
			},
		})
		return
	}
}

func Query(c *gin.Context) {
	code := c.PostForm("code")
	logs.Info("incoming query, code: " + code)
	if len(code) != 6 {
		c.JSON(http.StatusOK, gin.H{
			"code": helps.ParamsError,
			"msg":  "参数错误",
			"data": "",
		})
		return
	}
	url, err := services.UrlService{}.RecCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": helps.Success,
			"msg":  err.Error(),
			"data": "",
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": helps.Success,
			"msg":  "ok",
			"data": gin.H{
				"url": url,
			},
		})
		return
	}
}

func Path(c *gin.Context) {
	code := c.Param("code")
	logs.Info("incoming query, code: " + code)
	if len(code) != 6 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	url, err := services.UrlService{}.RecCode(code)

	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Header("Location", url)
	c.AbortWithStatus(302)
}
