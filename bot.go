package main

import (
	"database/sql"
	"fmt"
	"github.com/bigkevmcd/go-configparser"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"strconv"
	"time"
)

var (
	writer, _ = os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	log       = logger.Init("Logger", true, true, writer)
)

type Request struct {
	Id                 int64          `db:"id"`
	UserID             int64          `db:"user_id"`
	ExecutorID         sql.NullInt64  `db:"executor_id"`
	UserName           string         `db:"user_name"`
	ExecutorName       sql.NullString `db:"executor_name"`
	RequestDescription string         `db:"request_desc"`
	State              string         `db:"state"`
	Price              float64        `db:"price"`
	CreationTime       string         `db:"creation_time"`
	CompletionDate     string         `db:"completion_time"`
}

type requestRepo map[int]*Request

func processDbQuery(p *configparser.ConfigParser, query string) {
	dbDriver, err := p.Get("DB", "DRIVER")

	if err != nil {
		log.Fatal(err)
	}

	dbCreds, err := p.Get("DB", "CREDENTIALS")

	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Open(dbDriver, dbCreds)

	if err != nil {
		log.Fatal(err)
	}

	db.MustExec(query)
}

func processDbInsertQuery(p *configparser.ConfigParser, query string) int {
	dbDriver, err := p.Get("DB", "DRIVER")

	if err != nil {
		log.Fatal(err)
	}

	dbCreds, err := p.Get("DB", "CREDENTIALS")

	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Open(dbDriver, dbCreds)

	if err != nil {
		log.Fatal(err)
	}

	res := db.MustExec(query)

	resInt, err := res.LastInsertId()

	if err != nil {
		log.Fatal(err)
	}

	return int(resInt)
}

func processDbSelectQuery(p *configparser.ConfigParser, selectQuery string) []Request {
	dbDriver, err := p.Get("DB", "DRIVER")

	if err != nil {
		log.Fatal(err)
	}

	dbCreds, err := p.Get("DB", "CREDENTIALS")

	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Open(dbDriver, dbCreds)

	if err != nil {
		log.Fatal(err)
	}

	var res []Request

	err = db.Select(&res, selectQuery)
	if err != nil {
		log.Fatal(err)
	}

	return res
}

