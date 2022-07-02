package main

import (
	"fmt"
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
)

type ConnTestResult struct {
	Local  bool
	Remote bool
}

func testConnection(config Config) (ConnTestResult, error) {
	result := ConnTestResult{}
	{
		var localAddresses []string
		if config.LocalAP != "" {
			localAddresses = []string{config.LocalAP}
		} else {
			localAddresses = defaultLocalAddresses
		}
		pingResults, err := pingAll(localAddresses)
		if err != nil {
			fmt.Println("failed local pingAll")
			return result, err
		}
		result.Local = atLeastOneTrue(pingResults)
	}
	{
		var remoteAddresses []string
		if config.Traceroute != nil {
			remoteAddresses = config.Traceroute

		} else if config.RemoteAP != "" {
			remoteAddresses = []string{config.RemoteAP}

		} else {
			remoteAddresses = defaultRemoteAddresses
		}
		pingResults, err := pingAll(remoteAddresses)
		if err != nil {
			fmt.Println("failed remote pingAll")
			return result, err
		}
		if config.Traceroute == nil {
			result.Remote = atLeastOneTrue(pingResults)
		} else {
			result.Remote = allTrue(pingResults)
			printTraceroute(config.Traceroute, pingResults)
		}
	}
	return result, nil
}

func printTraceroute(traceroute []string, results map[string]bool) {
	fmt.Println("Traceroute:")
	for hop, ip := range traceroute {
		fmt.Printf("%-2d %-15s %s\n", hop, ip, statusString(results[ip]))
	}
}

func pingAll(iplist []string) (map[string]bool, error) {
	p := fastping.NewPinger()
	pingResults := map[string]bool{}

	for _, ip := range iplist {
		ra, err := net.ResolveIPAddr("ip4:icmp", ip)
		if err != nil {
			fmt.Println(err)
			continue
		}
		p.AddIPAddr(ra)
		pingResults[ip] = false
	}

	wait := make(chan bool, len(iplist))

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		pingResults[addr.String()] = true
	}

	p.OnIdle = func() {
		close(wait)
	}

	err := p.Run()
	if err != nil {
		fmt.Println(err)
		return pingResults, err
	}

	<-wait
	return pingResults, nil
}

func atLeastOneTrue(in map[string]bool) (out bool) {
	for _, b := range in {
		if b {
			out = true
			break
		}
	}
	return out
}

func allTrue(in map[string]bool) (out bool) {
	out = true
	for _, b := range in {
		if !b {
			out = false
			break
		}
	}
	return out
}
