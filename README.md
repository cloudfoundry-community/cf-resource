# Cloud Foundry Resource

An output only resource (at the moment) that will deploy an application to a
Cloud Foundry deployment.

## Source Configuration

* `api`: *Required.* The address of the Cloud Controller in the Cloud Foundry
  deployment.
* `username`: *Required.* The username used to authenticate.
* `password`: *Required.* The password used to authenticate.
* `organization`: *Optional\*.* The organization to push the application to. (One
  of `organization` or `organisation` must be specified.)
* `organisation`: *Optional\*.* The organisation to push the application to. (One
  of `organization` or `organisation` must be specified.)
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
