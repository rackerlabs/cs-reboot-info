package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace"
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

	regions := Regions(provider, opts)

	fmt.Printf("Regions with a compute endpoint: %#v\n", regions)

	if useLocaltime {
		fmt.Println("Dummy switch so Go shuts up about unused variables!")
	}

	// Iterate through regions

	// Iterate through servers in each region

	// Pull the metadata key

	// Output a CSV row
	if outputCSV {
		fmt.Println("I would be writing a CSV file here!")
	}
}

// Regions acquires the service catalog and returns a slice of every region that contains a first-
// or next-gen compute endpoint.
func Regions(provider *gophercloud.ProviderClient, opts gophercloud.AuthOptions) []string {
	service := rackspace.NewIdentityV2(provider)

	result := tokens.Create(service, tokens.WrapOptions(opts))
	catalog, err := result.ExtractServiceCatalog()
	if err != nil {
		fmt.Printf("Unable to retrieve the service catalog: %v\n", err)
		os.Exit(1)
	}

	regions := make([]string, 5)
	for _, entry := range catalog.Entries {
		if entry.Name == "compute" {
			for _, endpoint := range entry.Endpoints {
				regions = append(regions, endpoint.Region)
			}
		}
	}
	return regions
}

// ShowTime renders a time as both UTC and guessed local TZ
func ShowTime(ts time.Time) string {
	return ""
}
