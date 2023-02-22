package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

var (
	environments = map[string]string{
		"Stage2":     "https://stork.staging",
		"Production": "https://stork.prod",
	}
)

func storkAuth(environment *url.URL) (*cookiejar.Jar, error) {
	storkEmail, ok := os.LookupEnv("STORK_USER")
	if !ok {
		fmt.Print("Unset environment variables! Please define:\n" +
			"'STORK_USER' and 'STORK_PASS' in your shell\n" +
			"with values from Vault: stork-dhcp/tools\n")
		os.Exit(127)
	}
	storkPassword, ok := os.LookupEnv("STORK_PASS")
	if !ok {
		fmt.Print("Unset environment variables! Please define:\n" +
			"'STORK_USER' and 'STORK_PASS' in your shell\n" +
			"with values from Vault: stork-dhcp/tools\n")
		os.Exit(127)
	}

	postBody, _ := json.Marshal(map[string]string{
		"useremail":    storkEmail,
		"userpassword": storkPassword,
	})

	resp, err := http.Post(environment.String()+"/api/sessions", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		fmt.Printf("\nan error occured authenticating with: %s", err.Error())
	}

	if resp.StatusCode == 200 {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, errors.New("could not create auth cookie storage")
		}
		jar.SetCookies(environment, resp.Cookies())
		return jar, nil
	} else {
		return nil, fmt.Errorf("stork returned %d on auth", resp.StatusCode)
	}

}
