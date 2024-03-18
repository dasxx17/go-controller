# go-controller

## Prerequisites
1. Kubernetes Cluster
2. Go (tested on v1.22.1)
3. kubectl

## Example Kubernetes Cluster Setup with minikube
Any local kube cluster deployment mechanism can be used for setup.  To create the cluster using minikube, following the steps below:
1. [Install minikube](https://minikube.sigs.k8s.io/docs/start/)
2. Create the cluster by running `minikube start`

## Compilation
To start the controller, run:
* `make run`

## Verification
To apply the sample resource, run:
* `make apply`

To port-forward the deployment, run:
* `make port-forward`

To view the UI, open your browser and navigate to:
* `localhost:9098`

To create a sample POST, run:
* `make post-test`

To test a sample GET, run:
* `make get-test`

### Updating
Feel free to update the values in the sample CRD file `sample-myappresource.yaml` and rerun `make apply`

## Cleanup
To delete the resources, run:
* `make delete`

## Running Tests
To run tests, run:
* `make test`

## Controller Code
The controller code is located in `internal/controller/myappresource_controller.go`.
