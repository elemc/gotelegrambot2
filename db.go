// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	db *pg.DB
)

func InitDatabase() (err error) {
	var pgo *pg.Options

	if pgo, err = pg.ParseURL(options.PgSQLDSN); err != nil {
		return
	}
	log.Debugf("Try to connect to postgrsql server...")
	db = pg.Connect(pgo)
	err = createTables()
	return
}

func dbSaveMessage(msg *tgbotapi.Message) (err error) {
	temp := *msg.Chat
	if err = db.Select(&temp); err != nil && err == pg.ErrNoRows {
		if err = db.Insert(msg.Chat); err != nil {
			return
		}
	} else if err != nil {
		return
	}
	return db.Insert(convertMessage(msg))
}

func createTables() (err error) {
	tables := []interface{}{
		&Message{},
		&tgbotapi.Chat{},
	}

	for _, t := range tables {
		if err = db.CreateTable(t, &orm.CreateTableOptions{IfNotExists: true}); err != nil {
			return
		}
	}
	return
}

func getChats() (chats []*tgbotapi.Chat, err error) {
	err = db.Model(&chats).Select()
	return
}
