package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type LogsCmd struct {
	LogsInstance string `kong:"arg='',name='kea-instance',help='e.g. NYC3, S2R8'"`
}

type KeaLogEntry struct {
	Name     string   `json:"appName"`
	Contents []string `json:"contents"`
}

func (l *LogsCmd) Run() error {
	searchInstance := l.LogsInstance

	var envName = "Production"
	var envURL = environments[envName]

	if isStage2(searchInstance) {
		envName = "Stage2"
		envURL = environments[envName]
	}

	storkURL, err := url.Parse(envURL)
	if err != nil {
		log.Printf("Error parsing environment URLs: %v", err.Error())
		return err
	}

	jar, err := storkAuth(storkURL)
	if err != nil {
		return err
	}

	logURL, _ := url.Parse(envURL)
	logId, err := getLogID("kea-dhcp4", searchInstance, logURL, jar)
	if err != nil {
		fmt.Printf("%s: %s: %s\n", envName, searchInstance, err.Error())
		return err
	}

	logURL.Path = fmt.Sprintf("/api/logs/%d", logId)

	client := &http.Client{Jar: jar}
	resp, err := client.Get(logURL.String())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Error: Stork returned %d on search!", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()

	if len(body) == 0 {
		fmt.Printf("Unexpected output from Stork server: %s", err.Error())
	}

	k := KeaLogEntry{}
	jsonErr := json.Unmarshal(body, &k)
	if jsonErr != nil {
		return jsonErr
	}

	fmt.Printf("%s: Recent Log Entries\n", envName)
	for _, value := range k.Contents {
		fmt.Printf("%s\n", value)
	}
	return nil
}
