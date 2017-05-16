// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name> 			*/
/* ------------------------------------------------ */

package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

type Options struct {
	APIKey            string
	PgSQLDSN          string
	LogLevel          string
	ServerAddr        string
	Debug             bool
	StaticDirPath     string
	MaximumFloodLevel int

	CacheDuration     time.Duration
	CacheUpdatePeriod time.Duration
}

var options *Options

func LoadConfig() (err error) {
	log.Warnf("Load configuration file...")

	viper.SetConfigName(configName)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("/usr/local/etc")

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	options = &Options{
		APIKey:            viper.GetString("main.api_key"),
		PgSQLDSN:          viper.GetString("pgsql.dsn"),
		LogLevel:          viper.GetString("log.level"),
		ServerAddr:        viper.GetString("http.addr"),
		Debug:             viper.GetBool("main.debug"),
		StaticDirPath:     viper.GetString("main.static_path"),
		MaximumFloodLevel: viper.GetInt("main.maximum_flood_level"),
		CacheDuration:     viper.GetDuration("cache.duration"),
		CacheUpdatePeriod: viper.GetDuration("cache.update_period"),
	}
	return
}
