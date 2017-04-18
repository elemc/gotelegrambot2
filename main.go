// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	"flag"
	"sync"

	log "github.com/Sirupsen/logrus"
)

var (
	configName string
	wg         sync.WaitGroup
)

func init() {
	flag.StringVar(&configName, "config", "gotelegrambot", "configuration file name")
}

func main() {
	var err error
	flag.Parse()

	if err = LoadConfig(); err != nil {
		log.Fatalf("Unable to load configuration file %s: %s", configName, err)
	}
	if level, err := log.ParseLevel(options.LogLevel); err != nil {
		log.Warnf("Unable to parse log level %s: %s", options.LogLevel, err)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(level)
	}

	log.Warnf("Application started...")

	if err = InitDatabase(); err != nil {
		log.Fatalf("Unable to connect to database: %s", err)
	}

	wg.Add(1)
	go func() {
		if err = botServe(); err != nil {
			log.Errorf("Unable to serve bot: %s", err)
		}
	}()

	wg.Add(1)
	go func() {
		if err = httpServe(); err != nil {
			log.Errorf("Unable to serve HTTP server: %s", err)
		}
	}()

	wg.Wait()
}
