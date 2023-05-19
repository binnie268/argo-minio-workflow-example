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

Here are the step by step instructions to create an end to end Azure Kubernetes Services infrastructure running Argo workflow and Minio:
1.	Create an Azure free account and subscription. New users will get $200 credit and additional free services for 30 days.
2.	Once your subscription is ready, let’s start by creating a resource group called argominio-rg.
3.	Create an AKS by following the instructions here, remember to  use the existing resource group we just created.
a.	NOTES: You can select the free AKS pricing tier, and the Standard B2s node size, with 1 Node count range to save on cost.
4.	Once the Kubernetes cluster is created follow these instructions to connect to it locally, make sure you have a local Kubernetes env installed.
a.	az login (follow the interactive sign in)
b.	az aks get-credentials --resource-group myResourceGroup --name myAKSCluster
5.	We will utilize Azure File Share as the main storage and connect minio to it.
6.	First, we deploy a PVC Claim. This will dynamically create a Storage account and File Share for you in Azure. It will create it in the same resource as your AKS resource Group:
    ```
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: minio-pvc
      namespace: minio
    spec:
      accessModes:
        - ReadWriteMany
      storageClassName: managed-csi
      resources:
        requests:
          storage: 1Gi
    ```
7.	Now find the resource group starting with MC_ which your Kubernetes resource creation created. You will see a storage account type and inside it is your file share.
8.	Now we will deploy argo, minio, and give it the proper permissions.
9.	Deploy minio using the existing PVC storage that we just created mounted to Azure File Share: 
    ```
    helm install --namespace minio minio --set persistence.existingClaim=minio-pvc bitnami/minio
    ```
b.	Please note down the username and password that came with it. You will need it shortly. 
10.	Install argo-workflows into Kubernetes with this command:
    ```
    kubectl create namespace argo
    kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.3.10/install.yaml
    ```
11.	You can make sure that the two tools are deployed into Kubernetes by running:
    ```
    kubectl get pods -n argo
    kubectl get pods -n minio
    ```	
12.	Now we will put our first file into minio to prepare for an argo workflow run (in an actual application you would automate this process)
13.	The csv downloaded previously will be needed in these next steps
14.	Now, port forward your minio console so that you can access it through the UI:
    ```
    kubectl port-forward service/minio 9001 --namespace minio
    ```
15.	Open up a browser and go to http://localhost:9001, enter the username and password noted down earlier for minio.
16.	Create a bucket called argo, and a folder inside it called ‘inputs’.
17.	Upload the csv file downloaded from the github to ‘inputs’ folder.
18.	Next, we want to get argo ready to have permissions to execute workflows and access minio.
    1. Create a minio secret and replace the username and password with the one you have saved earlier:
      ```
      apiVersion: v1
      kind: Secret
      metadata:
        name: minio
        namespace: argo
      type: Opaque
      data:
        root-password: replace-this-with-yours
        root-user: replace-this-with-yours
      ```
    2. Configure argo’s s3 repository by implementing the following. Run the following command to replace the current resource: 
      ```kubectl create configmap -o yaml --dry-run | kubectl apply -f -```
      ```
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: workflow-controller-configmap
        namespace: argo
      data:
        config: |
          instanceID: my-ci-controller
          artifactRepository:
            archiveLogs: true
            s3:
              endpoint: minio.minio:9000
              bucket: argo
              region: us-east-2
              insecure: true
              accessKeySecret:
                name: minio
                key: root-user
              secretKeySecret:
                name: minio
                key: root-password
         ```
    3. Configure the proper roles for the argo workflow controller:
       1. Download the yaml here.
ii.	On the same directory as the yaml file, run kubectl create role -o yaml --dry-run | kubectl apply -f -
d.	Configure the proper roles for the argo server:
i.	Down the yaml here.
ii.	On the same directory as the yaml file run, kubectl create role -o yaml --dry-run | kubectl apply -f -
e.	Now, clone the main.go code here. Make sure you have go installed.
f.	Expose the argo console by running this command: kubectl -n argo port-forward deployment/argo-server 2746:2746
g.	Open a browser and go to localhost:2746
h.	On the same path as the main.go, run “go run main.go”
i.	This program will construct an argo workflow template containing the instructions to sum flight prices, average them, and then pick a flight and submit it to argo. Argo will then orchestrate the management of each pod and take care of everything for us. It automates the process from adding the flight prices, to giving the results to the next step which is to average the flight prices, and then finally decide on a destination. All we have to do is just construct and submit a template of instructions to argo. That is the real beauty of argo .
j.	Check in the argo console for the workflow you just ran in the “argo” namespace
k.	After it has finished running, check minio console under argo/outputs for the flight decision and and localhost:2746 will show the workflow completed as below.
![image](https://github.com/binnie268/argo-minio-workflow-example/assets/29080449/d6694de1-bcea-45f3-883f-47d1daa165e4)

