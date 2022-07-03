package main

import (
	"fmt"
	"imons/pkg/database"
	"os"
	"time"
)

func main() {
	fmt.Println("Running Network Status Monitor")

	config := Config{}
	parseArgs(&config)

	var err error
	var db database.Interface

	if config.Database.Host != "" {
		db, err = database.NewInfluxdb1(
			config.Database.Host,
			config.Database.User,
			config.Database.Pass,
		)
		if err != nil {
			fmt.Println("error: failed to connect to database @", config.Database.Host)
			fmt.Println(err)
			os.Exit(0)
		}
	}

	lastConnTest := time.Unix(0, 0)
	lastSpeedTest := time.Unix(0, 0)

	for {
		// give the system a chance to reconnect to the
		// network after waking up from hybernation/sleep
		if !config.Testing {
			time.Sleep(30 * time.Second)
		}

		now := time.Now().Local()
		var hasInternet bool

		if now.After(lastConnTest.Add(config.ConnectionInterval)) {
			fmt.Println(" > testing network connection...")
			result, err := testConnection(config)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("   Local Area Network: %s\n", statusString(result.Local))
				fmt.Printf("   Internet: %s\n", statusString(result.Remote))
				hasInternet = result.Remote

				if db != nil {
					db.Write(config.Database.Name, database.DataPoint{
						Name:     "Connection",
						ClientID: config.MacAddr,
						Time:     now,
						Feilds: map[string]interface{}{
							"LAN":      result.Local,
							"Internet": result.Remote,
						},
					})
				}
			}
			lastConnTest = now
		}

		if hasInternet && now.After(lastSpeedTest.Add(config.SpeedTestInterval)) {
			fmt.Println(" > testing network speed...")
			result, err := testSpeed(config)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("   Latency: %d ms, Download: %.0f mbps, Upload: %.0f mbps\n", result.Latency, result.Download, result.Upload)

				if db != nil {
					db.Write(config.Database.Name, database.DataPoint{
						Name:     "Speed",
						ClientID: config.MacAddr,
						Time:     now,
						Feilds: map[string]interface{}{
							"Latency":  result.Latency,
							"Upload":   result.Upload,
							"Download": result.Download,
						},
					})
				}
			}
			lastSpeedTest = now
		}

		// flush results
		if db != nil {
			db.Flush(config.Database.Name)
		}

		if config.Testing {
			break
		}

		// sleep until next test
		nextConnTest := lastConnTest.Add(config.ConnectionInterval)
		nextSpeedTest := lastSpeedTest.Add(config.SpeedTestInterval)

		wait1 := nextConnTest.Sub(now)
		wait2 := nextSpeedTest.Sub(now)
		time.Sleep(minDuration(wait1, wait2))
	}

}

func minDuration(t1, t2 time.Duration) time.Duration {
	if t1 < t2 {
		return t1
	} else {
		return t2
	}
}

func statusString(ok bool) string {
	if ok {
		return "ok"
	} else {
		return "no response"
	}
}
