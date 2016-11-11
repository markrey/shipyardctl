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
	"text/template"
	"os"
	"io/ioutil"
	"path/filepath"
	"github.com/30x/zipper"
	"github.com/30x/shipyardctl/utils"
	"github.com/30x/shipyardctl/mgmt"

	"github.com/spf13/cobra"
)

type Bundle struct {
	Name string
	BasePath string
	PublicPath string
}

var savePath string
var base string
var publicPath string
var fileMode os.FileMode

var uploadBundleCmd = &cobra.Command{
	Use:   "upload-bundle",
	Short: "upload an Edge proxy bundle",
	Long: `This uploads the target API proxy bundle archive.

Example of use:

$ shipyardctl upload-bundle -o myOrg -e myEnv -z /path/to/bundle.zip -n proxyName`,
	Run: func(cmd *cobra.Command, args []string) {
		err := uploadProxy(org, env, name, zipPath)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
	PreRunE: func(cmd * cobra.Command, args []string) error {
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
			return fmt.Errorf("Missing required flag \"--zip-path\"")
		}

		return nil
	},
}

func uploadProxy(org string, env string, name string, dir string) (err error) {
	var mgmtTarget string
	if config.IsUsingE2ELogin() {
		mgmtTarget = "https://api.e2e.apigee.net"
	} else {
		mgmtTarget = "https://api.enterprise.apigee.com"
	}

	list, err := mgmt.ListProxies(mgmtTarget, org, config.GetCurrentToken())
	if err != nil { return }

	if utils.ContainsString(name, list) {
		return fmt.Errorf("Error: there already exists a proxy in %s with name %s\n", org, name)
	}

	err = mgmt.UploadProxyBundle(mgmtTarget, org, env, config.GetCurrentToken(), dir, name, verbose)
	if err == nil {
		fmt.Println("\nSuccessfully uploaded proxy bundle!")
	}

	return
}

// bundleCmd represents the bundle command
var bundleCmd = &cobra.Command{
	Use:   "bundle <name>",
	Short: "generate an Edge proxy bundle",
	Long: `This generates the appropriate API proxy bundle for an
environment built and deployed on Shipyard.

Example of use:

$ shipyardctl create bundle exampleName`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Missing required arg.")
			fmt.Println("Usage:\t\n"+cmd.Use)
			return
		}

		name := args[0]

		// make a temp dir
		tmpdir, err := ioutil.TempDir("", orgName+"_"+envName)
		if err != nil {
			fmt.Println("Failed to make a temporary directory")
			return
		}

		if verbose {
			fmt.Println("Creating tmpdir at: "+tmpdir)
		}

		defer os.RemoveAll(tmpdir)

		zipDir, err := makeBundle(name, tmpdir, base, publicPath)
		checkError(err, "Failed to make proxy zip")

		// move zip to designated savePath
		if savePath != "" {
			err = os.Rename(zipDir, filepath.Join(savePath, name+".zip"))
			if verbose {
				fmt.Println("Moving proxy folder to "+savePath)
			}
			checkError(err, "Unable to move apiproxy to target save directory")
		} else { // move apiproxy from tmpdir to cwd
			cwd, err := os.Getwd()
			err = os.Rename(zipDir, filepath.Join(cwd, name+".zip"))
			if verbose {
				fmt.Println("Moving proxy folder to CWD")
			}
			checkError(err, "Unable to move apiproxy bundle to cwd")
		}

		if verbose {
				fmt.Println("Deleting tmpdir")
			}
	},
}


