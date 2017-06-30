// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name> 			*/
/* ------------------------------------------------ */

package main

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// FilesCacheMemory type is a thread-safe files cache
type FilesCacheMemory struct {
	cache map[string]string
	mutex sync.RWMutex
}

// Get function get file from cache by name
func (fc *FilesCacheMemory) Get(name string) (value string) {
	var ok bool
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()
	if value, ok = fc.cache[name]; ok {
		return
	}
	return ""
}

// Set function store file information in cache
func (fc *FilesCacheMemory) Set(name, value string) {
	fc.mutex.Lock()
	fc.cache[name] = value
	fc.mutex.Unlock()
}

// Update function updates files in cache
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
