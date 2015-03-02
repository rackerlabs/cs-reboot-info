package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rackspace/gophercloud"
	osV2Servers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace"
	rsV2Servers "github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
	"github.com/rackspace/gophercloud/rackspace/identity/v2/tokens"
	rsV1Servers "github.com/smashwilson/gophercomputev1"
)

const (
	metadataKey     = "rax:reboot_window"
	metadataTimeFmt = "2006-01-02T15:04:05Z"
	appVer          = "1.1"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "cs-reboot-info v%s\n", appVer)
		fmt.Fprintf(os.Stderr, "Usage: %s [--csv] username apikey\n\n", os.Args[0])

		flag.CommandLine.PrintDefaults()
	}

	outputToCSV := flag.Bool("csv", false,
		"Output a CSV file to 'cs-reboot-info-output.csv' in the current directory.")
	outputVersion := flag.Bool("version", false,
		"Report the version of this binary.")
	flag.Parse()

	if *outputVersion {
		fmt.Printf("cs-reboot-info version %s\n", appVer)
		os.Exit(0)
	}

	if flag.NArg() != 2 {
		fmt.Println("You must supply a username and API key as the last two arguments.")
		flag.Usage()
		os.Exit(1)
	}

	username, apiKey := flag.Arg(0), flag.Arg(1)

	opts := gophercloud.AuthOptions{
		Username: username,
		APIKey:   apiKey,
	}

	provider, err := rackspace.AuthenticatedClient(opts)
	if err != nil {
		fmt.Printf("Unable to authenticate: %v\n", err)
		os.Exit(1)
	}

	provider.UserAgent.Prepend(fmt.Sprintf("cs-reboot-info/%s", appVer))

	regions, fg := Regions(provider, opts)
	fmt.Printf("cs-reboot-info version %s\n\n", appVer)
	fmt.Printf("Regions with a Cloud Servers endpoint: %s\n", strings.Join(regions, ", "))
	if fg {
		fmt.Println("Found both First and Next Generation endpoints.")
		fmt.Println()
	}

	fgRegion := "DFW"
	if len(regions) == 1 && regions[0] == "LON" {
		fgRegion = "LON"
	}

	var entries []entry

	// Iterate through regions with an NG compute endpoint. Collect data about each server.
	for _, region := range regions {
		compute, err := rackspace.NewComputeV2(provider, gophercloud.EndpointOpts{
			Region: region,
		})
		if err != nil {
			fmt.Printf("Unable to locate a Next Gen Cloud Servers endpoint in region %s: %v\n", region, err)
			continue
		}

		err = rsV2Servers.List(compute, nil).EachPage(func(page pagination.Page) (bool, error) {
			s, err := osV2Servers.ExtractServers(page)
			if err != nil {
				return false, err
			}

			for _, server := range s {
				entry, err := ConstructEntry(server, "Next Gen", region)
				if err != nil {
					if strings.Contains(err.Error(), "not present") {
						continue
					}
					fmt.Printf("%s\n", err)
					continue
				} else {
					entries = append(entries, *entry)
				}
			}

			return true, nil
		})
	}

	// Iterate through regions with an FG compute endpoint. Collect data about each server.
	compute, err := rsV1Servers.NewComputeV1(provider, gophercloud.EndpointOpts{
		Availability: gophercloud.AvailabilityPublic,
	})
	if err != nil {
		fmt.Printf("Unable to locate a First Gen Cloud Servers endpoint. Skipping...\n")
	} else {
		err = rsV1Servers.List(compute, rsV1Servers.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
			s, err := osV2Servers.ExtractServers(page)
			if err != nil {
				return false, err
			}

			for _, server := range s {
				entry, err := ConstructEntry(server, "First Gen", fgRegion)
				if err != nil {
					if strings.Contains(err.Error(), "not present") {
						continue
					}
					fmt.Printf("%s\n", err)
					continue
				} else {
					entries = append(entries, *entry)
				}
			}

			return true, nil
		})
		if err != nil {
			fmt.Printf("Error listing servers: %+v\n", err)
		}
	}

	if len(entries) > 0 {
		fmt.Printf("There are %d Cloud Servers with an automated reboot scheduled.", len(entries))
		outputTabular(entries)
		if *outputToCSV {
			outputCSV(entries)
		}
	} else {
		fmt.Printf("cs-reboot-info version %s\n\n", appVer)
		fmt.Printf("You have no Cloud Servers with an automated reboot scheduled.\n")
		os.Exit(0)
}

// Regions acquires the service catalog and returns a slice of every region that contains a next-gen
// server endpoint, and a boolean indicating whether or not this customer has access to FG servers.
func Regions(provider *gophercloud.ProviderClient, opts gophercloud.AuthOptions) ([]string, bool) {
	service := rackspace.NewIdentityV2(provider)

	result := tokens.Create(service, tokens.WrapOptions(opts))
	catalog, err := result.ExtractServiceCatalog()
	if err != nil {
		fmt.Printf("Unable to retrieve the service catalog: %v\n", err)
		os.Exit(1)
	}

	var regions []string
	var fg bool
	for _, entry := range catalog.Entries {
		if entry.Type == "compute" {
			for _, endpoint := range entry.Endpoints {
				if endpoint.Region == "" {
					fg = true
				} else {
					regions = append(regions, endpoint.Region)
				}
			}
		}
	}
	return regions, fg
}

// ConstructEntry extracts the metadata key and builds an entry for a server.
func ConstructEntry(server osV2Servers.Server, genType, region string) (*entry, error) {
	window, ok := server.Metadata[metadataKey]
	if !ok {
		return nil, fmt.Errorf("Metadatum %s was not present in the result for server %s", metadataKey, server.ID)
	}

	windowString, ok := window.(string)
	if !ok {
		return nil, fmt.Errorf("Metadatum %s for server %s was not a string: %#v", metadataKey, server.ID, window)
	}

	// Expected format: 2014-01-28T00:00:00Z;2014-01-28T03:00:00Z

	parts := strings.Split(windowString, ";")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Unexpected metadatum format for server %s: %s", server.ID, windowString)
	}

	start, err := time.Parse(metadataTimeFmt, parts[0])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse window start time for server %s: %s", server.ID, parts[0])
	}

	end, err := time.Parse(metadataTimeFmt, parts[1])
	if err != nil {
		return nil, fmt.Errorf("Unable to parse window end time for server %s: %s", server.ID, parts[1])
	}

	e := &entry{
		Server:      server,
		Region:      region,
		GenType:     genType,
		WindowStart: start,
		WindowEnd:   end,
	}
	return e, nil
}
