package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/trufflesecurity/driftwood/pkg/exp/client"
	"github.com/trufflesecurity/driftwood/pkg/exp/fingerprint"
)

func main() {
	app := &cli.App{
		Name:      "driftwood",
		Usage:     "See if a private key is associated with a TLS certificate or GitHub user's SSH key.",
		UsageText: "driftwood `private key file`",
		Version:   "v1.0.0",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "json",
			},
			&cli.BoolFlag{
				Name: "include-expired",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()

			logger := logrus.New()
			if c.Bool("json") {
				logger.SetFormatter(&logrus.JSONFormatter{
					DisableTimestamp: true,
				})
			} else {
				logger.SetFormatter(&logrus.TextFormatter{
					DisableTimestamp: true,
				})
			}

			if len(args) == 0 {
				cli.ShowAppHelpAndExit(c, 1)
				return nil
			}

			priv, err := ioutil.ReadFile(args[0])
			if err != nil {
				logger.Fatalf("File cannot be read: %s", err)
				return nil
			}

			fprint, err := fingerprint.PEMKey(priv)
			if err != nil {
				logger.Fatalf("Error computing public key fingerprint: %s", err)
			}
			result, err := client.Lookup(fprint, c.Bool("include-expired"))
			if err != nil {
				logger.Fatalf("Error looking up public key fingerprint: %s", err)
			}

			foundSomething := false
			for _, res := range result.CertificateResults {
				logger.WithFields(logrus.Fields{
					"verification_link": fmt.Sprintf("https://crt.sh/?q=%s", res.CertificateFingerprint),
					"expires":           res.ExpirationTimestamp,
				}).Print("TLS certificate result")
				foundSomething = true
			}

			for _, res := range result.GitHubSSHResults {
				logger.WithFields(logrus.Fields{
					"username": res.Username,
				}).Print("GitHub user SSH key result")
				foundSomething = true
			}

			if !foundSomething {
				logger.Info("Didn't find any results")
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
