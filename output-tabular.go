package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type entry struct {
	Server      servers.Server
	Region      string
	GenType     string
	WindowStart time.Time
	WindowEnd   time.Time
}

var testServer = servers.Server{
	ID:   "8c65cb68-0681-4c30-bc88-6b83a8a26aee",
	Name: "Gophercloud-pxpGGuey",
}

func parseTime(t time.Time) string {
	return t.Format("02 Jan 15:04")
}

func elide(value string) string {
	if len(value) > 20 {
		return value[0:17] + "..."
	}
	return value
}

func hashes(num int) string {
	return strings.Repeat("-", num)
}

func outputTabular(entries []entry) {
	fmt.Println("")
	fmt.Printf("| %-15s | %-36s | %-20s | %-27s | %-27s |\n", "Type", "Server ID", "Server Name", "Reboot Window (UTC)", "Reboot Window (Local)")
	fmt.Printf("| %-15s | %-36s | %-20s | %-27s | %-27s |\n", hashes(15), hashes(36), hashes(20), hashes(27), hashes(27))
	for _, s := range entries {
		region := "(" + s.Region + ")"
		if s.GenType == "First Gen" {
			region = ""
		}
		fmt.Printf("| %-9s %-5s | %-36s | %-20s | %-12s - %-12s | %-12s - %-12s |\n", s.GenType, region,
			s.Server.ID, elide(s.Server.Name), parseTime(s.WindowStart), parseTime(s.WindowEnd),
			parseTime(s.WindowStart.Local()), parseTime(s.WindowEnd.Local()))
	}
}
