package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"io/ioutil"

	"net/http"
	"net/url"

	"encoding/base32"
	"encoding/json"

	"gopkg.in/telegram-bot-api.v4"
)

type Stream struct {
	Title    *string `json:title`
	Stream   *string `json:stream`
	Id       *string `json:id`
	Duration uint64  `json:duration`
	Cover    *string `json:cover`
}

func info(link string) (*Stream, error) {
	// Validate Link
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if (strings.Contains(u.Host, "youtube.com") ||
		strings.Contains(u.Host, "youtu.be")) == false {
		return nil, errors.New("not youtube")
	}

	// Get info
	id := base32.StdEncoding.EncodeToString([]byte(link))
	resp, err := http.Get("http://tubestreamer.ru/api/v1/info/" + id)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// Decode info
	var stream Stream
	e := json.Unmarshal(body, &stream)
	if e != nil {
		return nil, e
	}

	return &stream, nil
}

func worker(id int, bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			if update.InlineQuery == nil {
				continue
			} else {
				log.Printf("%d: [%s] %s", id, update.InlineQuery.From.UserName, update.InlineQuery.Query)

				var stream *Stream
				stream, err := info(update.InlineQuery.Query)
				if err == nil {
					reply := "http://tubestreamer.ru/stream/" + *stream.Id
					article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, *stream.Title, reply)
					article.ThumbURL = *stream.Cover
					inlineConf := tgbotapi.InlineConfig{
						InlineQueryID: update.InlineQuery.ID,
						IsPersonal:    false,
						Results:       []interface{}{article},
					}

					bot.AnswerInlineQuery(inlineConf)
				}
			}
		} else {
			log.Printf("%d: [%s] %s (not inline)", id, update.Message.From.UserName, update.Message.Text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "This is an inline bot.")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
}

func main() {
	var debug bool
	var workers uint
	var token string
	flag.BoolVar(&debug, "debug", false, "enable debug logs")
	flag.UintVar(&workers, "workers", 2, "workers count")
	flag.StringVar(&token, "token", os.Getenv("TELEGRAM_TOKEN"), "telegram bot token")
	flag.Parse()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for w := 1; w <= int(workers); w++ {
		go worker(w, bot, updates)
	}

	done := make(chan bool, 1)
	<-done
}
