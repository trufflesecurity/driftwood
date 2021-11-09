package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/trufflesecurity/driftwood/pkg/exp/client"
	"github.com/trufflesecurity/driftwood/pkg/exp/parser"
)

var version = "dev"

func main() {
	app := &cli.App{
		Name:      "driftwood",
		Usage:     "Verify if a private key is used for important things.",
		UsageText: "driftwood <path to key or '-' for stdin>",
		Version:   version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "pretty-json",
			},
			&cli.BoolFlag{
				Name: "debug",
			},
			&cli.BoolFlag{
				Name: "public-key",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()

			logger := logrus.New()
			logger.SetOutput(os.Stderr)
			if c.Bool("json") {
				logger.SetFormatter(&logrus.JSONFormatter{
					DisableTimestamp: true,
				})
			} else {
				logger.SetFormatter(&logrus.TextFormatter{
					DisableTimestamp: true,
				})
			}

			if c.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			}

			if len(args) == 0 {
				cli.ShowAppHelpAndExit(c, 1)
				return nil
			}

			file := args[0]

			var inKey []byte
			var err error
			if file == "-" {
				buff := bytes.NewBuffer(nil)
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					buff.WriteString(scanner.Text() + "\n")
				}
				inKey = buff.Bytes()
			} else {
				inKey, err = ioutil.ReadFile(args[0])
				if err != nil {
					logger.Fatalf("File cannot be read: %s", err)
					return nil
				}
			}

			var publicKey []byte
			if ! c.Bool("public-key") {
				publicKey, err = parser.PublicKey(inKey)
				if err != nil {
					logger.Fatalf("Error computing public key: %s", err)
				}
			} else {
				publicKey = inKey
			}

			result, err := client.Lookup(version, publicKey)
			if err != nil {
				logger.Fatalf("Error looking up public key: %s", err)
			}

			if !c.Bool("pretty-json") {
				out, err := json.Marshal(result)
				if err != nil {
					logger.Fatalf("Error marshalling result: %s", err)
				}
				fmt.Println(string(out))
			} else {
				out, err := json.MarshalIndent(result, "", "\t")
				if err != nil {
					logger.Fatalf("Error marshalling result: %s", err)
				}
				fmt.Println(string(out))
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
