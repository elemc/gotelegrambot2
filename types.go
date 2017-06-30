// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name> 			*/
/* ------------------------------------------------ */

package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
)

// Message is a type from tgbotapi with little changes for store messages in database
type Message struct {
	MessageID             int                       `json:"message_id"`
	UserFrom              *tgbotapi.User            `json:"from"` // optional
	Date                  int                       `json:"date"`
	Chat                  *tgbotapi.Chat            `json:"chat"`
	ForwardFrom           *tgbotapi.User            `json:"forward_from"`            // optional
	ForwardFromChat       *tgbotapi.Chat            `json:"forward_from_chat"`       // optional
	ForwardDate           int                       `json:"forward_date"`            // optional
	ReplyToMessage        *Message                  `json:"reply_to_message"`        // optional
	EditDate              int                       `json:"edit_date"`               // optional
	Text                  string                    `json:"text"`                    // optional
	Entities              *[]tgbotapi.MessageEntity `json:"entities"`                // optional
	Audio                 *tgbotapi.Audio           `json:"audio"`                   // optional
	Document              *tgbotapi.Document        `json:"document"`                // optional
	Photo                 *[]tgbotapi.PhotoSize     `json:"photo"`                   // optional
	Sticker               *tgbotapi.Sticker         `json:"sticker"`                 // optional
	Video                 *tgbotapi.Video           `json:"video"`                   // optional
	Voice                 *tgbotapi.Voice           `json:"voice"`                   // optional
	Caption               string                    `json:"caption"`                 // optional
	Contact               *tgbotapi.Contact         `json:"contact"`                 // optional
	Location              *tgbotapi.Location        `json:"location"`                // optional
	Venue                 *tgbotapi.Venue           `json:"venue"`                   // optional
	NewChatMember         *tgbotapi.User            `json:"new_chat_member"`         // optional
	LeftChatMember        *tgbotapi.User            `json:"left_chat_member"`        // optional
	NewChatTitle          string                    `json:"new_chat_title"`          // optional
	NewChatPhoto          *[]tgbotapi.PhotoSize     `json:"new_chat_photo"`          // optional
	DeleteChatPhoto       bool                      `json:"delete_chat_photo"`       // optional
	GroupChatCreated      bool                      `json:"group_chat_created"`      // optional
	SuperGroupChatCreated bool                      `json:"supergroup_chat_created"` // optional
	ChannelChatCreated    bool                      `json:"channel_chat_created"`    // optional
	MigrateToChatID       int64                     `json:"migrate_to_chat_id"`      // optional
	MigrateFromChatID     int64                     `json:"migrate_from_chat_id"`    // optional
	PinnedMessage         *Message                  `json:"pinned_message"`          // optional
}

// Chat is a type from tgbotapi with little changes for store chats in database
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`      // optional
	UserName  string `json:"username"`   // optional
	FirstName string `json:"first_name"` // optional
	LastName  string `json:"last_name"`  // optional
}

func convertMessage(m *tgbotapi.Message) (msg *Message) {
	if m == nil {
		return nil
	}
	var (
		data []byte
		err  error
	)
	if data, err = json.Marshal(m); err != nil {
		log.Errorf("Unable to marshal bot message [%+v]: %s", *m, err)
		return nil
	}
	msg = new(Message)
	if err = json.Unmarshal(data, msg); err != nil {
		log.Errorf("Unable to unmarshal bot message [%s]: %s", data, err)
		return nil
	}
	return
}
