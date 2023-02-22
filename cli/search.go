package cli

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type SearchCmd struct {
	LeaseSearch string `kong:"arg='',name='MAC or IP address',help='e.g. 78:12:b6:d9:ce:58 or 10.30.2.4'"`
}

type Lease struct {
	Total int `json:"total"`
	Items []struct {
		AppName   string `json:"appName"`
		Hostname  string `json:"hostname"`
		HwAddress string `json:"hwAddress"`
		IpAddress string `json:"ipAddress"`
		SubnetID  int    `json:"subnetId"`
		State     int    `json:"state"`
	} `json:"items"`
}

func (s *SearchCmd) Run() error {
	searchTerm := s.LeaseSearch

	// Basic input validation
	// If searchTerm appears to be a valid IP or MAC address we continue
	if inputTest := net.ParseIP(searchTerm); inputTest == nil {
		_, err := net.ParseMAC(searchTerm)
		if err != nil {
			return err
		}
	}

	for envName, environment := range environments {
		envURL, err := url.Parse(environment)
		if err != nil {
			log.Fatalf("Error parsing environment URLs: %v", err.Error())
		}

		jar, err := storkAuth(envURL)
		if err != nil {
			fmt.Print(err.Error())
		}

		client := &http.Client{Jar: jar}

		envURL.Path = "/api/leases"
		queryValues := envURL.Query()
		queryValues.Set("text", searchTerm)
		envURL.RawQuery = queryValues.Encode()

		resp, err := client.Get(envURL.String())
		if err != nil {
			log.Print(err.Error())
		}

		if resp.StatusCode != 200 {
			log.Printf("Error: Stork returned %d on search!", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		// Manually calling close due to for loop
		// We aren't checking errors on this, because they don't really impact the outcome
		_ = resp.Body.Close()

		if len(body) == 0 {
			log.Printf("Unexpected output from Stork server: %s", err.Error())
		}

		l := Lease{}

		jsonErr := json.Unmarshal(body, &l)
		if jsonErr != nil {
			return jsonErr
		}

		switch {
		case l.Total == 0:
			fmt.Printf("%s:\nNo results found for: %s\n\n", envName, searchTerm)
		case l.Total == 1:
			var state string
			if l.Items[0].State == 0 {
				state = "Active"
			} else {
				state = "Inactive - check logs"
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoWrapText(false)
			table.SetBorder(false)
			table.SetHeader([]string{"Kea Instance",
				"Hostname",
				"Hardware Address (MAC)",
				"IP Address",
				"Subnet ID",
				"Lease State"})
			for lease := range l.Items {
				// Build the table structure for each daemon entry
				data := []string{
					l.Items[lease].AppName,
					l.Items[lease].Hostname,
					l.Items[lease].HwAddress,
					l.Items[lease].IpAddress,
					strconv.Itoa(l.Items[lease].SubnetID),
					state,
				}
				table.Append(data)
			}
			fmt.Printf("\n%s:\n", envName)
			table.Render()
			fmt.Print("\n")
		case l.Total >= 2:
			fmt.Printf("Ambiguous results returned for: %s\n"+
				"Multiple lease records returned! Check the Kea instance for more details:\n"+
				"Environment: %s\n"+
				"Lease count: %d\n",
				searchTerm, envName, l.Total)
		default:
			log.Printf("%s: unknown error occured searching for: %s", envName, searchTerm)
		}
		// Reset the switch var for the next run of the loop as there's a chance we won't have a valid value
		l.Total = 0
	}
	return nil
}
