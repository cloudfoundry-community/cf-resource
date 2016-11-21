# Cloud Foundry Resource

An output only resource (at the moment) that will deploy an application to a
Cloud Foundry deployment.

## Source Configuration

* `api`: *Required.* The address of the Cloud Controller in the Cloud Foundry
  deployment.
* `username`: *Required.* The username used to authenticate.
* `password`: *Required.* The password used to authenticate.
* `organization`: *Required.* The organization to push the application to.
* `space`: *Required.* The space to push the application to.
* `skip_cert_check`: *Optional.* Check the validity of the CF SSL cert.
  Defaults to `false`.

## Behaviour

### `out`: Deploy an application to a Cloud Foundry

Pushes an application to the Cloud Foundry detailed in the source
configuration. A [manifest][cf-manifests] that describes the application must
be specified.

[cf-manifests]: http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html

#### Parameters

* `manifest`: *Required.* Path to a application manifest file.
* `path`: *Optional.* Path to the application to push. If this isn't set then
  it will be read from the manifest instead.
* `current_app_name`: *Optional.* This should be the name of the application
  that this will re-deploy over. If this is set the resource will perform a
  zero-downtime deploy.
* `new_app_name`: *Optional.* This is the name of the app in cloud foundry. 
  It will override the `name` value in the manifest file.

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
      new_app_name: job-deploy-app-test
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
