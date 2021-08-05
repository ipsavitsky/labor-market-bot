package main

import (
	"github.com/bigkevmcd/go-configparser"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"time"
)

func errCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	p, err := configparser.NewConfigParserFromFile("default.cfg")
	errCheck(err)

	dbDriver, err := p.Get("DB", "DRIVER")
	errCheck(err)

	dbCreds, err := p.Get("DB", "CREDENTIALS")
	errCheck(err)

	_, err = sqlx.Open(dbDriver, dbCreds)
	errCheck(err)
	log.Println("database connected!")

	apiToken, err := p.Get("TELEGRAM", "API_TOKEN")
	errCheck(err)

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
			_, err := b.Send(m.Sender, "нихуя се че захотел?!\n"+m.Text)
			errCheck(err)
		}
	})

	b.Start()
}
