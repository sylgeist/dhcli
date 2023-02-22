package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type KeaSubnet struct {
	Total int `json:"total"`
	Items []struct {
		SubnetID int `json:"id"`
	} `json:"items"`
}

type KeaApp struct {
	Total int `json:"total"`
	Items []struct {
		AppName string `json:"name"`
		AppID   int    `json:"id"`
	} `json:"items"`
}

type KeaLogs struct {
	Name    string `json:"name"`
	Details struct {
		Daemons []struct {
			LogTargets []struct {
				LogId   int    `json:"id"`
				LogName string `json:"name"`
			} `json:"logTargets"`
		} `json:"daemons"`
	} `json:"details"`
}

func getSubnetID(subnet string, envURL *url.URL, jar *cookiejar.Jar) (subnetid int, err error) {
	s := KeaSubnet{}
	client := &http.Client{Jar: jar}

	envURL.Path = "/api/subnets"
	queryValues := envURL.Query()
	queryValues.Set("text", subnet)
	envURL.RawQuery = queryValues.Encode()

	resp, err := client.Get(envURL.String())
	if err != nil {
		return 0, errors.New("error looking up subnet ID")
	}

	if resp.StatusCode != 200 {
		return 0, errors.New("error looking up subnet ID")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.New("error looking up subnet ID")
	}
	defer resp.Body.Close()

	if len(body) == 0 {
		return 0, errors.New("error looking up subnet ID")
	}

	jsonErr := json.Unmarshal(body, &s)
	if jsonErr != nil {
		return 0, errors.New("error looking up subnet ID")
	}

	if s.Total == 1 {
		return s.Items[0].SubnetID, nil
	} else {
		return 0, errors.New("no results found for subnet")
	}
}

func getAppID(appname string, envURL *url.URL, jar *cookiejar.Jar) (appid int, err error) {
	a := KeaApp{}
	client := &http.Client{Jar: jar}

	envURL.Path = "/api/apps"
	queryValues := envURL.Query()
	queryValues.Set("limit", "25")
	envURL.RawQuery = queryValues.Encode()

	resp, err := client.Get(envURL.String())
	if err != nil {
		return 0, errors.New("error looking up app ID")
	}

	if resp.StatusCode != 200 {
		return 0, errors.New("error looking up app ID")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.New("error looking up app ID")
	}
	defer resp.Body.Close()

	if len(body) == 0 {
		return 0, errors.New("error looking up app ID")
	}

	jsonErr := json.Unmarshal(body, &a)
	if jsonErr != nil {
		return 0, errors.New("error looking up app ID")
	}

	if a.Total >= 1 {
		for _, instanceData := range a.Items {
			if appname == instanceData.AppName {
				return instanceData.AppID, nil
			}
		}
	}
	return 0, errors.New("no results found for app instance")
}

func getLogID(logname string, appname string, envURL *url.URL, jar *cookiejar.Jar) (logid int, err error) {
	l := KeaLogs{}
	client := &http.Client{Jar: jar}

	// Since everything is ID based, we need to get that value first
	appId, err := getAppID(appname, envURL, jar)
	if err != nil {
		return 0, errors.New("error looking up log id")
	}

	envURL.Path = fmt.Sprintf("/api/apps/%d", appId)

	resp, err := client.Get(envURL.String())
	if err != nil {
		return 0, errors.New("error looking up app ID")
	}

	if resp.StatusCode != 200 {
		return 0, errors.New("error looking up app details")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.New("error parsing app details")
	}
	defer resp.Body.Close()

	if len(body) == 0 {
		return 0, errors.New("unexpected empty response")
	}

	jsonErr := json.Unmarshal(body, &l)
	if jsonErr != nil {
		return 0, errors.New("error parsing response")
	}

	for _, daemon := range l.Details.Daemons {
		if daemon.LogTargets[0].LogName == logname {
			return daemon.LogTargets[0].LogId, nil
		}
	}
	return 0, errors.New("no results found for log")
}

func isStage2(region string) bool {
	if strings.HasPrefix(region, "S2") {
		return true
	}
	return false
}
