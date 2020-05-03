FROM golang:1.11.5

ADD . /go/src/covid-19-bot/

WORKDIR /go/src/covid-19-bot/

RUN go get -u github.com/line/line-bot-sdk-go/linebot && go get github.com/jinzhu/gorm && go get github.com/go-sql-driver/mysql

CMD go run main.go