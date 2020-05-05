package main

import (
	"covid-19-bot/entity"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Info struct {
	Id     int    `json:"id"`
	Name   string `json:"name_ja"`
	Cases  int    `json:"cases"`
	Deaths int    `json:"deaths"`
}

func main() {
	db, err := gorm.Open("mysql", "user1:Password_01@tcp(db:3306)/covid19?charset=utf8&parseTime=True")
	if err != nil {
		log.Println(err)
	}

	var info_today []Info

	// APIを叩いて情報を取得
	req, err := http.NewRequest("GET", "https://covid19-japan-web-api.now.sh/api/v1/prefectures", nil)
	if err != nil {
		log.Fatal(err)
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(body, &info_today); err != nil {
		log.Fatal(err)
	}

	var cases_today int
	var deaths_today int

	var covid19_info entity.Covid19Info

	cases_today = info_today[12].Cases
	deaths_today = info_today[12].Deaths

	covid19_info = entity.Covid19Info{
		Cases:  cases_today,
		Deaths: deaths_today,
	}

	db.Create(&covid19_info)

	defer db.Close()
}