func main() {

	p, err := configparser.NewConfigParserFromFile("default.cfg")

	if err != nil {
		log.Fatal(err)
	}

	apiToken, err := p.Get("TELEGRAM", "API_TOKEN")

	if err != nil {
		log.Fatal(err)
	}

	requests := make(requestRepo)

	b, err := tb.NewBot(tb.Settings{
		Token:  apiToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
	}

	b.Handle("/help", func(m *tb.Message) {
		_, err := b.Send(m.Sender, "хелп в работе")

		if err != nil {
			log.Fatal(err)
		}
	})

	b.Handle("/add_request", func(m *tb.Message) {
		_, err := b.Send(m.Sender, "Введите описание запроса пж", &tb.ReplyMarkup{
			ForceReply: true,
		})
		if err != nil {
			log.Fatal(err)
		}
	})

	b.Handle("/list_requests", func(m *tb.Message) {
		reqArr := processDbSelectQuery(p, "select * from requests where state = 'free'")
		//log.Println("lines retrieved: " + string(len(reqArr)))
		for _, curReq := range reqArr {
			_, err := b.Send(m.Sender, fmt.Sprintf("id: %d\nmessage: %s\nprice: %f\nexp_date: %s", curReq.Id, curReq.RequestDescription, curReq.Price, curReq.CompletionDate))
			if err != nil {
				log.Fatal(err)
			}
		}

		if len(reqArr) == 0 {
			_, err := b.Send(m.Sender, "No free requests currently")
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Handle("/my_requests", func(m *tb.Message) {
		reqArr := processDbSelectQuery(p, fmt.Sprintf("select * from requests where user_id = %d", m.Sender.ID))
		//log.Println("lines retrieved: " + string(len(reqArr)))
		for _, curReq := range reqArr {
			_, err := b.Send(m.Sender, fmt.Sprintf("id: %d\nmessage: %s\nprice: %f\nexp_date: %s", curReq.Id, curReq.RequestDescription, curReq.Price, curReq.CompletionDate))
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Handle("/accept", func(m *tb.Message) {
		requestId, _ := strconv.Atoi(m.Payload)
		//search the fucking db
		reqArr := processDbSelectQuery(p, fmt.Sprintf("select * from requests where id = %d", requestId))
		if len(reqArr) == 1 {
			processDbQuery(p, fmt.Sprintf("update requests set executor_id = %d, executor_name = '%s', state = 'in_work' where id = %d", m.Sender.ID, m.Sender.Username, requestId))
			_, err := b.Send(m.Sender, fmt.Sprintf("Реквест %d принят и закреплен за вами!", requestId))
			if err != nil {
				log.Fatal(err)
			}
			_, err = b.Send(&tb.User{ID: int(reqArr[0].UserID)}, fmt.Sprintf("Реквест %d принят пользователем @%s!", requestId, m.Sender.Username))
			if err != nil {
				log.Fatal(err)
			}

		} else if len(reqArr) == 0 {
			_, err := b.Send(m.Sender, fmt.Sprintf("your request is NOT found!"))
			if err != nil {
				log.Fatal(err)
			}
		} else {
			//todo fucking die
		}
	})

	b.Handle("/accepted_requests", func(m *tb.Message) {
		reqArr := processDbSelectQuery(p, fmt.Sprintf("select * from requests where executor_id = %d", m.Sender.ID))
		for _, curReq := range reqArr {
			_, err := b.Send(m.Sender, fmt.Sprintf("id: %d\nmessage: %s\nprice: %f\nexp_date: %s", curReq.Id, curReq.RequestDescription, curReq.Price, curReq.CompletionDate))
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		//todo make unadressable individually
		isReplyTo := m.ReplyTo
		if isReplyTo == nil {
			return
		} else if isReplyTo.Text == "Введите описание запроса пж" {
			log.Info(m.Text)
			requests[m.Sender.ID] = &Request{RequestDescription: m.Text}
			_, err := b.Send(m.Sender, "Задайте предположительную дату выполнения реквеста", &tb.ReplyMarkup{
				ForceReply: true,
			})
			if err != nil {
				log.Fatal(err)
			}
		} else if isReplyTo.Text == "Задайте предположительную дату выполнения реквеста" {
			log.Info(m.Text)
			requests[m.Sender.ID].CompletionDate = m.Text
			_, err := b.Send(m.Sender, "Задайте предположительную цену реквеста", &tb.ReplyMarkup{
				ForceReply: true,
			})
			if err != nil {
				log.Fatal(err)
			}
		} else if isReplyTo.Text == "Задайте предположительную цену реквеста" {
			log.Info(m.Text)
			//todo how to fucking do regexp....
			//matched, err := regexp.MatchString("^([0-9]+\.?[0-9]*|\.[0-9]+)$", m.Text)
			//if !matched {
			//	todo do a loopback
			//log.Println("DIE!!!!")
			//}
			priceVal, _ := strconv.ParseFloat(m.Text, 64)
			requests[m.Sender.ID].Price = priceVal
			req := requests[m.Sender.ID]
			strRequest := fmt.Sprintf("insert into requests (user_id, user_name, request_desc, price, completion_time) values (%d, '%s', '%s', %f, '%s')", m.Sender.ID, m.Sender.Username, req.RequestDescription, req.Price, req.CompletionDate)
			id := processDbInsertQuery(p, strRequest)
			log.Infof("id of latest request is %d", id)
			_, err := b.Send(m.Sender, fmt.Sprintf("Запрос %d успешно добавлен в базу", id))
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	b.Start()
}
