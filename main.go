package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	entity "covid-19-bot/entity"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/line/line-bot-sdk-go/linebot"
)

type Info struct {
	Id     int    `json:"id"`
	Name   string `json:"name_ja"`
	Cases  int    `json:"cases"`
	Deaths int    `json:"deaths"`
}

func main() {
	sc := make(chan os.Signal, 1)   // シグナル用のチャンネル作って
	signal.Notify(sc, os.Interrupt) // シグナルを登録する

	// 最初のデータを登録する
	info, _ := getInfo()
	registerDB(info)

loop:
	for {
		select {
		case <-sc:
			{
				log.Println("interrupt")
				break loop
			}
		case <-time.After(60 * time.Second):
			{
				pushMessage()
			}
		}
	}
}

func pushMessage() {
	info, time_now := getInfo()

	registerDB(info)

	cases_before, deaths_before := getTheDayBefore(info)

	log.Printf("%s: Cases=%d , Deaths=%d\n", info[12].Name, info[12].Cases, info[12].Deaths) //　東京の場合は12,都道府県によって数字は変更

	message := time_now + "\n" + info[12].Name + "都内のCOVID-19感染情報\n\n" + "総感染者数：" + strconv.Itoa(info[12].Cases) + "\n総死者数：" + strconv.Itoa(info[12].Deaths) + "\n\n前日比\n" + "総感染者数：" + cases_before + "\n" + "総死者数：" + deaths_before

	line_message := linebot.NewTextMessage(message)

	var messages []linebot.SendingMessage

	messages = append(messages, line_message)

	bot, err := linebot.New(os.Getenv("LINE_CHANNEL_SECRET"), os.Getenv("LINE_CHANNEL_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := bot.PushMessage(os.Getenv("LINE_USER_ID"), messages...).Do(); err != nil {
		log.Fatal(err)
	}
}

// 今日の情報を取得
func getInfo() (info []Info, time_now string) {
	var info_today []Info

	req, err := http.NewRequest("GET", "https://covid19-japan-web-api.now.sh/api/v1/prefectures", nil)
	if err != nil {
		log.Fatal(err)
	}

	jst, _ := time.LoadLocation("Asia/Tokyo")
	t := time.Now().In(jst) // 日本での時刻を取得
	const layout = "2006年01月02日 15時04分"
	s := t.Format(layout)

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

	return info_today, s
}

// DBに今日の情報を登録
func registerDB(info []Info) {
	var cases_today int
	var deaths_today int

	var covid19_info entity.Covid19Info

	db, err := gorm.Open("mysql", "user1:Password_01@tcp(db:3306)/covid19?charset=utf8&parseTime=True")
	if err != nil {
		log.Println(err)
	}

	cases_today = info[12].Cases
	deaths_today = info[12].Deaths

	covid19_info = entity.Covid19Info{
		Cases:  cases_today,
		Deaths: deaths_today,
	}

	db.Create(&covid19_info)

	defer db.Close()
}

// 前日比のデータを計算、取得
func getTheDayBefore(info []Info) (cases string, deaths string) {
	var cases_before int
	var deaths_before int

	var cases_before_str string
	var deaths_before_str string

	var covid19_info_before entity.Covid19Info
	var covid19_info entity.Covid19Info

	db, err := gorm.Open("mysql", "user1:Password_01@tcp(db:3306)/covid19?charset=utf8&parseTime=True")
	if err != nil {
		log.Println(err)
	}

	//　今日のデータを取得
	if err := db.Where("cases = ?", info[12].Cases).First(&covid19_info).Error; err != nil {
		log.Println(err)
	}

	before_id := covid19_info.ID - 1

	// 前日のデータを取得
	if err := db.Where("id = ?", before_id).First(&covid19_info_before).Error; err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Printf("今日のデータ：Id=%d , Cases=%d , Deaths=%d\n", covid19_info.ID, covid19_info.Cases, covid19_info.Deaths)
	log.Printf("前日のデータ：Id=%d , Cases=%d , Deaths=%d\n", covid19_info_before.ID, covid19_info_before.Cases, covid19_info_before.Deaths)

	// 前日比を計算
	cases_before = covid19_info.Cases - covid19_info_before.Cases
	deaths_before = covid19_info.Deaths - covid19_info_before.Deaths

	if cases_before > 0 {
		cases_before_str = "+" + strconv.Itoa(cases_before)
	} else if cases_before < 0 {
		cases_before_str = "-" + strconv.Itoa(cases_before)
	} else if cases_before == 0 {
		cases_before_str = "±" + strconv.Itoa(cases_before)
	}

	if deaths_before > 0 {
		deaths_before_str = "+" + strconv.Itoa(deaths_before)
	} else if deaths_before < 0 {
		deaths_before_str = "-" + strconv.Itoa(deaths_before)
	} else if deaths_before == 0 {
		deaths_before_str = "±" + strconv.Itoa(deaths_before)
	}

	return cases_before_str, deaths_before_str
}
