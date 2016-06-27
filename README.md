#shipyardctl

This project is a command line interface that wraps the Shipyard build and deploy APIs.

**This is not meant to replace `kubectl`, but merely to wrap the many available API resources of Shipyard**

###Local Build
```sh
git clone https://github.com/30x/shipyardctl.git
cd shipyardctl
go install
```

###Environment
`shipyardctl` expects the following three environment variables be in place in order to use it.

- `APIGEE_ORG`: Your Apigee org name
- `APIGEE_ENVIRONMENT_NAME`: Your Apigee env name
- `APIGEE_TOKEN`: Your JWT access token generated from Apigee credentials
- `CLUSTER_TARGET`: The _protocol_ and _host name_ of the k8s cluster (i.e. "http://test.cluster.name")

###Usage

The list of available commands is as follows:
```
  ▾ shipyardctl
    ▾ image
        create
        get
    ▾ environment
        create
        get
        patch
        delete
    ▾ deployment
        create
        get
        patch
        delete
    ▾ bundle
        create
```

Please also see `shipyardctl --help` for more information on the available commands and their arguments.

####Walk through

**Build an image of a Node.js app**
```sh
> shipyardctl create image "example" 1 "./example-app.zip"
> export PTS_URL="<copy the Pod Template Spec URL generated and output by the build image command>"
```
The build command takes the name of your application, the revision number and the path to your zipped Node app.
_Note: there must be a valid package.json in the root of zipped application_

**Verify image creation**
```sh
> shipyardctl get image example 1
```
This merely retrieves the available information for the image specified by the applicaiton name and revision number

**Create a new environment**
```sh
> shipyardctl create environment "test" "test.host.name1" "test.host.name2"
```
Here we create a new environment with the name "test" and the accepted hostnames of "test.host.name1" and "test.host.name2"

**Retrieve the newly created environment by name**
```sh
> shipyardctl get environment "test"
```
Here we have retrieved the newly created environment, by name.

**Update the environment's set of accepted hostnames**
```sh
> shipyardctl patch environment "test" "test.host.name3" "test.host.name4"
```
The environment "test" will be updated to accept traffic from the following hostnames, explicitly.

**Create a new deployment**
```sh
> export PUBLIC_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> export PRIVATE_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> shipyardctl create deployment "test" "example" $PUBLIC_HOST $PRIVATE_HOST 1 $PTS_URL
```
This creates a new deployment within the "test" environment with the previously generated PTS URL. The number 1 represents the number
of replicas to be made and "example" is the name of the deployment.

**Retrieve newly created deployment by name**
```sh
> shipyardctl get deployment "test" "example"
```
The response will include all available information on the active deployment in the given environment.

**Update the deployment**
```sh
> shipyardctl patch deployment "test" "example" '{"replicas": 3, "publicHosts": "replacement.host.name"}'
```
Updating a deployment by name, in a given environment, takes a JSON string that includes the properties to be changed.
This includes:
- number of replicas
- public host
- private host
- pod template spec URL
- pod template spec

**Create Apigee Edge Proxy bundle**
```sh
> shipyardctl create bundle "example" --save ~/Desktop
```
This command, given your environment is configured and your applications name, will generate a valid proxy bundle for
the application that is deployed on Shipyard. Zip this folder (named "apiproxy"), name it with your application name, and
upload this to Apigee Edge.

> We are unable to zip the bundle for you as the zip generated by the native Go lang `archive/zip` package is not compatible
> with native Java zip packages. See [this forum](http://webmail.dev411.com/p/gg/golang-nuts/155g3s6g53/go-nuts-re-zip-files-created-with-archive-zip-arent-recognised-as-zip-files-by-java-util-zip) for an explanation.

**Delete the deployment**
```sh
> shipyardctl delete deployment "test" "example"
```
This deletes the named deployment.

**Delete the environment**
```sh
> shipyardctl delete environment "test"
```
This deletes the named deployment