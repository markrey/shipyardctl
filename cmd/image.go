// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"net/http"
	"io"
	"mime/multipart"
	"bytes"
	"os"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image <appName> <revision> <zipPath>",
	Short: "builds a Docker image with Shipyard",
	Long: `This command is used to build Docker images with Shipyard's build API
from a given, zipped Node.js application.

Within the project zip, there must be a valid package.json.

Example of use:

$ apigeectl build image example 1 "./path/to/zipped/app"`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			fmt.Println("Missing required args\n")
			fmt.Println("Usage:\n\t "+cmd.Use+"\n")
			return
		}

		appName := args[0]
		revision := args[1]
		zipPath := args[2]

		zip, err := os.Open(zipPath)
		if err != nil {
			log.Fatal(err)
		}
		defer zip.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(part, zip)

		writer.WriteField("namespace", orgName)
		writer.WriteField("revision", revision)
		writer.WriteField("application", appName)

		err = writer.Close()
		if err != nil {
			log.Fatal(err)
		}

		req, err := http.NewRequest("POST", clusterTarget + buildPath, body)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer " + authToken)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		response, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatal(err)
		} else {
			// dump response to stdout
			defer response.Body.Close()
			if response.StatusCode >= 200 && response.StatusCode < 300 {
				fmt.Println("\nImage build successful\n")
			}
			_, err := io.Copy(os.Stdout, response.Body)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

var getImageCmd = &cobra.Command{
	Use:   "image <appName> <revision>",
	Short: "retrieves a built image's info'",
	Long: `This command retrieves the build specified by the application name and revision number.

The image must've be built by a successful 'apigeectl build image' command

Example of use:

$ apigeectl get image example 1`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Missing required args\n")
			fmt.Println("Usage:\n\t "+cmd.Use+"\n")
			return
		}

		appName := args[0]
		revision := args[1]

		req, err := http.NewRequest("GET", clusterTarget + imagePath + orgName + "/applications/" + appName + "/images/"+revision, nil)
		req.Header.Set("Authorization", "Bearer " + authToken)
		response, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatal(err)
		} else {
			defer response.Body.Close()
			_, err := io.Copy(os.Stdout, response.Body)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	createCmd.AddCommand(imageCmd)
	getCmd.AddCommand(getImageCmd)
}
