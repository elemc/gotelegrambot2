// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

type Options struct {
	APIKey        string
	PgSQLDSN      string
	LogLevel      string
	ServerAddr    string
	Debug         bool
	StaticDirPath string

	// TODO: remove it, only for converter
	CouchbaseCluster      string
	CouchbaseBucketName   string
	CouchbaseBucketSecret string
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
		APIKey:        viper.GetString("main.api_key"),
		PgSQLDSN:      viper.GetString("pgsql.dsn"),
		LogLevel:      viper.GetString("log.level"),
		ServerAddr:    viper.GetString("http.addr"),
		Debug:         viper.GetBool("main.debug"),
		StaticDirPath: viper.GetString("main.static_path"),

		// TODO: remove it, only for converter, Couchbase
		CouchbaseCluster:      viper.GetString("couchbase.cluster"),
		CouchbaseBucketName:   viper.GetString("couchbase.bucket_name"),
		CouchbaseBucketSecret: viper.GetString("couchbase.bucket_secret"),
	}
	return
}
