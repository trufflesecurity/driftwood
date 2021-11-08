package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

var (
	client = &http.Client{
		Timeout:   time.Second * 3,
		Transport: http.DefaultTransport,
	}
)

func Lookup(version string, publicKey []byte) (result Result, err error) {
	req, err := http.NewRequest("POST", "https://keychecker.trufflesecurity.com/publickey", nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", fmt.Sprintf("Driftwood %s", version))
	var body io.ReadCloser
	for {
		req.Body = io.NopCloser(bytes.NewBuffer(publicKey))
		res, err := client.Do(req)
		if err != nil {
			return result, err
		}
		if logrus.GetLevel() == logrus.DebugLevel {
			o, _ := httputil.DumpResponse(res, true)
			fmt.Println(string(o))
		}
		if res.StatusCode == http.StatusTooManyRequests || res.StatusCode == http.StatusServiceUnavailable {
			if s, ok := res.Header["Retry-After"]; ok {
				var retryTime time.Time
				if retryTime, err = time.Parse(time.RFC1123, s[0]); err == nil {
					wait := time.Until(retryTime.Add(time.Second * 3))
					logrus.Infof("Hit rate limit, retrying in %d seconds", int(wait.Seconds()))
					time.Sleep(wait)
					continue
				}
			}
		}
		if res.StatusCode == http.StatusOK {
			body = res.Body
			defer res.Body.Close()
			break
		}
	}

	rawMap := map[string]interface{}{}
	err = json.NewDecoder(body).Decode(&rawMap)
	if err != nil {
		return
	}
	var md mapstructure.Metadata
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			Metadata: &md,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				ToTimeHookFunc()),
			Result: &result,
		})
	if err != nil {
		return
	}
	if err = decoder.Decode(rawMap); err != nil {
		return
	}
	result.X = map[string]interface{}{}
	for _, k := range md.Unused {
		result.X[k] = rawMap[k]
	}

	for i := range result.CertificateResults {
		result.CertificateResults[i].VerificationURL = fmt.Sprintf("https://crt.sh/?q=%s", result.CertificateResults[i].CertificateFingerprint)
	}

	return
}

func ToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

type CertificateResult struct {
	CertificateFingerprint string
	ExpirationTimestamp    time.Time
	VerificationURL        string
}

type GitHubSSHResult struct {
	Username string
}

type Result struct {
	CertificateResults []CertificateResult    `json:",omitempty"`
	GitHubSSHResults   []GitHubSSHResult      `json:",omitempty"`
	Error              string                 `json:",omitempty"`
	X                  map[string]interface{} `json:",omitempty"`
}
