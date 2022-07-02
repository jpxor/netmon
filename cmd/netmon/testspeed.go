package main

import (
	"fmt"

	"github.com/showwin/speedtest-go/speedtest"
)

type SpeedTestResult struct {
	Latency  int64   // ms
	Upload   float64 // mbps
	Download float64 // mbps
}

func testSpeed(config Config) (SpeedTestResult, error) {
	result := SpeedTestResult{}

	user, err := speedtest.FetchUserInfo()
	if err != nil {
		fmt.Println("error: failed to fetch user info for speedtest")
		return result, err
	}
	serverList, err := speedtest.FetchServers(user)
	if err != nil {
		fmt.Println("error: failed to fetch server info for speedtest")
		return result, err
	}
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		fmt.Println("error: failed to fetch server info for speedtest")
		return result, err
	}
	for _, s := range targets {
		s.PingTest()
		s.DownloadTest(false)
		s.UploadTest(false)

		result.Latency = s.Latency.Milliseconds()
		result.Upload = s.ULSpeed
		result.Download = s.DLSpeed
		break
	}
	return result, nil
}
