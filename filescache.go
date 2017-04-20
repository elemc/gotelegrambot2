// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Алексей Панов <a.panov@maximatelecom.ru> */
/* ------------------------------------------------ */

package main

import (
	"sync"

	log "github.com/Sirupsen/logrus"
)

type FilesCacheMemory struct {
	cache map[string]string
	mutex sync.RWMutex
}

func (fc *FilesCacheMemory) Get(name string) (value string) {
	var ok bool
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()
	if value, ok = fc.cache[name]; ok {
		return
	}
	return ""
}

func (fc *FilesCacheMemory) Set(name, value string) {
	fc.mutex.Lock()
	fc.cache[name] = value
	fc.mutex.Unlock()
}

func (fc *FilesCacheMemory) Update() {
	var (
		files []FileCache
		err   error
	)
	log.Debugf("Start update files cache...")

	if files, err = getFilesFromCache(); err != nil {
		log.Errorf("Unable to get files from cache for local memory cache store")
		return
	}
	for _, file := range files {
		fc.Set(file.FileID, file.FileName)
	}
	log.Debugf("Finish update files cache.")
}
