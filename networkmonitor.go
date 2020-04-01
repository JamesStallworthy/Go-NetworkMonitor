package main

import (
	"flag"
	"fmt"
	"net/url"
	"time"

	client "github.com/influxdata/influxdb1-client"
	"github.com/sparrc/go-ping"
)

func main() {
	urlPtr := flag.String("endpoint-url", "www.google.com", "URL to ping")
	influxURLPtr := flag.String("influxdb-url", "example.com:8086", "influxDBUrl")
	influxUserPtr := flag.String("influxdb-user", "username", "influx username")
	influxPasswordPtr := flag.String("influxdb-password", "password", "influx password")
	locationPtr := flag.String("location", "home", "location where this service is being run from")

	flag.Parse()

	host, err := url.Parse(fmt.Sprintf("https://%s", *influxURLPtr))
	if err != nil {
		panic(err)
	}

	conf := client.Config{
		URL:      *host,
		Username: *influxUserPtr,
		Password: *influxPasswordPtr,
	}

	con, err := client.NewClient(conf)
	forever(*con, *urlPtr, *locationPtr)
}

func forever(con client.Client, url string, location string) {
	ping, success := pingAddress(url)

	if success {
		writeToInflux(ping, con, location, url)
	}

	fmt.Println("Waiting 1 minute")
	nextTime := time.Now().Truncate(time.Minute)
	nextTime = nextTime.Add(time.Minute)
	time.Sleep(time.Until(nextTime))

	forever(con, url, location)
}

func pingAddress(address string) (time.Duration, bool) {
	pinger, err := ping.NewPinger(address)
	if err != nil {
		fmt.Println(err)
	} else {
		pinger.SetPrivileged(true)
		pinger.Timeout = time.Second * 10
		pinger.Count = 3
		pinger.Run()

		stats := pinger.Statistics()
		return stats.AvgRtt, true
	}
	return -1, false
}

func writeToInflux(pingLength time.Duration, con client.Client, location string, url string) {
	_, err := con.WriteLineProtocol(fmt.Sprintf("ping,location=%s,url=%s value=%d", location, url, pingLength), "network", "", "s", "one")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(fmt.Sprintf("Successful write to influxdb: %s", pingLength.String()))
	}
}
