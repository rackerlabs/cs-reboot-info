package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
	"github.com/rackspace/gophercloud/rackspace/identity/v2/tokens"
)

func main() {
	outputCSV := *flag.Bool("csv", false, "Output a CSV file")
	useLocaltime := *flag.Bool("localtime", false, "Use local timestamps instead of UTC")

	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("You must supply a username and API key after all other args.")
		os.Exit(1)
	}

	username, apiKey := flag.Arg(0), flag.Arg(1)

	opts := gophercloud.AuthOptions{
		Username: username,
		APIKey:   apiKey,
	}

	provider, err := rackspace.AuthenticatedClient(opts)
	if err != nil {
		fmt.Printf("Unable to authenticate: %v", err)
	}

	regions, fg := Regions(provider, opts)

	fmt.Printf("Regions with a compute endpoint: %#v\n", regions)
	if fg {
		fmt.Println("You do have a first-gen endpoint, too.")
	}

	if useLocaltime {
		fmt.Println("Dummy switch so Go shuts up about unused variables!")
	}

	var entries []entry

	// Iterate through regions with an NG compute endpoint. Collect data about each server.
	for _, region := range regions {
		compute, err := rackspace.NewComputeV2(provider, gophercloud.EndpointOpts{
			Region: region,
		})
		if err != nil {
			fmt.Printf("Unable to locate a compute endpoint in region %s: %v\n", region, err)
			continue
		}

		err = servers.List(compute, nil).EachPage(func(page pagination.Page) (bool, error) {
			s, err := servers.ExtractServers(page)
			if err != nil {
				return false, err
			}

			for _, server := range s {
				entry := entry{
					Server:      server,
					Region:      region,
					GenType:     "v2",
					WindowStart: time.Now(),
					WindowEnd:   time.Now(),
				}
				entries = append(entries, entry)
			}

			return true, nil
		})
	}

	// Pull the metadata key

	// Output a CSV row
	if outputCSV {
		fmt.Println("I would be writing a CSV file here!")
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

// ShowTime renders a time as both UTC and guessed local TZ
func ShowTime(ts time.Time) string {
	return ""
}
