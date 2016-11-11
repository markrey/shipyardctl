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
  "strconv"
  "os"
  "io/ioutil"
  "encoding/json"
  "bufio"
  "strings"

  "github.com/spf13/cobra"
)

var name string
var org string
var env string
var bp string
var zipPath string
var revision int
var port int

var deployNodeContainerCmd = &cobra.Command{
  Use:   "deploy-node-container",
  Short: "deploys a given Node.js bundle to Shipyard and creates the proper Edge proxy",
  Long: `This command consumes a Node.js application archive, deploys it to Shipyard,
and creates an appropriate Edge proxy.

$ shipyardctl deploy-node-container -o myOrg -e myEnv -b /basePath --node-version 5`,
  PreRunE: func(cmd *cobra.Command, args []string) error {
    RequireAuthToken()

    if org == "" {
      return fmt.Errorf("Missing required flag \"--org\"")
    }

    if env == "" {
      return fmt.Errorf("Missing required flag \"--env\"")
    }

    if name == "" {
      return fmt.Errorf("Missing required flag \"--name\"")
    }

    if zipPath == "" {
      return fmt.Errorf("Missing required flag \"--directory\"")
    }

    return nil
  },
  Run: func(cmd *cobra.Command, args []string) {
    if err := buildProvisionDeploy(); err != nil {
      fmt.Println(err)
      return
    }
  },
}

func buildProvisionDeploy() error {
  orgName = org
  MakeBuildPath() // make build service path with orgName as imagespace
  var importOverwrite bool

  // check for existing image
  fmt.Printf("\nChecking for already existing application revision: %s %d\n", name, revision)
  status := getImageRevision(name, strconv.Itoa(revision), verbose)
  if status == 200 {
    if importOverwrite = askToOverwriteImportRevision(name); importOverwrite {
      // delete previous image version
      status = deleteImage(name, strconv.Itoa(revision), verbose)
      if status != 200 {
        return fmt.Errorf("Error deleting application revision for overwrite: received bad status code %d", status)
      }
    } else {
      fmt.Printf("\nUsing existing image revision\n")
    }
  } else if status == 404 { // 404 means there is no image by this name, so we can make one. Anything else is bad
    importOverwrite = true
  } else  {
    return fmt.Errorf("Error checking for existing application revision: received bad status code %d", status)
  }

  pubPath := fmt.Sprintf("%d:%s", port, bp) // this needs to go away some how

  // build image
  if importOverwrite {
    fmt.Printf("\nImporting application: %s %d\n", name, revision)
    status = createImage(name, strconv.Itoa(revision), pubPath, zipPath, verbose)
    if status != 201 {
      return fmt.Errorf("Error importing application: received bad status code %d", status)
    }
  }

  // begin environment creation
  shipyardEnv := fmt.Sprintf("%s:%s", org, env)
  var hostname string
  if config.IsUsingE2ELogin() { // make a hostname based on org/env used for proxy
    hostname = fmt.Sprintf("%s-%s.e2e.apigee.net", org, env)
  } else {
    hostname = fmt.Sprintf("%s-%s.apigee.net", org, env)
  }

  // check for existing environment, if there is already an environment, we can continue
  fmt.Printf("\nChecking for existing environment: %s\n", shipyardEnv)
  status = getEnvironment(shipyardEnv, verbose)
  if status == 500 { // no previously existing environment
    status = createEnv(shipyardEnv, []string{hostname}, verbose) // create new environment
    if status != 201 { // environment creation failed
      return fmt.Errorf("Error provisioning application hosting environment: received status code %d", status)
    } else { // creation succeeded
      fmt.Println("\nCreation of " + shipyardEnv + " was successful\n")
    }
  }

  pts := buildPTSURL(org, name, revision, pubPath)

  // check for existing deployment
  fmt.Printf("Checking for existing deployment of %s in %s\n", name, shipyardEnv)
  status = getDeploymentNamed(shipyardEnv, name, verbose)
  if status == 200 { // existing deployment, stop
    if askToUpateDeployment(name) { //prompt to replace with new revision
      updateBody, err := json.Marshal(&DeploymentImageUpdate{pts})
      if err != nil { return err }

      // update deployment with new PTS URL revision
      fmt.Printf("\nUpdating \"%s\" with revision %d\n", name, revision)
      status = patchDeployment(shipyardEnv, name, string(updateBody), verbose)
      if status != 200 { // failed deployment update
        return fmt.Errorf("Error updating deployment: received status code %d", status)
      }
      fmt.Printf("\nUpdating \"%s\" with revision %d was successful!\n", name, revision)
    }
  } else if status == 500 { // create new deployment
    status = createDeployment(shipyardEnv, name, hostname, hostname, 1, pts, []EnvVar{}, verbose)
    if status != 201 { // creation failed
      return fmt.Errorf("Error creating deployment: received status code %d", status)
    } else { // creation succeeded
      fmt.Println("\nCreation of " + name + " in " + shipyardEnv + " was successful")
    }
  } else {
    return fmt.Errorf("Error checking for existing deployment: received status code %d", status)
  }

  // build a proxy bundle zip and upload to Edge
  err := buildAndUploadProxy(org, env, name, bp, bp)
  if err != nil { return err }

  return nil
}

