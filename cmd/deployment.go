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
	"os"
	"io"
	"log"
	"encoding/json"
	"bytes"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type EnvVar struct {
	Name string
	Value string
}

type Deployment struct {
	DeploymentName string
	PublicHosts string
	PrivateHosts string
	Replicas int64
	PtsUrl string
	EnvVars []EnvVar
}

type DeploymentImageUpdate struct {
	PtsUrl string
}

const (
	NAME = 0
	VALUE = 1
)

var previous bool

// represents the get deployment command
var deploymentCmd = &cobra.Command{
	Use:   "deployment <environmentName> <deploymentName>",
	Short: "retrieves an active deployment's available information'",
	Long: `Given the name of an active deployment, this will retrieve the currently
available information in JSON format.

Example of use:
$ shipyardctl get deployment dep1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()

		if len(args) == 0 {
			fmt.Println("Missing required arg <environmentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		envName = args[0]

		// get all of the active deployments
		if all {
			status := getDeploymentAll(envName)
			if !CheckIfAuthn(status) {
				// retry once more
				status := getDeploymentAll(envName)
				if status == 401 {
					fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
					fmt.Println("Command failed.")
				}
			}
		} else { // get active deployment by name
			if len(args) < 2 {
				fmt.Println("Missing required arg <deplymentName>\n")
				fmt.Println("Usage:\n\t" + cmd.Use + "\n")
				return
			}

			// get deployment name from arguments
			depName = args[1]

			status := getDeploymentNamed(envName, depName, true)
			if !CheckIfAuthn(status) {
				// retry once more
				status := getDeploymentNamed(envName, depName, true)
				if status == 401 {
					fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
					fmt.Println("Command failed.")
				}
			}
		}
	},
}

func getDeploymentNamed(envName string, depName string, printBody bool) int {
	// build API call
	req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName + "/deployments/" + depName, nil)
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

	// dump response body to stdout
	defer response.Body.Close()

	if response.StatusCode != 401 && printBody{
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

func getDeploymentAll(envName string) int {
	req, err := http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName + "/deployments" , nil)
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

var deleteDeploymentCmd = &cobra.Command{
	Use:   "deployment <environmentName> <deploymentName>",
	Short: "deletes an active deployment",
	Long: `Given the name of an active deployment and the environment it belongs to,
this will delete it.

Example of use:
$ shipyardctl delete deployment org1:env1 dep1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()

		// check and pull required arguments
		if len(args) == 0 {
			fmt.Println("Missing required arg <environmentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		envName = args[0]

		if len(args) < 2 {
			fmt.Println("Missing required arg <deplymentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		depName = args[1]

		status := deleteDeployment(envName, depName)
		if !CheckIfAuthn(status) {
			// retry once more
			status := deleteDeployment(envName, depName)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func deleteDeployment(envName string, depName string) int {
	// build API call URL
	req, err := http.NewRequest("DELETE", clusterTarget + enroberPath + "/" + envName + "/deployments/" + depName, nil)
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

	// dump response body to stdout
	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Println("\nDeletion of " + depName + " in " + envName + " was successful\n")
	}

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

// deployment creation command
var createDeploymentCmd = &cobra.Command{
	Use:   "deployment <environmentName> <depName> <publicHost> <privateHost> <replicas> <ptsUrl>",
	Short: "creates a new deployment in the given environment with given name",
	Long: `A deployment requires a name, accepted public and private hosts, the number
of replicas and the URL that locates the appropriate Pod Template Spec built by Shipyard.
It also requires an active environment to deploy to.

Example of use:
$ shipyardctl create deployment org1:env1 dep1 "test.host.name" "test.host.name" 2 "https://pts.url.com" --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()

		// check and pull required args
		if len(args) < 6 {
			fmt.Println("Missing required args\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		envName = args[0]
		depName = args[1]
		publicHost := args[2]
		privateHost := args[3]
		replicas, err := strconv.ParseInt(args[4], 0, 64)
		if err != nil {
			log.Fatal(err)
		}
		ptsUrl := args[5]
		vars := parseEnvVars()

		status := createDeployment(envName, depName, publicHost, privateHost, replicas, ptsUrl, vars, true)
		if !CheckIfAuthn(status) {
			// retry once more
			status := createDeployment(envName, depName, publicHost, privateHost, replicas, ptsUrl, vars, true)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}

		if status == 201 {
			fmt.Println("\nCreation of " + depName + " in " + envName + " was successful")
		}
	},
}

func createDeployment(envName string, depName string, publicHost string, privateHost string, replicas int64, ptsUrl string, vars []EnvVar, printBody bool) int {
	// prepare arguments in a Deployment struct and Marshal into JSON
	js, err := json.Marshal(Deployment{depName, publicHost, privateHost, replicas, ptsUrl, vars})
	if err != nil {
		log.Fatal(err)
	}

	// build API call with request body (deployment information)
	req, err := http.NewRequest("POST", clusterTarget + enroberPath + "/" + envName + "/deployments", bytes.NewBuffer(js))

	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Authorization", "Bearer " + authToken)
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseResponse(response)
	}

	// dump response to stdout
	defer response.Body.Close()

	if response.StatusCode != 401 && printBody {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

// patch/update deployment command
var patchDeploymentCmd = &cobra.Command{
	Use:   "deployment <environmentName> <depName> <updateData>",
	Short: "updates an active deployment",
	Long: `Once deployed, a deployment can be updated by passing a JSON object
with the corresponding mutations. All properties, except for the deployment name are mutable.
That includes, the public or private hosts, replicas, PTS URL entirely, or the PTS itself.

Example of use:
$ shipyardctl patch deployment org1:env1 dep1 '{"replicas": 3, "publicHosts": "test.host.name.patch"}' --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()

		// check and pull required args
		if len(args) < 3 {
			fmt.Println("Missing required args\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		envName = args[0]
		depName = args[1]
		updateData := args[2]

		status := patchDeployment(envName, depName, updateData, true)
		if !CheckIfAuthn(status) {
			// retry once more
			status := patchDeployment(envName, depName, updateData, true)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
		if status >= 200 && status < 300 {
			fmt.Println("\nPatch of " + depName + " in " + envName + " was successful\n")
		}
	},
}

func patchDeployment(envName string, depName string, updateData string, printBody bool) int {
	// build API call
	// the update data will come in from command line as a JSON string
	req, err := http.NewRequest("PATCH", clusterTarget + enroberPath + "/" + envName + "/deployments/"+depName, bytes.NewBuffer([]byte(updateData)))

	req.Header.Set("Authorization", "Bearer " + authToken)
	if verbose {
		PrintVerboseRequest(req)
	}

	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		PrintVerboseResponse(response)
	}

	defer response.Body.Close()

	if response.StatusCode != 401 && printBody {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

var logsCmd = &cobra.Command{
	Use:   "logs <environmentName> <deploymentName>",
	Short: "retrieves an active deployment's available logs",
	Long: `Given the name of an active deployment, this will retrieve the currently
available logs.

Example of use:
$ shipyardctl get logs org1:env1 dep1 --token <token>`,
	Run: func(cmd *cobra.Command, args []string) {
		RequireAuthToken()

		if len(args) == 0 {
			fmt.Println("Missing required arg <environmentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		envName = args[0]

		if len(args) < 2 {
			fmt.Println("Missing required arg <deplymentName>\n")
			fmt.Println("Usage:\n\t" + cmd.Use + "\n")
			return
		}

		// get deployment name from arguments
		depName = args[1]

		status := getDeploymentLogs(envName, depName)
		if !CheckIfAuthn(status) {
			// retry once more
			status := getDeploymentLogs(envName, depName)
			if status == 401 {
				fmt.Println("Unable to authenticate. Please check your SSO target URL is correct.")
				fmt.Println("Command failed.")
			}
		}
	},
}

func getDeploymentLogs(envName string, depName string) int {
	var req *http.Request
	var err error
	// build API call
	if previous {
		req, err = http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName + "/deployments/" + depName + "/logs?previous=true", nil)
	} else {
		req, err = http.NewRequest("GET", clusterTarget + enroberPath + "/" + envName + "/deployments/" + depName + "/logs", nil)
	}
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

	// dump response body to stdout
	defer response.Body.Close()

	if response.StatusCode != 401 {
		_, err = io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	return response.StatusCode
}

func init() {
	getCmd.AddCommand(deploymentCmd)
	deploymentCmd.Flags().BoolVarP(&all, "all", "a", false, "Retrieve all deployments")

	getCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&previous, "previous", "p", false, "used to retrieve previous container's logs")

	deleteCmd.AddCommand(deleteDeploymentCmd)
	createCmd.AddCommand(createDeploymentCmd)
	createDeploymentCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Environment variables to set in the deployment")
	patchCmd.AddCommand(patchDeploymentCmd)
}

func parseEnvVars() (parsed []EnvVar) {
	var temp string

	if len(envVars) > 0 {
		for i := range envVars {
			temp = envVars[i]
			split := strings.Split(temp, "=")
			parsed = append(parsed, EnvVar{split[NAME], split[VALUE]})
		}
	} else {
		return []EnvVar{}
	}

	return parsed
}
