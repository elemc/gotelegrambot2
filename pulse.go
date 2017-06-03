// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name>				*/
/* ------------------------------------------------ */

package main

import (
	"bytes"
	"fmt"
	"html"
	"sync"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mmcdole/gofeed"
)

// FeedLocks is a type for locking feeders
type FeedLocks struct {
	locks map[string]bool
	mutex sync.Mutex
}

var (
	feedLocks = FeedLocks{locks: make(map[string]bool)}
)

func feedAdd(url string) (err error) {
	var feed *gofeed.Feed
	parser := gofeed.NewParser()
	if feed, err = parser.ParseURL(url); err != nil {
		return
	}
	err = dbAddFeed(url, feed.Title)
	return
}

func feedDel(url string) (err error) {
	err = dbDelFeed(url)
	return
}

func updateFeeds() {
	var (
		feeds []Feeder
		err   error
	)
	defer wg.Done()

	for {
		time.Sleep(options.FeedsUpdatePeriod)
		if feeds, err = dbGetAllFeeds(); err != nil {
			log.Errorf("Unable to get all feeds: %s", err)
			continue
		}
		for _, feed := range feeds {
			log.Debugf("Update feed %s (%s)", feed.Name, feed.URL)
			updateFeed(feed)
		}
	}
}

func updateFeed(feed Feeder) {
	var (
		fd  *gofeed.Feed
		err error
	)

	if feed.URL == "" {
		log.Warnf("Feeder with empty URL %s", feed.Name)
		return
	}

	if feedLocks.getFeedLock(feed) { // is locked, skip it
		log.Debugf("Skip feed %s, its running now", feed.URL)
		return
	}

	feedLocks.lockFeeder(feed)
	defer feedLocks.unlockFeeder(feed)

	parser := gofeed.NewParser()
	if fd, err = parser.ParseURL(feed.URL); err != nil {
		log.Errorf("Unable to parse feed URL [%s]: %s", feed.URL, err)
		return
	}
	if fd == nil {
		log.Warnf("Feeder for %s is nil", feed.URL)
		return
	}

	for _, item := range fd.Items {
		if item == nil {
			log.Warnf("Item for feeder %s is nil", feed.URL)
			continue
		}
		news := feedNewsFromItem(fd, item)

		if news.URL == "" {
			log.Warnf("News with empty URL for feed %s", feed.URL)
			continue
		}
		if dbNewsFound(news) { // this news is found, skip it
			continue
		}

		if err = dbNewsAdd(news); err != nil {
			log.Errorf("Unable to insert news to database: %s", err)
			continue
		}
		sendMessageToAllChats(news.String())
	}
}

func feedNewsFromItem(fd *gofeed.Feed, item *gofeed.Item) FeedNews {
	if fd == nil || item == nil {
		return FeedNews{}
	}
	news := FeedNews{
		URL:         fd.Link,
		GUID:        item.GUID,
		Title:       item.Title,
		Link:        item.Link,
		Description: html.EscapeString(item.Description),
		FeedTitle:   fd.Title,
	}
	if item.Image != nil {
		news.ImageURL = item.Image.URL
		news.ImageTitle = item.Image.Title
	}
	return news
}

func (fn *FeedNews) String() string {
	tmplText := `*{{ .FeedTitle }}*
[{{ .Title }}]({{ .Link }})`
	tmpl, err := template.New("message").Parse(tmplText)
	if err != nil {
		log.Errorf("Unable to parse template: %s", err)
		return ""
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, fn); err != nil {
		log.Errorf("Unable to execute template: %s", err)
	}

	text := buf.String()
	if fn.ImageURL != "" {
		text = fmt.Sprintf("%s\n[%s](%s)", text, fn.ImageTitle, fn.ImageURL)
	}

	return text
}

func (fl *FeedLocks) getFeedLock(feed Feeder) (result bool) {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()
	var ok bool
	if result, ok = fl.locks[feed.URL]; !ok {
		return false
	}
	return
}

func (fl *FeedLocks) lockUnlockFeeder(feed Feeder, lock bool) {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()
	fl.locks[feed.URL] = lock
}

func (fl *FeedLocks) lockFeeder(feed Feeder) {
	fl.lockUnlockFeeder(feed, true)
}

func (fl *FeedLocks) unlockFeeder(feed Feeder) {
	fl.lockUnlockFeeder(feed, false)
}
