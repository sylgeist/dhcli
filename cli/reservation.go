package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

type ResCmd struct {
	ResTerm string `kong:"arg='',name='IP address/subnet or Region',help='e.g. 10.4.2.5/27, NYC3'"`
}

type KeaRes struct {
	Total int `json:"total"`
	Items []struct {
		AddressReservations []struct {
			Address string `json:"address"`
		} `json:"addressReservations"`
		HostIdentifiers []struct {
			IdHexValue string `json:"idHexValue"`
		} `json:"hostIdentifiers"`
		LocalHosts []struct {
			AppName string `json:"appName"`
		} `json:"localHosts"`
	} `json:"items"`
}

func (r *ResCmd) Run() error {

	// Basic input validation
	// If searchTerm appears to be a valid IP or region
	ipaddr, _ := regexp.Compile("(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}")
	region, _ := regexp.Compile("(\\w{3}\\d+)")
	searchTerm := r.ResTerm

	if !(ipaddr.MatchString(searchTerm) || region.MatchString(searchTerm)) {
		return errors.New("Invalid input!")
	}

	var envName = "Production"
	var envURL = environments[envName]

	if region.MatchString(searchTerm) {
		if isStage2(searchTerm) {
			envName = "Stage2"
			envURL = environments[envName]
		}
	}

	storkURL, err := url.Parse(envURL)
	if err != nil {
		fmt.Printf("Error parsing environment URLs: %v", err.Error())
		return err
	}

	jar, err := storkAuth(storkURL)
	if err != nil {
		return err
	}

	client := &http.Client{Jar: jar}

	storkURL.Path = "/api/hosts"
	queryValues := storkURL.Query()
	queryValues.Del("text")
	queryValues.Set("limit", "100")

	switch {
	case ipaddr.MatchString(searchTerm):
		subnetURL, _ := url.Parse(envURL)
		subnetID, err := getSubnetID(searchTerm, subnetURL, jar)
		if err != nil {
			fmt.Printf("%s: %s: %s", envName, searchTerm, err.Error())
			return err
		}
		queryValues.Set("subnetId", strconv.Itoa(subnetID))
	case region.MatchString(searchTerm):
		appURL, _ := url.Parse(envURL)
		appID, err := getAppID(searchTerm, appURL, jar)
		if err != nil {
			fmt.Printf("%s: %s: %s", envName, searchTerm, err.Error())
			return err
		}
		queryValues.Set("appId", strconv.Itoa(appID))
	}
	storkURL.RawQuery = queryValues.Encode()

	resp, err := client.Get(storkURL.String())
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Error: Stork returned %d on search!", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if len(body) == 0 {
		fmt.Printf("%s: unexpected output from Stork server", envName)
	}

	k := KeaRes{}
	jsonErr := json.Unmarshal(body, &k)
	if jsonErr != nil {
		return jsonErr
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetBorder(false)
	table.SetHeader([]string{"Kea Instance", "Hardware Address (MAC)", "IP Address"})
	for reservation := range k.Items {
		// Build the table structure for each daemon entry
		data := []string{
			k.Items[reservation].LocalHosts[0].AppName,
			k.Items[reservation].HostIdentifiers[0].IdHexValue,
			k.Items[reservation].AddressReservations[0].Address,
		}
		table.Append(data)
	}
	fmt.Printf("\n%s: (%d reserved addresses)\n", envName, k.Total)
	table.Render()
	fmt.Println()
	return nil
}
