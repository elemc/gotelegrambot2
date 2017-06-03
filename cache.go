// -*- Go -*-
/* ------------------------------------------------ */
/* Golang source                                    */
/* Author: Alexei Panov <me@elemc.name> 			*/
/* ------------------------------------------------ */

package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-pg/pg"
)

// Cache is a type for store flooder information
type Cache struct {
	FlooderID int
	UserID    int
	Timestamp time.Time
}

func cacheUpdate() {
	var (
		caches []Cache
		err    error
	)

	for {
		time.Sleep(options.CacheUpdatePeriod)
		if err = db.Model(&caches).Select(); err != nil {
			log.Errorf("Unable to get caches: %s", err)
			continue
		}

		for _, cache := range caches {
			if time.Since(cache.Timestamp) >= options.CacheDuration {
				if err = cache.Remove(); err != nil {
					log.Errorf("Unable to remove cache record: %s", err)
				}
			}
		}
	}
}

func cacheGet(flooderID, userID int) (exists bool, duration time.Duration, err error) {
	var caches []Cache
	if err = db.Model(&caches).Where("flooder_id = ? AND user_id = ?", flooderID, userID).Select(); err != nil && err == pg.ErrNoRows {
		return false, 0, nil
	} else if err != nil {
		return
	}

	exists = false
	for _, cache := range caches {
		exists = true
		d := time.Since(cache.Timestamp)
		if d < duration || duration == 0 {
			duration = d
		}
		if d >= options.CacheDuration {
			if err = cache.Remove(); err != nil {
				log.Errorf("Unable to delete cache record in cacheGet: %s", err)
			}
		}
	}

	return
}

func cacheSet(flooderID, userID int) (err error) {
	cache := Cache{
		FlooderID: flooderID,
		UserID:    userID,
		Timestamp: time.Now(),
	}
	err = db.Insert(&cache)
	return
}

// Remove function for remove cache flooder record from database
func (c *Cache) Remove() (err error) {
	if _, err = db.Model(&[]Cache{}).Where("flooder_id = ? AND user_id = ? AND timestamp = ?", c.FlooderID, c.UserID, c.Timestamp).Delete(); err != nil {
		return
	}
	log.Debugf("Cache record with flooder ID=%d and user ID=%d removed", c.FlooderID, c.UserID)
	return
}
