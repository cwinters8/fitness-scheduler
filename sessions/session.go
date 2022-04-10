package sessions

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Session struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"` // assuming user ID is int64 because it is probably a primary key in a DB
	Title     string     `json:"title"`
	Routine   *Routine   `json:"routine,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
	Duration  int64      `json:"duration"` // in minutes
	Frequency Frequency  `json:"frequency"`
	Reminders []Reminder `json:"reminders"`
	Notes     string     `json:"notes"`
}

func (s *Session) Save(db *sql.DB) error {
	if s.Routine != nil && s.Routine.ID == 0 {
		err := s.Routine.Save(db)
		if err != nil {
			return errors.New("failed to save routine: " + err.Error())
		}
	}
	if s.Routine == nil {
		s.Routine = &Routine{} // handles nil case so we can use the ID field with a zero value
	}

	if s.Frequency.ID == 0 {
		err := s.Frequency.Save(db)
		if err != nil {
			return errors.New("failed to save frequency: " + err.Error())
		}
	}

	if s.Title == "" {
		s.Title = s.Routine.Name
	}

	if s.Duration == 0 {
		s.Duration = s.Routine.Duration
	}

	query := "insert into sessions (user_id, title, routine_id, timestamp, duration, frequency_id, notes) values (?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, s.UserID, s.Title, s.Routine.ID, s.Timestamp, s.Duration, s.Frequency.ID, s.Notes)
	if err != nil {
		return errors.New("failed to save session: " + err.Error())
	}
	s.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}

	for _, r := range s.Reminders {
		r.SessionID = s.ID
		err := r.Save(db)
		if err != nil {
			return errors.New("failed to save reminder: " + err.Error())
		}
	}

	return nil
}

func GetSession(id int64, db *sql.DB) (*Session, error) {
	query := "select user_id, title, routine_id, timestamp, duration, frequency_id, notes from sessions where id = ?"
	var session Session
	session.ID = id
	session.Routine = &Routine{}
	err := db.QueryRow(query, id).Scan(&session.UserID, &session.Title, &session.Routine.ID, &session.Timestamp, &session.Duration, &session.Frequency.ID, &session.Notes)
	if err != nil {
		return nil, errors.New("failed to select session from database: " + err.Error())
	}
	// get routine
	routine, err := GetRoutine(session.Routine.ID, db)
	if err != nil {
		return nil, errors.New("failed to get routine: " + err.Error())
	}
	session.Routine = routine

	// get frequency
	frequency, err := GetFrequency(session.Frequency.ID, db)
	if err != nil {
		return nil, errors.New("failed to get frequency: " + err.Error())
	}
	session.Frequency = *frequency

	// get reminders
	reminders, err := GetRemindersBySession(session.ID, db)
	if err != nil {
		return nil, errors.New("failed to get reminders: " + err.Error())
	}
	session.Reminders = reminders

	return &session, nil
}

// assuming a routine is comprised of the following fields based on the provided concept
type Routine struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Category       Category   `json:"category"`
	Description    string     `json:"description"`
	URL            string     `json:"url"`
	Duration       int64      `json:"duration"` // in minutes
	Votes          int64      `json:"votes"`
	UserID         int64      `json:"user_id"`
	Public         bool       `json:"public"`
	Views          int64      `json:"views"`
	TimesCompleted int64      `json:"times_completed"`
	Created        *time.Time `json:"created"`
	Modified       nullTime   `json:"modified,omitempty"`
}

func (r *Routine) Save(db *sql.DB) error {
	if r.Created == nil {
		now := time.Now()
		r.Created = &now
	}
	query := "insert into routines (name, category, description, url, duration, votes, user_id, public, views, times_completed, created) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, r.Name, r.Category, r.Description, r.URL, r.Duration, r.Votes, r.UserID, r.Public, r.Views, r.TimesCompleted, *r.Created)
	if err != nil {
		return errors.New("failed to save routine: " + err.Error())
	}
	r.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}
	return nil
}

func GetRoutine(id int64, db *sql.DB) (*Routine, error) {
	query := "select name, category, description, url, duration, votes, user_id, public, views, times_completed, created, modified from routines where id = ?"
	var routine Routine
	routine.ID = id
	err := db.QueryRow(query, id).Scan(&routine.Name, &routine.Category, &routine.Description, &routine.URL, &routine.Duration, &routine.Votes, &routine.UserID, &routine.Public, &routine.Views, &routine.TimesCompleted, &routine.Created, &routine.Modified)
	if err != nil {
		return nil, errors.New("failed to select routine from database: " + err.Error())
	}
	return &routine, nil
}

type Category string

const (
	Official  Category = "official"
	Community Category = "community"
)

type Frequency struct {
	ID        int64         `json:"id"`
	StartDate nullDate      `json:"start_date,omitempty"`
	EndDate   nullDate      `json:"end_date,omitempty"`
	Type      FrequencyType `json:"type"`
	Days      []int64       `json:"day,omitempty"` // can be weekdays or days of the month
}

func (f *Frequency) Save(db *sql.DB) error {
	query := "insert into frequencies (start_date, end_date, type) values (?, ?, ?)"
	result, err := db.Exec(query, f.StartDate, f.EndDate, f.Type)
	if err != nil {
		return errors.New("failed to save frequency: " + err.Error())
	}
	f.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}

	for _, d := range f.Days {
		query := "insert into frequency_days (frequency_id, day) values (?, ?)"
		_, err := db.Exec(query, f.ID, d)
		if err != nil {
			return errors.New("failed to save frequency day: " + err.Error())
		}
	}

	return nil
}

func GetFrequency(id int64, db *sql.DB) (*Frequency, error) {
	query := "select start_date, end_date, type from frequencies where id = ?"
	var frequency Frequency
	frequency.ID = id
	err := db.QueryRow(query, id).Scan(&frequency.StartDate, &frequency.EndDate, &frequency.Type)
	if err != nil {
		return nil, errors.New("failed to select frequency from database: " + err.Error())
	}

	query = "select day from frequency_days where frequency_id = ?"
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, errors.New("failed to select frequency days from database: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var day int64
		err := rows.Scan(&day)
		if err != nil {
			return nil, errors.New("failed to scan frequency day: " + err.Error())
		}
		frequency.Days = append(frequency.Days, day)
	}

	return &frequency, nil
}

type nullDate struct {
	sql.NullTime
}

func (d *nullDate) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return errors.New("failed to unmarshal date: " + err.Error())
	}
	if len(s) > 0 {
		parsedTime, err := time.Parse("2006-01-02", s)
		if err != nil {
			return errors.New("failed to parse date: " + err.Error())
		}
		d.Valid = true
		d.Time = parsedTime
	} else {
		d.Valid = false
	}
	return nil
}

func (d nullDate) MarshalJSON() ([]byte, error) {
	if d.Valid {
		return json.Marshal(d.Time.Format("2006-01-02"))
	}
	return json.Marshal(nil)
}

type nullTime struct {
	sql.NullTime
}

func (t *nullTime) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return errors.New("failed to unmarshal date: " + err.Error())
	}
	if len(s) > 0 {
		parsedTime, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return errors.New("failed to parse date: " + err.Error())
		}
		t.Valid = true
		t.Time = parsedTime
	} else {
		t.Valid = false
	}
	return nil
}

func (t nullTime) MarshalJSON() ([]byte, error) {
	if t.Valid {
		return json.Marshal(t.Time.Format(time.RFC3339))
	}
	return json.Marshal(nil)
}

type FrequencyType string

const (
	Single  FrequencyType = "single"
	Daily   FrequencyType = "daily"
	Weekly  FrequencyType = "weekly"
	Monthly FrequencyType = "monthly"
)

type Weekday int64

const (
	Monday    Weekday = 0
	Tuesday   Weekday = 1
	Wednesday Weekday = 2
	Thursday  Weekday = 3
	Friday    Weekday = 4
	Saturday  Weekday = 5
	Sunday    Weekday = 6
)

type Reminder struct {
	ID           int64          `json:"id"`
	SessionID    int64          `json:"-"` // no need to send this to the client as part of the session data
	Time         nullTime       `json:"time,omitempty"`
	MinutesPrior *int64         `json:"minutes_prior,omitempty"`
	Status       ReminderStatus `json:"status"`
}

func (r *Reminder) Save(db *sql.DB) error {
	if r.Status == "" {
		r.Status = Pending
	}
	query := "insert into reminders (session_id, time, minutes_prior, status) values (?, ?, ?, ?)"
	result, err := db.Exec(query, r.SessionID, r.Time, *r.MinutesPrior, r.Status)
	if err != nil {
		return errors.New("failed to save reminder: " + err.Error())
	}
	r.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}
	return nil
}

func (r *Reminder) Remind(url string, db *sql.DB) error {
	session, err := GetSession(r.SessionID, db)
	if err != nil {
		return errors.New("failed to get session: " + err.Error())
	}

	// send the associated session to /notify
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(session)
	if err != nil {
		return errors.New("failed to encode session: " + err.Error())
	}
	_, err = http.Post(url+"/notify", "application/json", &buf)
	if err != nil {
		return errors.New("failed to notify: " + err.Error())
	}

	// mark the reminder complete if it is a one time reminder
	if r.Time.Valid && r.MinutesPrior == nil {
		r.Status = Complete
		query := "update reminders set status = ? where id = ?"
		_, err = db.Exec(query, r.Status, r.ID)
		if err != nil {
			return errors.New("failed to update reminder: " + err.Error())
		}
	}
	return nil
}

func GetReminders(db *sql.DB) ([]Reminder, error) {
	query := "select id, session_id, time, minutes_prior, status from reminders"
	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.New("failed to select reminders from database: " + err.Error())
	}
	defer rows.Close()
	var reminders []Reminder
	for rows.Next() {
		var reminder Reminder
		err := rows.Scan(&reminder.ID, &reminder.SessionID, &reminder.Time, &reminder.MinutesPrior, &reminder.Status)
		if err != nil {
			return nil, errors.New("failed to scan reminder: " + err.Error())
		}
		reminders = append(reminders, reminder)
	}
	return reminders, nil
}

func GetRemindersBySession(sessionID int64, db *sql.DB) ([]Reminder, error) {
	query := "select id, time, minutes_prior, status from reminders where session_id = ?"
	rows, err := db.Query(query, sessionID)
	if err != nil {
		return nil, errors.New("failed to select reminders from database: " + err.Error())
	}
	defer rows.Close()
	var reminders []Reminder
	for rows.Next() {
		var reminder Reminder
		err := rows.Scan(&reminder.ID, &reminder.Time, &reminder.MinutesPrior, &reminder.Status)
		if err != nil {
			return nil, errors.New("failed to scan reminder: " + err.Error())
		}
		reminders = append(reminders, reminder)
	}
	return reminders, nil
}

type ReminderStatus string

const (
	Pending  ReminderStatus = "pending"
	Complete ReminderStatus = "complete"
)
