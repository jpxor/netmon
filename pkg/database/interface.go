package database

import "time"

type Interface interface {
	Write(database string, datum DataPoint) error
	Flush(database string) error
}

type DataPoint struct {
	Name     string
	ClientID string
	Time     time.Time
	Feilds   map[string]interface{}
}
