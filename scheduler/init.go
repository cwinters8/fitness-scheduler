package scheduler

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"fitness-scheduler/sessions"
)

func Init(url string, db *sql.DB) error {
	// load all reminders into memory
	reminders, err := sessions.GetReminders(db)
	if err != nil {
		return errors.New("failed to load reminders: " + err.Error())
	}

	// spin off goroutines for each reminder and fire a notify request at the appropriate time
	for _, reminder := range reminders {
		go func(r sessions.Reminder) {
			if r.Status != sessions.Complete {
				now := time.Now()
				if r.Time.Valid {
					// if the reminder time is in the past, fire immediately
					if now.After(r.Time.Time) {
						err := r.Remind(url, db)
						if err != nil {
							log.Println("failed to remind user: " + err.Error())
							return
						}
					} else {
						// otherwise, wait until the reminder time
						time.Sleep(r.Time.Time.Sub(now))
						err := r.Remind(url, db)
						if err != nil {
							log.Println("failed to remind user: " + err.Error())
							return
						}
					}
				} else if r.MinutesPrior != nil {
					// get the time the reminder should fire
					session, err := sessions.GetSession(r.SessionID, db)
					if err != nil {
						log.Println("failed to get session: " + err.Error())
						return
					}
					timeToRemind := session.Timestamp.Add(-time.Minute * time.Duration(*r.MinutesPrior))
					if now.After(timeToRemind) {
						err := r.Remind(url, db)
						if err != nil {
							log.Println("failed to remind user: " + err.Error())
							return
						}
					} else {
						// otherwise, wait until the reminder time
						time.Sleep(timeToRemind.Sub(now))
						err := r.Remind(url, db)
						if err != nil {
							log.Println("failed to remind user: " + err.Error())
							return
						}
					}
				}
			}
		}(reminder)
	}
	return nil
}
