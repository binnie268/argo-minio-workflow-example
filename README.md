This repository contains all the necessary components to run an end to end argo workflow execution.

The project is broken into a few folders:

- csv: Contains the entry csv to feed the argo workflow process. It contains the list of flight prices for NYC to London and NYC to Paris.
- go: Contains the main source code to construct an argo workflow template that can be submitted to the argo server using the golang argo library
- kubernetes: Contains the kubernetes resources that are necessary to successfully run an end to end argo workflow with minio.
  - argo-role.yaml: Gives the proper permissions to argo workflow such as starting a pod, watching for updates, killing a pod and updating workflow status.
  - argo-server-role.yaml: Same as above but for argo-server.
  - workflow-controller-configmap.yaml: Sets up the access to minio.
  - pvc.yaml: Sets up the persistent volume that minio can attach to.
- python: Contains the python source code that to sum, average, and pick the cheapest averaged price flight.