func checkError(err error, customMsg string) {
	if err != nil {
		if customMsg != "" {
			fmt.Println(customMsg)
		}

		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
}

func init() {
	createCmd.AddCommand(bundleCmd)
	bundleCmd.Flags().StringVarP(&savePath, "save", "s", "", "Save path for proxy bundle")
	bundleCmd.Flags().StringVarP(&base, "basePath", "b", "", "Proxy base path. Defaults to /")
	bundleCmd.Flags().StringVarP(&publicPath, "publicPath", "p", "/", "Application public path. Defaults to /")


	RootCmd.AddCommand(uploadBundleCmd)
	uploadBundleCmd.Flags().StringVarP(&org, "org", "o", "", "Apigee org to deploy application to")
  uploadBundleCmd.Flags().StringVarP(&env, "env", "e", "", "Apigee environment within the org to deploy application to")
	uploadBundleCmd.Flags().StringVarP(&base, "basePath", "b", "", "Proxy base path. Defaults to /")
	uploadBundleCmd.Flags().StringVarP(&zipPath, "zip-path", "z", "", "Path to proxy bundle zip")
	uploadBundleCmd.Flags().StringVarP(&name, "name", "n", "", "name of the proxy being deployed")

	fileMode = 0755
}

func makeBundle(name string, tmpdir string, bp string, pubPath string) (zipPath string, err error) {
	// make apiproxy directory structure
	dir := filepath.Join(tmpdir, "apiproxy")
	err = os.Mkdir(dir, fileMode)

	if verbose {
		fmt.Println("Creating folder 'apiproxy' at: "+dir)
	}
	checkError(err, "Unable to make root apiproxy dir")

	proxiesDirPath := filepath.Join(dir, "proxies")
	err = os.Mkdir(proxiesDirPath, fileMode)
	if verbose {
		fmt.Println("Creating subfolder 'proxies' at: "+proxiesDirPath)
	}
	checkError(err, "Unable to make proxies dir")

	targetsDirPath := filepath.Join(dir, "targets")
	err = os.Mkdir(targetsDirPath, fileMode)
	if verbose {
		fmt.Println("Creating subfolder 'targets' at: "+targetsDirPath)
	}
	checkError(err, "Unable to make targets dir")

	policiesDirPath := filepath.Join(dir, "policies")
	err = os.Mkdir(policiesDirPath, fileMode)
	if verbose {
		fmt.Println("Creating subfolder 'policies' at: "+policiesDirPath)
	}
	checkError(err, "Unable to make policies dir")

	// bundle user info for templates
	if bp == "" {
		bp = pubPath
	}

	bundle := Bundle{name, bp, pubPath}

	// example.xml --> ./apiproxy/
	proxy_xml, err := os.Create(filepath.Join(dir, name+".xml"))
	err = proxy_xml.Chmod(fileMode)
	if verbose {
		fmt.Println("Creating file '"+name+".xml'")
	}
	checkError(err, "Unable to make "+name+".xml file")

	proxyTmpl, err := template.New("PROXY").Parse(PROXY_XML)
	if err != nil { return "", err }
	err = proxyTmpl.Execute(proxy_xml, bundle)
	if err != nil { return "", err }

	// AddCors.xml --> ./apiproxy/policies
	add_cors_xml, err := os.Create(filepath.Join(policiesDirPath, "AddCors.xml"))
	err = add_cors_xml.Chmod(fileMode)
	if verbose {
		fmt.Println("Creating file 'policies/AddCors.xml'")
	}
	checkError(err, "Unable to make AddCors.xml file")

	addCors, err := template.New("ADD_CORS").Parse(ADD_CORS)
	if err != nil { return "", err }
	err = addCors.Execute(add_cors_xml, bundle)
	if err != nil { return "", err }

	// default.xml --> ./apiproxy/proxies && ./apiproxy/targets
	proxy_default_xml, err := os.Create(filepath.Join(proxiesDirPath, "default.xml"))
	err = proxy_default_xml.Chmod(fileMode)
	if verbose {
		fmt.Println("Creating file 'proxies/default.xml'")
	}
	checkError(err, "Unable to make default.xml file")

	target_default_xml, err := os.Create(filepath.Join(targetsDirPath, "default.xml"))
	err = target_default_xml.Chmod(fileMode)
	if verbose {
		fmt.Println("Creating file 'targets/default.xml'")
	}
	checkError(err, "Unable to make default.xml file")

	proxyEndpoint, err := template.New("PROXY_ENDPOINT").Parse(PROXY_ENDPOINT)
	if err != nil { return "", err }
	err = proxyEndpoint.Execute(proxy_default_xml, bundle)
	if err != nil { return "", err }

	targetEndpoint, err := template.New("TARGET_ENDPOINT").Parse(TARGET_ENDPOINT)
	err = targetEndpoint.Execute(target_default_xml, bundle)
	if err != nil { return "", err }

	zipDir := filepath.Join(tmpdir, name+".zip")
	err = zipper.Archive(dir, zipDir)
	if err != nil { return "", err }

	return zipDir, nil
}
