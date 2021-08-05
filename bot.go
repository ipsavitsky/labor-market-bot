package main

import (
	"github.com/bigkevmcd/go-configparser"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"time"
)

type pendingRequest struct {
	message        string
	price          float64
	expirationDate string
}

type requestRepo map[int]*pendingRequest

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

	_, err = sqlx.Open(dbDriver, dbCreds)
	errCheck(err)
	log.Println("database connected!")
}

func main() {

	p, err := configparser.NewConfigParserFromFile("default.cfg")
	errCheck(err)

	processDbQuery(p, "")

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

	b.Handle(tb.OnText, func(m *tb.Message) {
		isReplyTo := m.ReplyTo
		if isReplyTo == nil {
			return
		} else if isReplyTo.Text == "Введите описание запроса пж" {
			log.Println(m.Text)
			requests[m.Sender.ID] = &pendingRequest{message: m.Text}
			_, err := b.Send(m.Sender, "Задайте предположительную дату выполнения реквеста", &tb.ReplyMarkup{
				ForceReply: true,
			})
			errCheck(err)
		} else if isReplyTo.Text == "Задайте предположительную дату выполнения реквеста" {
			log.Println(m.Text)
			requests[m.Sender.ID].expirationDate = m.Text
			_, err := b.Send(m.Sender, "Задайте предположительную цену реквеста", &tb.ReplyMarkup{
				ForceReply: true,
			})
			errCheck(err)
		} else if isReplyTo.Text == "Задайте предположительную цену реквеста" {
			log.Println(m.Text)
			requests[m.Sender.ID].price, err = strconv.ParseFloat(m.Text, 64)
			errCheck(err)
			log.Println("пора бы в логи записать...", requests)
		}
	})

	b.Start()
}
