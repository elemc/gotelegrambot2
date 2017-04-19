// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	bot        *tgbotapi.BotAPI
	photoCache map[int64]string
)

func botServe() (err error) {
	var (
		updates <-chan tgbotapi.Update
	)
	defer wg.Done()

	photoCache = make(map[int64]string)

	if bot, err = tgbotapi.NewBotAPI(options.APIKey); err != nil {
		return
	}
	bot.Debug = options.Debug
	log.Debug("Telegram bot initialized sucessful")

	go updatePhotoCache()

	updateOptions := tgbotapi.NewUpdate(0)
	updateOptions.Timeout = 60

	if updates, err = bot.GetUpdatesChan(updateOptions); err != nil {
		return
	}

	for update := range updates {
		log.Debugf("new update: %+v", *update.Message)
		if update.Message.Command() == "start" {
			continue
		}
		go func() {
			if err = saveMessage(update.Message); err != nil {
				log.Errorf("Unable to save message: %s", err)
			}
		}()
	}
	return
}

func saveMessage(msg *tgbotapi.Message) (err error) {
	// Files
	if msg.Audio != nil {
		go getFile(msg.Audio.FileID, msg.Chat.ID)
	}
	if msg.Document != nil {
		go getFile(msg.Document.FileID, msg.Chat.ID)
	}
	if msg.Photo != nil {
		for _, f := range *msg.Photo {
			go getFile(f.FileID, msg.Chat.ID)
		}
	}
	if msg.Sticker != nil {
		go getFile(msg.Sticker.FileID, msg.Chat.ID)
	}
	if msg.Video != nil {
		go getFile(msg.Video.FileID, msg.Chat.ID)
	}
	if msg.Voice != nil {
		go getFile(msg.Voice.FileID, msg.Chat.ID)
	}

	return dbSaveMessage(msg)
}

// getFile function for get file from telegram
func getFile(fileID string, chatID int64) {
	var (
		err error
		f   tgbotapi.File
	)

	fc := tgbotapi.FileConfig{}
	fc.FileID = fileID
	if f, err = bot.GetFile(fc); err != nil {
		log.Errorf("Unable to get file FileID [%s]: %s", fileID, err)
		return
	}

	// check directory
	dir := filepath.Dir(f.FilePath)
	path := filepath.Join(options.StaticDirPath, dir)
	if err = os.MkdirAll(path, 0755); err != nil {
		log.Errorf("Unable to make directories for FileID [%s]: %s", fileID, err)
		return
	}

	filename := filepath.Join(options.StaticDirPath, f.FilePath)
	if err = downloadImage(f.Link(options.APIKey), filename); err != nil {
		log.Errorf("Unable to download file for FileID [%s]: %s", fileID, err)
		return
	}
	log.Debugf("File downloaded for FileID [%s] in %s", fileID, filename)
}

func downloadImage(url string, filename string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return
	}
	return
}

func updatePhotoCache() {
	getUsers()
}
