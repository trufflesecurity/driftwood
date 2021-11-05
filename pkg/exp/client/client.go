package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

var (
	client *http.Client
)

func init() {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 3
	rc.HTTPClient.Timeout = time.Second * 3
	rc.Logger = nil
	client = rc.StandardClient()
}

func Lookup(fingerprint string, includeExpired bool) (result *DriftwoodResult, err error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("https://keychecker.trufflesecurity.com/%s", fingerprint), nil)
	if err != nil {
		return
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println(res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return
	}

	return
}

type DriftwoodResult struct {
	CertificateResults []struct {
		CertificateFingerprint string    `json:"CertificateFingerprint"`
		ExpirationTimestamp    time.Time `json:"ExpirationTimestamp"`
	} `json:"CertificateResults"`
	GitHubSSHResults []struct {
		Username string `json:"Username"`
	} `json:"GitHubSSHResults"`
}
