package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var defaultLocalAddresses = []string{
	"10.0.1.1",
	"192.168.0.1",
	"192.168.1.1",
	"192.168.2.1",
	"192.168.3.1",
	"192.168.50.1",
}

var defaultRemoteAddresses = []string{
	"208.67.222.222", // Open DNS
	"8.8.8.8",        // Google DNS
	"1.1.1.1",        // Cloudflare DNS
}

type Config struct {
	Testing    bool
	MacAddr    string
	Traceroute []string
	LocalAP    string
	RemoteAP   string
	Database   struct {
		Name string
		Host string
		User string
		Pass string
	}
	ConnectionInterval time.Duration
	SpeedTestInterval  time.Duration
}

func parseArgs(config *Config) {
	arg_traceroute := flag.String("traceroute", "", "A list of IP addresses obtained from 'traceroute'. Format: \"IP|IP|IP|IP\"")
	arg_localap := flag.String("local", "", "Address of a server or router on the local area network for checking LAN status")
	arg_remoteap := flag.String("remote", "", "Address of a remote server for checking internet status")
	arg_dbhost := flag.String("dbhost", `http://192.168.50.50:8086`, "Address of an influxdb for publishing data")
	arg_dbuser := flag.String("dbuser", "netmon", "influxdb username")
	arg_dbpass := flag.String("dbpass", "netmon", "influxdb password")
	arg_dbname := flag.String("dbname", "netmon", "influxdb database name")
	arg_interval := flag.String("interval", "5", "number of minutes between each connection test")
	arg_speedtest := flag.String("speedtest", "60", "number of minutes between each speed test")
	arg_mac := flag.String("mac", "", "set MAC address as unique client identifier")
	arg_devtest := flag.Bool("test", false, "runs once and then exits")
	flag.Parse()

	// -traceroute="192.168.50.1|38.147.245.129|38.108.78.198|38.147.246.129|38.122.69.49|38.88.240.186|1.1.1.1"
	if *arg_traceroute != "" {
		config.Traceroute = strings.Split(*arg_traceroute, "|")
	}

	if *arg_localap != "" {
		config.LocalAP = *arg_localap
	}

	if *arg_remoteap != "" {
		config.RemoteAP = *arg_remoteap
	}

	if *arg_dbhost != "" {
		config.Database.Host = *arg_dbhost
	}

	if *arg_dbname != "" {
		config.Database.Name = *arg_dbname
	}

	if *arg_dbuser != "" {
		config.Database.User = *arg_dbuser
	}

	if *arg_dbpass != "" {
		config.Database.Pass = *arg_dbpass
	}

	if *arg_dbpass != "" {
		config.Database.Pass = *arg_dbpass
	}

	if *arg_interval != "" {
		nmins, err := strconv.Atoi(*arg_interval)
		if err != nil {
			fmt.Printf("cannot convert interval '%s' to integer\n", *arg_interval)
			fmt.Println(err)
		}
		config.ConnectionInterval = time.Duration(nmins) * time.Minute
	}

	if *arg_speedtest != "" {
		nmins, err := strconv.Atoi(*arg_speedtest)
		if err != nil {
			fmt.Printf("cannot convert speedtest '%s' to integer\n", *arg_speedtest)
			fmt.Println(err)
		}
		config.SpeedTestInterval = time.Duration(nmins) * time.Minute
	}

	config.Testing = *arg_devtest

	// sanity checks
	if config.Database.Host != "" {
		err := false
		if config.Database.Name == "" {
			fmt.Println("Error: missing database name")
			err = true
		}
		if config.Database.User == "" {
			fmt.Println("Error: missing database username")
			err = true
		}
		if config.Database.Pass == "" {
			fmt.Println("Error: missing database password")
			err = true
		}
		if config.Traceroute == nil && config.RemoteAP != "" {
			fmt.Println("Error: select either traceroute OR remote")
			err = true
		}
		if err {
			os.Exit(-1)
		}
	} else {
		fmt.Println("Warning: missing database host: no data will be uploaded")
	}

	if *arg_mac != "" {
		config.MacAddr = *arg_mac
	} else {
		mac, err := getMacAddr()
		if err != nil {
			fmt.Println("Error: failed to get network device's MAC address: set on command line with: -mac=\"XX:XX:XX:XX:XX:XX\"")
			fmt.Println(err)
			os.Exit(-1)
		}
		config.MacAddr = mac
	}
	fmt.Println("Network MAC", config.MacAddr)
}

func getMacAddr() (string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	// try to get wifi MAC address first
	for _, ifa := range ifas {
		cleanName := strings.ToLower(strings.ReplaceAll(ifa.Name, "-", ""))

		if strings.Contains(cleanName, "wifi") ||
			strings.Contains(cleanName, "wlp") ||
			strings.Contains(cleanName, "wlan") {

			a := ifa.HardwareAddr.String()
			if a != "" {
				return a, nil
			}
		}
	}
	// fallback to ethernet MAC address
	for _, ifa := range ifas {
		cleanName := strings.ToLower(strings.ReplaceAll(ifa.Name, "-", ""))

		if strings.Contains(cleanName, "ethernet") ||
			strings.Contains(cleanName, "eth") ||
			strings.Contains(cleanName, "igb") ||
			strings.Contains(cleanName, "enp") {

			a := ifa.HardwareAddr.String()
			if a != "" {
				return a, nil
			}
		}
	}
	// fallback to first valid MAC address
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			return a, nil
		}
	}
	return "", fmt.Errorf("no valid MAC found")
}
