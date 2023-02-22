package cli

import (
	"bytes"
	"crypto/sha512"
	"do/doge/version"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/kardianos/osext"
)

const (
	BinaryName       = "dhcli"
	ArtifactsBaseURL = "https://artifactory/dhcli/"

	ProductionBaseURL = ArtifactsBaseURL + "production/"
	StagingBaseURL    = ArtifactsBaseURL + "staging/"

	ProductionManifestURL = ProductionBaseURL + "packing_slip.json"
	StagingManifestURL    = StagingBaseURL + "packing_slip.json"
)

// userAgent returns the user agent string the client should use.
func userAgent() string {
	hostname, _ := os.Hostname()
	return "dhcli-" + hostname
}

// HTTPRequest wraps retryablehttp adding in our custom User-Agent, and any
// other defaults we want for HTTP connections.
func httpRequest(
	method string,
	url string,
	body []byte,
) (*http.Response, error) {
	req, err := retryablehttp.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(
		"User-Agent",
		fmt.Sprintf(BinaryName+"-cli %s", runtime.GOOS),
	)
	req.Header.Set("Content-Type", "application/json")

	c := retryablehttp.NewClient()
	c.RetryMax = 10
	c.RetryWaitMin = 1 * time.Second
	c.Logger = nil
	return c.Do(req)
}

func httpGET(url string) ([]byte, error) {

	response, err := httpRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to fetch %s; HTTP error %d", url, response.StatusCode,
		)
	}

	out, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func replaceCurrentVersion(binary []byte) error {
	oldBinDir, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}

	oldBin, err := osext.Executable()
	if err != nil {
		return err
	}

	oldBinStat, err := os.Stat(oldBin)
	if err != nil {
		return err
	}

	oldBinACL := oldBinStat.Mode()

	tempDir, err := ioutil.TempDir(oldBinDir, BinaryName+"-cli-update")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	newBin, err := os.Create(filepath.Join(tempDir, BinaryName))
	if err != nil {
		return err
	}

	_, err = newBin.Write(binary)
	if err != nil {
		return err
	}

	err = os.Chmod(newBin.Name(), oldBinACL)
	if err != nil {
		newBin.Close()
		return err
	}

	err = newBin.Close()
	if err != nil {
		return err
	}

	oldFilename := fmt.Sprintf("%s.old", oldBin)

	// If a <binary>.old file exists from a previous update, remove that first.
	if _, err := os.Stat(oldFilename); os.IsNotExist(err) {
		_ = os.Remove(oldFilename)
	}

	// Move the currently running binary aside.
	if err = os.Rename(oldBin, oldFilename); err != nil {
		return err
	}

	// Put the new binary in place of the old binary.
	if err = os.Rename(newBin.Name(), oldBin); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		// Note: Windows does not allow overwriting a running binary. So tell
		// the user it can be removed instead.
		fmt.Fprintf(
			os.Stdout,
			"\nUpdate complete. The previous version can now be removed:"+
				"\n    %s\n\n",
			oldFilename,
		)
	} else {
		err = os.Remove(oldFilename)
		if err != nil {
			return err
		} else {
			fmt.Fprintln(os.Stdout, "\nUpdate complete.")
		}
	}

	return nil
}

func Update(staging bool, dryRun bool) (err error) {
	var manifestBytes []byte
	var manifestURL string
	var baseURL string

	// Download the manifest file from the artifacts server.
	if staging {
		manifestURL = StagingManifestURL
		baseURL = StagingBaseURL
		fmt.Fprintf(
			os.Stdout, "Fetching staging manifest from %s ...\n", manifestURL,
		)
	} else {
		manifestURL = ProductionManifestURL
		baseURL = ProductionBaseURL
		fmt.Fprintf(
			os.Stdout, "Fetching production manifest from %s ...\n", manifestURL,
		)
	}

	manifestBytes, err = httpGET(manifestURL)
	if err != nil {
		return err
	}
	manifestBytes = bytes.TrimSpace(manifestBytes)

	// Convert the (hopefully JSON) response into a Manifest struct.
	var m packingslip.Manifest
	err = json.Unmarshal(manifestBytes, &m)
	if err != nil {
		return err
	}

	// First check if we actually need to update or not.
	fmt.Fprintf(
		os.Stdout,
		" Current version: %s\nManifest version: %s\n",
		version.GetCommit(),
		m.Version,
	)
	if version.GetCommit() == m.Version {
		fmt.Fprintln(os.Stdout, "No update required.")
		os.Exit(0)
	}

	// Determine which filename to look for in the Manifest.
	filename := BinaryName
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}
	path := fmt.Sprintf("%s/%s/%s", runtime.GOOS, runtime.GOARCH, filename)

	file, found := m.Files[path]
	if found {
		binaryURL := baseURL + path
		var binary []byte
		fmt.Fprintf(os.Stdout, "Found manifest entry for '%s':\n%v", path, &file)
		fmt.Fprintf(os.Stdout, "Fetching %s ...\n", binaryURL)
		binary, err = httpGET(binaryURL)
		if err != nil {
			return err
		}

		// Verify the downloaded binary.
		if len(binary) != file.Size {
			return fmt.Errorf(
				"file size mismatch; expected %d, got %d",
				file.Size,
				len(binary),
			)
		}
		sha512sum := sha512.Sum512(binary)
		sha512text := fmt.Sprintf("%x", sha512sum)
		if sha512text != file.SHA2_512 {
			return fmt.Errorf(
				"SHA512 mismatch; expected %s, got %s",
				file.SHA2_512,
				sha512text,
			)
		} else {
			fmt.Fprintln(os.Stdout, "New binary successfully verified.")
		}

		if !dryRun {
			err = replaceCurrentVersion(binary)
			if err != nil {
				return err
			}
		} else {
			fmt.Fprintf(os.Stdout, "Dry run; skipping binary replacement.\n")
		}
	} else {
		return fmt.Errorf("path not found in manifest: %s", path)
	}

	return nil
}
