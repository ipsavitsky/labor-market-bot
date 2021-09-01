package main

import (
	"fmt"
	"github.com/bigkevmcd/go-configparser"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"time"
)

type Request struct {
	Id             int64   `db:"id"`
	Message        string  `db:"request_desc"`
	Price          float64 `db:"price"`
	ExpirationDate string  `db:"completion_time"`
}

type requestRepo map[int]*Request

func errCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func processDbQuery(p *configparser.ConfigParser, query string) {
	dbDriver, err := p.Get("DB", "DRIVER")
	errCheck(err)

	dbCreds, err := p.Get("DB", "CREDENTIALS")
	errCheck(err)

	db, err := sqlx.Open(dbDriver, dbCreds)
	errCheck(err)

	log.Println("database connected!")

	db.MustExec(query)
}

func processDbSelectQuery(p *configparser.ConfigParser, selectQuery string) []Request {
	dbDriver, err := p.Get("DB", "DRIVER")
	errCheck(err)

	dbCreds, err := p.Get("DB", "CREDENTIALS")
	errCheck(err)

	db, err := sqlx.Open(dbDriver, dbCreds)
	errCheck(err)

	log.Println("database connected!")

	res := []Request{}

	err = db.Select(&res, selectQuery)
	errCheck(err)

	return res
}

func getLastId(p *configparser.ConfigParser) int {
	dbDriver, err := p.Get("DB", "DRIVER")
	errCheck(err)

	dbCreds, err := p.Get("DB", "CREDENTIALS")
	errCheck(err)

	db, err := sqlx.Open(dbDriver, dbCreds)
	errCheck(err)

	log.Println("database connected!")

	res := struct {
		id int `db:"last_insert_id()"`
	}{}
	db.Select(&res, "select last_insert_id()")
	return res.id
}

func main() {
	p, err := configparser.NewConfigParserFromFile("default.cfg")
	errCheck(err)

	apiToken, err := p.Get("TELEGRAM", "API_TOKEN")
	errCheck(err)

	requests := make(requestRepo)

	b, err := tb.NewBot(tb.Settings{
		Token:  apiToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	errCheck(err)

	b.Handle("/help", func(m *tb.Message) {
		_, err := b.Send(m.Sender, "хелп в работе")
		errCheck(err)
	})

	b.Handle("/add_request", func(m *tb.Message) {
		_, err := b.Send(m.Sender, "Введите описание запроса пж", &tb.ReplyMarkup{
			ForceReply: true,
		})
		errCheck(err)
	})

	b.Handle("/list_requests", func(m *tb.Message) {
		reqArr := processDbSelectQuery(p, "select id, request_desc, price, completion_time from requests")
		//log.Println("lines retrieved: " + string(len(reqArr)))
		for _, curReq := range reqArr {
			_, err := b.Send(m.Sender, fmt.Sprintf("id: %d\nmessage: %s\nprice: %f\nexp_date: %s", curReq.Id, curReq.Message, curReq.Price, curReq.ExpirationDate))
			errCheck(err)
		}
	})

	b.Handle("/my_requests", func(m *tb.Message) {
		reqArr := processDbSelectQuery(p, fmt.Sprintf("select id, request_desc, price, completion_time from requests where user_id = %d", m.Sender.ID))
		//log.Println("lines retrieved: " + string(len(reqArr)))
		for _, curReq := range reqArr {
			_, err := b.Send(m.Sender, fmt.Sprintf("id: %d\nmessage: %s\nprice: %f\nexp_date: %s", curReq.Id, curReq.Message, curReq.Price, curReq.ExpirationDate))
			errCheck(err)
		}
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		isReplyTo := m.ReplyTo
		if isReplyTo == nil {
			return
		} else if isReplyTo.Text == "Введите описание запроса пж" {
			log.Println(m.Text)
			requests[m.Sender.ID] = &Request{Message: m.Text}
			_, err := b.Send(m.Sender, "Задайте предположительную дату выполнения реквеста", &tb.ReplyMarkup{
				ForceReply: true,
			})
			errCheck(err)
		} else if isReplyTo.Text == "Задайте предположительную дату выполнения реквеста" {
			log.Println(m.Text)
			requests[m.Sender.ID].ExpirationDate = m.Text
			_, err = b.Send(m.Sender, "Задайте предположительную цену реквеста", &tb.ReplyMarkup{
				ForceReply: true,
			})
			errCheck(err)
		} else if isReplyTo.Text == "Задайте предположительную цену реквеста" {
			log.Println(m.Text)
			//todo how to fucking do regexp....
			//matched, err := regexp.MatchString("^([0-9]+\.?[0-9]*|\.[0-9]+)$", m.Text)
			//if !matched {
			//	todo do a loopback
			//log.Println("DIE!!!!")
			//}
			requests[m.Sender.ID].Price, err = strconv.ParseFloat(m.Text, 64)
			errCheck(err)
			req := requests[m.Sender.ID]
			strRequest := fmt.Sprintf("insert into requests (user_id, user_name, request_desc, price, completion_time) values (%d, '%s', '%s', %f, '%s')", m.Sender.ID, m.Sender.Username, req.Message, req.Price, req.ExpirationDate)
			processDbQuery(p, strRequest)
			id := getLastId(p)
			_, err = b.Send(m.Sender, fmt.Sprintf("Запрос %d успешно добавлен в базу", id))
			errCheck(err)
		}
	})

	b.Start()
}
