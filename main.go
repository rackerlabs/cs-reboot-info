package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jrperritt/gophercloud/rackspace"
	rsV1Servers "github.com/jrperritt/gophercloud/rackspace/compute/v1/servers"
	"github.com/rackspace/gophercloud"
	osV2Servers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/pagination"
	rsV2Servers "github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
	"github.com/rackspace/gophercloud/rackspace/identity/v2/tokens"
)

const (
	metadataKey     = "rax:reboot_window"
	metadataTimeFmt = "2006-01-02T15:04:05Z"
)

func main() {
	outputToCSV := *flag.Bool("csv", false, "Output a CSV file with the results.")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("You must supply a username and API key as the last two arguments.")
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

	regions, fg := Regions(provider, opts)

	fmt.Printf("Regions with a Cloud Servers endpoint: %s\n", strings.Join(regions, ", "))
	if fg {
		fmt.Println("Found both First and Next Generation endpoints.")
	}

	var numUnaffected int
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
						numUnaffected++
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
	compute, err := rackspace.NewComputeV1(provider, gophercloud.EndpointOpts{
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
				entry, err := ConstructEntry(server, "First Gen", "DFW")
				if err != nil {
					if strings.Contains(err.Error(), "not present") {
						numUnaffected++
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

	if numUnaffected > 0 {
		fmt.Printf("The following %d Cloud Servers have an automated reboot scheduled:", len(entries))
	}

	if outputToCSV {
		outputCSV(entries)
	} else {
		outputTabular(entries)
	}
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
