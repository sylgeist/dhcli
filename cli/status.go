package cli

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type StatusCmd struct{}

type KeaStatus struct {
	DhcpDaemons []struct {
		Active     bool   `json:"active"`
		AppName    string `json:"appName"`
		AppVersion string `json:"appVersion"`
		Machine    string `json:"machine"`
		Uptime     int    `json:"uptime"`
	} `json:"dhcpDaemons"`
}

func (s *StatusCmd) Run() error {
	// Because we are hitting both Prod/Staging we don't want to error out if one of them is unavailable
	for envName, environment := range environments {
		envURL, err := url.Parse(environment)
		if err != nil {
			fmt.Printf("Error parsing environment URLs: %v", err.Error())
			return err
		}

		jar, err := storkAuth(envURL)
		if err != nil {
			fmt.Printf("Error while creating cookie jar %s", err.Error())
			return err
		}

		client := &http.Client{Jar: jar}

		// http://netboot-stork-01.nyc3.internal.digitalocean.com/api/docs#operation/getDhcpOverview
		envURL.Path = "/api/overview"

		resp, err := client.Get(envURL.String())
		if err != nil {
			fmt.Printf("Got error %s", err.Error())
		}

		if resp.StatusCode != 200 {
			fmt.Printf("Error: Stork returned %d on search!", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Manually calling close due to for loop
		// We aren't checking errors on this, because they don't really impact the outcome
		_ = resp.Body.Close()

		if len(body) == 0 {
			fmt.Printf("Unexpected output from Stork server: %s", err.Error())
		}

		k := KeaStatus{}
		jsonErr := json.Unmarshal(body, &k)
		if jsonErr != nil {
			fmt.Println(jsonErr)

		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetBorder(false)
		table.SetHeader([]string{"Active", "Kea Instance", "Version", "Host", "Uptime"})
		for daemon := range k.DhcpDaemons {
			// Formatting updates for visual clarity
			uptime, _ := time.ParseDuration(strconv.Itoa(k.DhcpDaemons[daemon].Uptime) + "s")
			keahost := strings.Replace(k.DhcpDaemons[daemon].Machine, ".internal.digitalocean.com", "", 1)

			// Build the table structure for each daemon entry
			data := []string{
				strconv.FormatBool(k.DhcpDaemons[daemon].Active),
				k.DhcpDaemons[daemon].AppName,
				k.DhcpDaemons[daemon].AppVersion,
				keahost,
				uptime.String(),
			}
			table.Append(data)
		}
		fmt.Printf("\n%s:\n", envName)
		table.Render()

	}
	return nil
}
