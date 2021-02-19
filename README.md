# Cloud Foundry Resource

An output only resource (at the moment) that will deploy an application to a
Cloud Foundry deployment.

## Source Configuration

Note: you must provide either `username` and `password` or `client_id` and `client_secret`.

* `api`: *Required.* The address of the Cloud Controller in the Cloud Foundry
  deployment.
* `username`: *Optional.* The username used to authenticate.
* `password`: *Optional.* The password used to authenticate.
* `client_id`: *Optional.* The client id used to authenticate.
* `client_secret`: *Optional.* The client secret used to authenticate.
* `organization`: *Required.* The organization to push the application to.
* `space`: *Required.* The space to push the application to.
* `skip_cert_check`: *Optional.* Check the validity of the CF SSL cert.
  Defaults to `false`.
* `verbose`: *Optional.* Invoke `cf` cli using `CF_TRACE=true` to print all API calls made to Cloud Foundry.

## Behaviour

### `out`: Deploy an application to a Cloud Foundry

Pushes an application to the Cloud Foundry detailed in the source
configuration. A [manifest][cf-manifests] that describes the application must
be specified.

[cf-manifests]: http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html

#### Parameters

* `manifest`: *Required.* Path to a application manifest file.
* `path`: *Optional.* Path to the application to push. If this isn't set then
  it will be read from the manifest instead. When using rolling app deployments, this is required.
* `use_rolling_app_deployment`: *Optional.* Use CC's native rolling deployments feature to upgrade 
  the app without downtime. Default false.
* `current_app_name`: *Optional.* This should be the name of the application
  that this will re-deploy over. If this is set the resource will perform a
  zero-downtime deploy.
* `environment_variables`: *Optional.* It is not necessary to set the variables in [manifest][cf-manifests] if this parameter is set.
* `vars`: *Optional.* Map of variables to pass to manifest
* `vars_files`: *Optional.* List of variables files to pass to manifest
* `docker_username`: *Optional.* This is used as the username to authenticate against a protected docker registry.
* `docker_password`: *Optional.* This should be the users password when authenticating against a protected docker registry.
* `show_app_log`: *Optional.* Tails the app log during startup, useful to debug issues when using blue/green deploys together with the `current_app_name` option.
* `no_start`: *Optional.* Deploys the app but does not start it. This parameter is ignored when `current_app_name` is specified.

## Pipeline example

```yaml
---
jobs:
- name: job-deploy-app
  public: true
  serial: true
  plan:
  - get: resource-web-app
  - task: build
    file: resource-web-app/build.yml
  - put: resource-deploy-web-app
    params:
      manifest: build-output/manifest.yml
      environment_variables:
        key: value
        key2: value2

resources:
- name: resource-web-app
  type: git
  source:
    uri: https://github.com/cloudfoundry-community/simple-go-web-app.git

- name: resource-deploy-web-app
  type: cf
  source:
    api: https://api.run.pivotal.io
    username: EMAIL
    password: PASSWORD
    organization: ORG
    space: SPACE
    skip_cert_check: false
```

## Development

### Prerequisites

* golang is *required* - version 1.9.x is tested; earlier versions may also
  work.
* docker is *required* - version 17.06.x is tested; earlier versions may also
  work.
* godep is used for dependency management of the golang packages.

### Running the tests

The tests have been embedded with the `Dockerfile`; ensuring that the testing
environment is consistent across any `docker` enabled platform. When the docker
image builds, the test are run inside the docker container, on failure they
will stop the build.

Run the tests with the following commands for both `alpine` and `ubuntu` images:

```sh
docker build -t cf-resource -f dockerfiles/alpine/Dockerfile .
docker build -t cf-resource -f dockerfiles/ubuntu/Dockerfile .
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
