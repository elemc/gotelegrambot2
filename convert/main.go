// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	"log"
	"sync"

	"gopkg.in/telegram-bot-api.v4"

	botdb "github.com/elemc/gotelegrambot/db"
)

const (
	configName = "gotelegrambot"
)

var (
	wg sync.WaitGroup
)

func main() {
	var (
		err error
	)

	log.Printf("Try to load configuration...")
	if err = LoadConfig(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Try to connect to couchbase...")
	botdb.InitCouchbase(options.CouchbaseCluster, options.CouchbaseBucketName, options.CouchbaseBucketSecret)

	log.Printf("Try to connect to pgsql...")
	if err = InitDatabase(); err != nil {
		log.Fatal(err)
	}

	if bot, err = tgbotapi.NewBotAPI(options.APIKey); err != nil {
		return
	}
	bot.Debug = options.Debug
	log.Print("Telegram bot initialized sucessful")

	var chats []*tgbotapi.Chat
	if chats, err = botdb.GetChats(); err != nil {
		log.Fatal(err)
	}

	for _, chat := range chats {
		if chat.ID < 0 {
			chat.Type = "group"
		} else {
			chat.Type = "private"
		}
		if err = dbSaveChat(chat); err != nil {
			log.Printf("Unable to save chat with ID %d: %s", chat.ID, err)
			continue
		}

		var msgs []*tgbotapi.Message
		if msgs, err = botdb.GetMessages(chat.ID); err != nil {
			log.Printf("Unable to get messages for chat ID %d: %s", chat.ID, err)
			continue
		}
		for _, msg := range msgs {
			if err = dbSaveUser(msg.From); err != nil {
				log.Printf("Unable to save user %s: %s", msg.From.String(), err)
				continue
			}

			if err = saveMessage(msg); err != nil {
				log.Printf("Unable to save message %s: %s", msg.Text, err)
				continue
			}
		}
	}
}
