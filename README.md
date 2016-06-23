#apigeectl

This project is a command line interface that wraps the Shipyard build and deploy APIs.

**This is not meant to replace `kubectl`, but merely to wrap the many available API resources of Shipyard**

###Local Build
```sh
git clone https://github.com/30x/apigeectl.git
cd apigeectl
go install
```

###Environment
`apigeectl` expects the following three environment variables be in place in order to use it.

- `APIGEE_ORG`: Your Apigee org name
- `APIGEE_ENVIRONMENT_NAME`: Your Apigee env name
- `APIGEE_TOKEN`: Your JWT access token generated from Apigee credentials
- `CLUSTER_TARGET`: The _protocol_ and _host name_ of the k8s cluster (i.e. "http://test.cluster.name")

###Usage

The list of available commands is as follows:
```
  ▾ apigeectl
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
```

Please also see `apigeectl --help` for more information on the available commands and their arguments.

####Walk through

**Build an image of a Node.js app**
```sh
> apigeectl create image "example" 1 "./example-app.zip"
> export PTS_URL="<copy the Pod Template Spec URL generated and output by the build image command>"
```
The build command takes the name of your application, the revision number and the path to your zipped Node app.
_Note: there must be a valid package.json in the root of zipped application_

**Verify image creation**
```sh
> apigeectl get image example 1
```
This merely retrieves the available information for the image specified by the applicaiton name and revision number

**Create a new environment**
```sh
> apigeectl create environment "test" "test.host.name1" "test.host.name2"
```
Here we create a new environment with the name "test" and the accepted hostnames of "test.host.name1" and "test.host.name2"

**Retrieve the newly created environment by name**
```sh
> apigeectl get environment "test"
```
Here we have retrieved the newly created environment, by name.

**Update the environment's set of accepted hostnames**
```sh
> apigeectl patch environment "test" "test.host.name3" "test.host.name4"
```
The environment "test" will be updated to accept traffic from the following hostnames, explicitly.

**Create a new deployment**
```sh
> export PUBLIC_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> export PRIVATE_HOST "$APIGEE_ORG-$APIGEE_ENVIRONMENT_NAME.apigee.net"
> apigeectl create deployment "test" "example" $PUBLIC_HOST $PRIVATE_HOST 1 $PTS_URL
```
This creates a new deployment within the "test" environment with the previously generated PTS URL. The number 1 represents the number
of replicas to be made and "example" is the name of the deployment.

**Retrieve newly created deployment by name**
```sh
> apigeectl get deployment "test" "example"
```
The response will include all available information on the active deployment in the given environment.

**Update the deployment**
```sh
> apigeectl patch deployment "test" "example" '{"replicas": 3, "publicHosts": "replacement.host.name"}'
```
Updating a deployment by name, in a given environment, takes a JSON string that includes the properties to be changed.
This includes:
- number of replicas
- public host
- private host
- pod template spec URL
- pod template spec

**Delete the deployment**
```sh
> apigeectl delete deployment "test" "example"
```
This deletes the named deployment.

**Delete the environment**
```sh
> apigeectl delete environment "test"
```
This deletes the named deployment