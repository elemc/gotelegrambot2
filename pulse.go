// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import "github.com/mmcdole/gofeed"

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
