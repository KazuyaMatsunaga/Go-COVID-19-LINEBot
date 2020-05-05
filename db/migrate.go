package main

import (
	"covid-19-bot/entity"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func main() {
	db, err := gorm.Open("mysql", "user1:Password_01@tcp(db:3306)/covid19?charset=utf8&parseTime=True")
	if err != nil {
		log.Println(err)
	}
	db.CreateTable(&entity.Covid19Info{})
	defer db.Close()
}