func buildAndUploadProxy(org string, env string, name string, bp string, pubPath string) (err error) {
  // make a temporary folder for workspace
  tmpdir, err := ioutil.TempDir("", orgName+"_"+envName)
  if err != nil { return }
  defer os.RemoveAll(tmpdir)

  // make proxy bundle zip in a temp folder
  zipDir, err := makeBundle(name, tmpdir, bp, pubPath)
  if err != nil { return }

  // check for existing, then upload proxy bundle to Edge
  err = uploadProxy(org, env, name, zipDir)
  if err != nil { return }

  return
}

// this is hacky and shouldn't have to happen
func buildPTSURL(space string, app string, rev int, pubpath string) string {
  if config.IsUsingE2ELogin() {
    return fmt.Sprintf(clusterTarget + "/imagespaces/generatepodspec?imageURI=977777657611.dkr.ecr.us-west-2.amazonaws.com/%s/%s:%d&publicPath=%s",
      space, app, rev, pubpath)
  } else {
    return fmt.Sprintf(clusterTarget + "/imagespaces/generatepodspec?imageURI=414481387506.dkr.ecr.us-east-1.amazonaws.com/%s/%s:%d&publicPath=%s",
      space, app, rev, pubpath)
  }
}

func askToUpateDeployment(name string) bool {
  consolereader := bufio.NewReader(os.Stdin)
  fmt.Printf("\nThere already exists a deployment for \"%s\". Replace with this revision? (yes/no): ", name)

  input, err := consolereader.ReadString('\n')
  if err != nil {
    fmt.Print(err)
    os.Exit(1)
  }

  update := strings.TrimSpace(input)
  return update == "yes"
}

func askToOverwriteImportRevision(name string) bool {
  consolereader := bufio.NewReader(os.Stdin)
  fmt.Printf("\nThere already exists an import revision for \"%s\". Replace with this upload? (yes/no): ", name)

  input, err := consolereader.ReadString('\n')
  if err != nil {
    fmt.Print(err)
    os.Exit(1)
  }

  update := strings.TrimSpace(input)
  return update == "yes"
}

func init() {
  RootCmd.AddCommand(deployNodeContainerCmd)
  deployNodeContainerCmd.Flags().StringVarP(&org, "org", "o", "", "Apigee org to deploy application to")
  deployNodeContainerCmd.Flags().StringVarP(&env, "env", "e", "", "Apigee environment within the org to deploy application to")
  deployNodeContainerCmd.Flags().StringVarP(&bp, "base-path", "b", "/", "base path of the application being deployed")
  deployNodeContainerCmd.Flags().StringVarP(&nodeVersion, "node-version", "i", "4", "Version of node as supported by mhart/alpine-node images")
  deployNodeContainerCmd.Flags().StringVarP(&name, "name", "n", "", "name of the application being deployed")
  deployNodeContainerCmd.Flags().StringVarP(&zipPath, "directory", "d", "", "Path to archive containing source code zip file")
  deployNodeContainerCmd.Flags().IntVarP(&revision, "revision", "r", 1, "Revision of the application")
  deployNodeContainerCmd.Flags().IntVarP(&port, "port", "p", 9000, "Port the application listens on")
}
