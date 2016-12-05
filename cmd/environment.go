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
	"log"
	"io"
	"os"
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"
)

type Environment struct {
	EnvironmentName string
	HostNames []string
}

type EnvironmentPatch struct {
	HostNames []string
}

// environmentCmd represents the environment command

var environmentCmd = &cobra.Command{
	Use:   "environment <environmentName>",
	Short: "retrieves either active environment information",
	Long: `Given an environment name, this will retrieve the available information of the
active environment(s) in JSON format. Example usage looks like:

$ shipyardctl get environment org1:env1

OR

$ shipyardctl get environment org1:env1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()

		if len(args) == 0 {
			fmt.Println("Missing required arg <environmentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		envName = args[0]
		status := getEnvironment(envName)
		if !CheckIfAuthn(status) {
			// retry once more
			status := getEnvironment(envName)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func getEnvironment(envName string) int {
	req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName, nil)
	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer " + authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseResponse(response)
	}

	defer response.Body.Close()

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

var deleteEnvCmd = &cobra.Command{
	Use:   "environment <environmentName>",
	Short: "deletes an active environment",
	Long: `Given the name of an active environment, this will delete it.

Example of use:
$ shipyardctl delete environment org1:env1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()

		if len(args) == 0 {
			fmt.Println("Missing required arg <environmentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		envName = args[0]
		status := deleteEnv(envName)
		if !CheckIfAuthn(status) {
			// retry once more
			status := deleteEnv(envName)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func deleteEnv(envName string) int {
	req, err := http.NewRequest("DELETE", clusterTarget + enroberPath + "/" + envName, nil)
	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer " + authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
			PrintVerboseResponse(response)
		}

	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Println("\nDeletion of " + envName + " was successful\n")
	}

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

var patchEnvCmd = &cobra.Command{
	Use:   "environment <environmentName> <hostnames...>",
	Short: "update an active environment",
	Long: `Given the name of an active environment and a space delimited
set of hostnames, the environment will be updated. A patch of the hostnames
will replace them entirely.

Example of use:
$ shipyardctl patch org1:env1 "test.host.name3" "test.host.name4" --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()

		if len(args) == 0 {
			fmt.Println("Missing required arg <environmentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		envName = args[0]

		if len(args) < 2 {
			fmt.Println("Missing required arg(s) <hostnames...>")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		hostnames := args[1:]
		status := patchEnv(envName, hostnames)
		if !CheckIfAuthn(status) {
			// retry once more
			status := patchEnv(envName, hostnames)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func patchEnv(envName string, hostnames []string) int {
	js, _ := json.Marshal(EnvironmentPatch{hostnames})

	req, err := http.NewRequest("PATCH", clusterTarget + enroberPath + "/" + envName, bytes.NewBuffer(js))

	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer " + authToken)
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseResponse(response)
	}

	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Println("\nPatch of " + envName + " was successful\n")
	}

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

func init() {
	getCmd.AddCommand(environmentCmd)

	deleteCmd.AddCommand(deleteEnvCmd)
	patchCmd.AddCommand(patchEnvCmd)
}
