// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"gopkg.in/telegram-bot-api.v4"
)

type FileCache struct {
	FileID   string `sql:",pk"`
	FileName string
}

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

func dbSaveChat(chat *tgbotapi.Chat) (err error) {
	tempChat := &tgbotapi.Chat{ID: chat.ID}
	if err = db.Select(tempChat); err != nil && err == pg.ErrNoRows {
		return db.Insert(chat)
	} else if err != nil {
		return
	}
	return
}

func dbSaveUser(user *tgbotapi.User) (err error) {
	tempUser := &tgbotapi.User{ID: user.ID}
	if err = db.Select(tempUser); err != nil && err == pg.ErrNoRows {
		return db.Insert(user)
	} else if err != nil {
		return
	}
	return
}

func dbSaveMessage(msg *tgbotapi.Message) (err error) {
	if err = dbSaveChat(msg.Chat); err != nil {
		return
	}
	if err = dbSaveUser(msg.From); err != nil {
		return
	}

	return db.Insert(convertMessage(msg))
}

func createTables() (err error) {
	tables := []interface{}{
		&Message{},
		&tgbotapi.Chat{},
		&tgbotapi.User{},
		&FileCache{},
	}

	for _, t := range tables {
		if err = db.CreateTable(t, &orm.CreateTableOptions{IfNotExists: true}); err != nil {
			return
		}
	}
	return
}

func getChats() (chats []tgbotapi.Chat, err error) {
	err = db.Model(&chats).Select()
	return
}

func getChatYears(chatID int64) (years []string, err error) {
	var intyears []int
	if _, err = db.Query(&intyears, `SELECT date_part('year', to_timestamp("date")) FROM messages WHERE chat @> '{"id": ?}'`, chatID); err != nil {
		return
	}
	sort.Ints(intyears)

	for _, iy := range intyears {
		years = appendStringToSliceIfNotFound(years, fmt.Sprintf("%d", iy))
	}
	return
}

func getChatMonths(chatID int64, year int) (months []string, err error) {
	var intmonths []int
	if _, err = db.Query(&intmonths, `SELECT date_part('month', to_timestamp("date")) FROM messages WHERE chat @> '{"id": ?}' AND date_part('year', to_timestamp("date")) = ?`, chatID, year); err != nil {
		return
	}
	sort.Ints(intmonths)

	for _, im := range intmonths {
		months = appendStringToSliceIfNotFound(months, fmt.Sprintf("%02d", im))
	}
	return
}

func getChatDays(chatID int64, year, month int) (days []string, err error) {
	var intdays []int
	if _, err = db.Query(&intdays, `SELECT date_part('day', to_timestamp("date")) FROM messages WHERE chat @> '{"id": ?}' AND date_part('year', to_timestamp("date")) = ? AND date_part('month', to_timestamp("date")) = ?`, chatID, year, month); err != nil {
		return
	}
	sort.Ints(intdays)

	for _, id := range intdays {
		days = appendStringToSliceIfNotFound(days, fmt.Sprintf("%02d", id))
	}
	return
}

func getMessages(chatID int64, year, month, day int) (msgs []Message, err error) {
	beginTime := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).Unix()
	endTime := time.Date(year, time.Month(month), day, 23, 59, 59, 100, time.Local).Unix()

	if err = db.Model(&msgs).Order("date").Where("date >= ? AND date <= ? AND chat @> '{\"id\": ?}'", beginTime, endTime, chatID).Select(); err != nil {
		return
	}
	return
}

func getUsers() (users []tgbotapi.User, err error) {
	if err = db.Model(&users).Select(); err != nil {
		log.Error(err)
		return
	}
	return
}

func getUser(name string) (user *tgbotapi.User, err error) {
	if name == "" {
		return nil, fmt.Errorf("user name is empty")
	}
	tuser := []tgbotapi.User{}
	if name[0] == '@' {
		name = name[1:]
	}
	if strings.Contains(name, " ") {
		fl := strings.Split(name, " ")
		first := fl[0]
		last := fl[1]
		if err = db.Model(&tuser).Where("first_name = ? AND last_name = ?", first, last).WhereOr("first_name = ? AND last_name = ?", last, first).Select(); err != nil {
			if err == pg.ErrNoRows {
				return nil, ErrorUserNotFound
			}
			return
		}
	} else {
		if err = db.Model(&tuser).Where("first_name = ?", name).WhereOr("last_name = ?", name).WhereOr("user_name = ?", name).Select(); err != nil {
			if err == pg.ErrNoRows {
				return nil, ErrorUserNotFound
			}
			return
		}
	}
	if len(tuser) > 1 {
		var us []string
		for _, u := range tuser {
			us = append(us, fmt.Sprintf("@%s (%s %s)", u.UserName, u.FirstName, u.LastName))
		}
		text := fmt.Sprintf("``` Список: \n\t%s ```", strings.Join(us, "\n\t"))
		log.Warn(text)
		return nil, fmt.Errorf("%s", text)
	}

	return &tuser[0], nil
}

func getFileFromCache(fileID string) (file FileCache, err error) {
	file.FileID = fileID
	err = db.Select(&file)
	return
}

func dbSaveFileToCahce(fileID, filename string) (err error) {
	f := FileCache{
		FileID:   fileID,
		FileName: filename,
	}
	filesCache.Set(fileID, filename)
	err = db.Insert(&f)
	return
}

func getFilesFromCache() (files []FileCache, err error) {
	if err = db.Model(&files).Select(); err != nil {
		return
	}
	return
}
