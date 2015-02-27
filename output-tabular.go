package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

func main() {
	t, _ := time.Parse("2014-01-28T00:00:00Z", "2014-01-28T00:00:00Z")
	entries := []entry{
		entry{Server: testServer, Region: "IAD", WindowStart: t, WindowEnd: t, GenType: "First Gen"},
		entry{Server: testServer, Region: "IAD", WindowStart: t, WindowEnd: t, GenType: "First Gen"},
	}
	outputTabular(entries)
}

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
	return t.Format("Mon 02 Jan 15:04")
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
	fmt.Printf("| %-10s | %-6s | %-36s | %-20s | %-43s |\n", hashes(10), hashes(6), hashes(36), hashes(20), hashes(43))
	fmt.Printf("| %-10s | %-6s | %-36s | %-20s | %-43s |\n", "Generation", "Region", "Server ID", "Server Name", "Reboot Window")
	fmt.Printf("| %-10s | %-6s | %-36s | %-20s | %-43s |\n", hashes(10), hashes(6), hashes(36), hashes(20), hashes(43))
	for _, s := range entries {
		fmt.Printf("| %-10s | %-6s | %-36s | %-20s | %-16s - %-16s %-7s |\n",
			s.GenType, s.Region, s.Server.ID, elide(s.Server.Name), parseTime(s.WindowStart),
			parseTime(s.WindowEnd), "(UTC)")

		fmt.Printf("| %-10s | %-6s | %-36s | %-20s | %-16s - %-16s %-7s |\n", "",
			"", "", "", parseTime(s.WindowStart.Local()), parseTime(s.WindowEnd.Local()), "(Local)")
		fmt.Printf("| %-10s | %-6s | %-36s | %-20s | %-43s |\n", hashes(10), hashes(6), hashes(36), hashes(20), hashes(43))
	}
}
