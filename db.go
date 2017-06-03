// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name> 			*/
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

// FileCache type for store file cache in database
type FileCache struct {
	FileID   string `sql:",pk"`
	FileName string
}

// Flooder type for store flood level in database
type Flooder struct {
	UserID int `sql:",pk"`
	Level  int
}

// Feeder type for store RSS/Atom feeds in database
type Feeder struct {
	URL  string `sql:",pk"`
	Name string
}

// FeedNews type for store news from feeds in database
type FeedNews struct {
	URL         string `sql:",pk"`
	GUID        string `sql:",pk"`
	Title       string
	Link        string
	ImageURL    string
	ImageTitle  string
	Description string
	FeedTitle   string
}

// InsultWord type for store insult target and words in database
type InsultWord struct {
	Word   string `sql:",pk"`
	IsWord bool
}

var (
	db *pg.DB

	// ErrorFeedAlreadyExists is a generic error for feed already exists in database message
	ErrorFeedAlreadyExists = fmt.Errorf("feed already exists in database")

	// ErrorWordAlreadyExists is a generic error for insult word or target already exists in database message
	ErrorWordAlreadyExists = fmt.Errorf("insult word or target already exists in database")

	// ErrorFeedNotFound is a generic error for a feed not found in database message
	ErrorFeedNotFound = fmt.Errorf("feed not found in database")

	// ErrorWordNotFound is a generic error for a insult word or target not found in database message
	ErrorWordNotFound = fmt.Errorf("insult word or target not found in database")
)

// InitDatabase function for initialize pgsql database
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
		&Flooder{},
		&Cache{},
		&Feeder{},
		&FeedNews{},
		&InsultWord{},
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

func dbSetFloodLevel(userID, level int) (err error) {
	flooder := Flooder{UserID: userID}
	if err = db.Select(&flooder); err != nil && err != pg.ErrNoRows {
		return
	} else if err == pg.ErrNoRows {
		flooder.Level = level
		if err = db.Insert(&flooder); err != nil {
			return
		}
		return
	}
	flooder.Level = level
	err = db.Update(&flooder)
	return
}

func dbAddFloodLevel(userID int) (currentLevel int, err error) {
	flooder := Flooder{UserID: userID}
	if err = db.Select(&flooder); err != nil && err != pg.ErrNoRows {
		return
	} else if err == pg.ErrNoRows {
		flooder.Level = 1
		if err = db.Insert(&flooder); err != nil {
			return
		}
		return 1, nil
	}

	flooder.Level++
	if err = db.Update(&flooder); err != nil {
		return
	}
	currentLevel = flooder.Level
	return
}

func dbGetFloodLevel(userID int) (level int, err error) {
	flooder := Flooder{UserID: userID}
	if err = db.Select(&flooder); err != nil && err != pg.ErrNoRows {
		return
	} else if err == pg.ErrNoRows {
		return 0, nil
	}

	level = flooder.Level
	return
}

func dbAddFeed(url string, name string) (err error) {
	if _, err = dbGetFeed(url); err != nil && err != pg.ErrNoRows {
		return
	}
	err = db.Insert(&Feeder{URL: url, Name: name})
	return
}

func dbGetFeed(url string) (feed Feeder, err error) {
	feed.URL = url
	err = db.Select(&feed)
	return
}

func dbDelFeed(url string) (err error) {
	var feed Feeder
	if feed, err = dbGetFeed(url); err != nil && err != pg.ErrNoRows {
		return
	} else if err == pg.ErrNoRows {
		return ErrorFeedNotFound
	}
	err = db.Delete(&feed)
	return
}

func dbGetAllFeeds() (feeds []Feeder, err error) {
	err = db.Model(&feeds).Select()
	return
}

func dbNewsFound(news FeedNews) bool {
	if err := db.Select(&news); err != nil && err != pg.ErrNoRows {
		log.Errorf("Unable to get feed news with URL=%s and GUID=%s: %s", news.URL, news.GUID, err)
	} else if err == pg.ErrNoRows {
		return false
	}
	return true
}

func dbNewsAdd(news FeedNews) (err error) {
	err = db.Insert(&news)
	return
}

func dbInsultFoundWordOrTarget(word string, isWord bool) bool {
	t := &InsultWord{Word: word, IsWord: isWord}
	if err := db.Select(t); err != nil && err == pg.ErrNoRows {
		return false
	}
	return true
}

func dbInsultAddWordOrTarget(word string, isWord bool) (err error) {
	if dbInsultFoundWordOrTarget(word, isWord) {
		return ErrorWordAlreadyExists
	}
	err = db.Insert(&InsultWord{Word: word, IsWord: isWord})

	return
}

func dbInsultGetWordsOrTargets(isWord bool) (list []string, err error) {
	var words []InsultWord
	if err = db.Model(&words).Select(); err != nil {
		return
	}
	for _, word := range words {
		if word.IsWord != isWord {
			continue
		}
		list = append(list, word.Word)
	}
	return
}

func dbInsultDelWordOrTarget(word string, isWord bool) (err error) {
	if !dbInsultFoundWordOrTarget(word, isWord) {
		return ErrorWordNotFound
	}

	err = db.Delete(&InsultWord{Word: word, IsWord: isWord})
	return
}
