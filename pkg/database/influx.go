package database

import (
	"fmt"

	db1 "github.com/influxdata/influxdb1-client/v2"
)

type InfluxDatabase struct {
	client  db1.Client
	batches map[string]db1.BatchPoints
}

func NewInfluxdb1(host, user, pass string) (Interface, error) {
	client, err := db1.NewHTTPClient(
		db1.HTTPConfig{
			Addr:     host,
			Username: user,
			Password: pass,
		})
	if err != nil {
		fmt.Println("Error: failed to create InfluxDB Client: ", err.Error())
		return nil, err
	}
	return &InfluxDatabase{
		client:  client,
		batches: map[string]db1.BatchPoints{},
	}, nil
}

func (db *InfluxDatabase) Write(database string, datum DataPoint) error {
	if db.batches[database] == nil {
		batch, err := db1.NewBatchPoints(db1.BatchPointsConfig{
			Precision: "ms",
			Database:  database,
		})
		if err != nil {
			fmt.Println("error: influxdb failed to create batch object: ", err.Error())
			return err
		}
		db.batches[database] = batch
	}
	pt, err := db1.NewPoint(
		datum.Name,
		map[string]string{
			"MAC": datum.ClientID,
		},
		datum.Feilds,
		datum.Time,
	)
	if err != nil {
		fmt.Println("Error: failed to create InfluxDB data point: ", err.Error())
		return err
	}
	db.batches[database].AddPoint(pt)
	return nil
}

func (db *InfluxDatabase) Flush(database string) error {
	if db.batches[database] != nil {
		err := db.client.Write(db.batches[database])
		if err != nil {
			fmt.Println("error: influxdb failed to write during flush: ", err.Error())
			return err
		}
		db.batches[database] = nil
	}
	return nil
}
