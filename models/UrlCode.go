package models

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type UrlCode struct {
	Id        int `gorm:"primary_key"`
	MD5       string
	Code      string
	Url       string
	Click     int
	CreatedAt int
	ExpireDay int
	Ip        string
}

func (UrlCode) AddUrl(url, ip string, expireDay int) int {
	var uc UrlCode
	uc.Url = url
	uc.Code = ""
	uc.MD5 = MD5(url)
	uc.CreatedAt = int(time.Now().Unix())
	uc.Ip = ip
	uc.ExpireDay = expireDay
	DB.Create(&uc)
	return uc.Id
}

func (UrlCode) GetByUrl(url string) UrlCode {
	var result UrlCode
	DB.Where("md5 = ?", MD5(url)).Find(&result)
	return result
}

func (UrlCode) GetByCode(code string) UrlCode {
	var uc UrlCode
	DB.Where("code = ?", code).First(&uc)
	return uc
}

func (UrlCode) UpdateCode(id int, code string) error {
	var uc UrlCode
	DB.Find(&uc, id)
	uc.Code = code
	DB.Save(&uc)
	if DB.Error != nil {
		return DB.Error
	}
	return nil
}
