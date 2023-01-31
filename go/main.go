package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	argoClient "github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	"github.com/argoproj/pkg/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/utils/pointer"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
)
var namespace = "argo"
var labelMap = map[string]string{"workflows.argoproj.io/controller-instanceid": "my-ci-controller"}

var determineFlightWorkflow = wfv1.Workflow{
	ObjectMeta: metav1.ObjectMeta{
		GenerateName: "hello-world",
		Namespace: namespace,
		Labels: labelMap,
	},
	Spec: wfv1.WorkflowSpec{
		Entrypoint: "determine-flight",
		ServiceAccountName: "argo",
		Templates: []wfv1.Template{
			{	
				Name: "determine-flight",
				DAG: &wfv1.DAGTemplate{
					Tasks: []wfv1.DAGTask{
						{
							Name: "add",
							Template: "add-flight-prices",
						},
						{
							Name:    "avg",
							Dependencies: []string{"add"},
							Template: "avg-flight-prices",
							Arguments: wfv1.Arguments{
								Artifacts: wfv1.Artifacts{
									{
										Name: "summed_flights",
										From: "{{workflow.outputs.artifacts.summed_flight_artifact}}",
									},
								},
							},
						},
						{
							Name:    "pick",
							Dependencies: []string{"avg"},
							Template: "pick-flight",
							Arguments: wfv1.Arguments{
								Artifacts: wfv1.Artifacts{
									{
										Name: "avged_flights",
										From: "{{workflow.outputs.artifacts.avged_flights_artifact}}",
									},
								},
							},
						},
					},
				},
			},
		{
			Name: "add-flight-prices",
			Inputs: wfv1.Inputs{
				Artifacts: wfv1.Artifacts{
					{
						Name: "my-art",
						Path: "my-artifact.csv",
						ArtifactLocation: wfv1.ArtifactLocation{
							S3: &wfv1.S3Artifact{
								Key: "inputs/flight_prices.csv",
								S3Bucket: wfv1.S3Bucket{
									Endpoint: "minio.minio:9000",
									Bucket: "argo",
									Insecure: &[]bool{true}[0],
									UseSDKCreds: false,
									Region: "us-east-2",
									AccessKeySecret: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "minio",
										},
										Key: "root-user" ,
									},
									SecretKeySecret: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "minio",
										},
										Key: "root-password",
									},
								},
							},
							},
						},
					},
				},
				Container: &corev1.Container{
					ImagePullPolicy: "IfNotPresent",
					Image:   "binnie268/blog.add-flight-prices:1.0.0",
					Command: []string{"python", "add-flight-prices.py"},
				},
				Outputs: wfv1.Outputs{
					Artifacts: wfv1.Artifacts{
						{
							Name: "summed_flights",
							Path: "summed_flights.csv",
							GlobalName: "summed_flight_artifact",
						},
					},
				},
			},
			{
				Name: "avg-flight-prices",
				Inputs: wfv1.Inputs{
					Artifacts: wfv1.Artifacts{
						{
							Name: "summed_flights",
							Path: "summed_flights.csv",
						},
					},
				},
				Container: &corev1.Container{
						ImagePullPolicy: "IfNotPresent",
						Image:   "binnie268/blog.avg-flight-prices:1.0.0",
						Command: []string{"python", "avg-flight-prices.py"},
				},
				Outputs: wfv1.Outputs{
					Artifacts: wfv1.Artifacts{
						{
							Name: "avged_flights",
							Path: "avged_flights.csv",
							GlobalName: "avged_flights_artifact",
						},
					},
				},
			},
			{
				Name: "pick-flight",
				Inputs: wfv1.Inputs{
					Artifacts: wfv1.Artifacts{
						{
							Name: "avged_flights",
							Path: "avged_flights.csv",
						},
					},
				},
				Container: &corev1.Container{
						ImagePullPolicy: "IfNotPresent",
						Image:   "binnie268/blog.pick-flight:1.0.0",
						Command: []string{"python", "pick-flight.py"},
				},
				Outputs: wfv1.Outputs{
					Artifacts: wfv1.Artifacts{
						{
							Name: "flight_decision",
							Path: "flight_decision.csv",
							Archive: &wfv1.ArchiveStrategy{None: &wfv1.NoneStrategy{}},
							GlobalName: "flight_decision_artifact",
							ArtifactLocation: wfv1.ArtifactLocation{
								S3: &wfv1.S3Artifact{
									Key: "outputs/flight_decision2.csv",
									S3Bucket: wfv1.S3Bucket{
										Endpoint: "minio.minio:9000",
										Bucket: "argo",
										Insecure: &[]bool{true}[0],
										Region: "us-east-2",
										UseSDKCreds: false,
										AccessKeySecret: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "minio",
											},
											Key: "root-user" ,
										},
										SecretKeySecret: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "minio",
											},
											Key: "root-password",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
func argoInit() {
	// Argo Client Setup
	config, _ := argoClient.GetConfig().ClientConfig()
	wfClient := wfclientset.NewForConfigOrDie(config).ArgoprojV1alpha1().Workflows(namespace)

	// Submit the workflow
	ctx := context.Background()
	createdWf, err := wfClient.Create(ctx, &determineFlightWorkflow, metav1.CreateOptions{})

		// Wait for the workflow to complete
		fieldSelector := fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name=%s", createdWf.Name))
		watchIf, err := wfClient.Watch(ctx, metav1.ListOptions{FieldSelector: fieldSelector.String(), TimeoutSeconds: pointer.Int64Ptr(180)})
		errors.CheckError(err)
		defer watchIf.Stop()
		for next := range watchIf.ResultChan() {
			wf, ok := next.Object.(*wfv1.Workflow)
			if !ok {
				continue
			}
			if !wf.Status.FinishedAt.IsZero() {
				fmt.Printf("Workflow %s %s at %v. Message: %s.\n", wf.Name, wf.Status.Phase, wf.Status.FinishedAt, wf.Status.Message)
				break
			}
		}
	

}
func main() {
	argoInit()
}
