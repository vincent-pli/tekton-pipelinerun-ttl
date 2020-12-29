package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	v1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektoncdclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/apis"
)

const (
	//PREFIX help to decide if the task is a job activity one
	PREFIX           = "job-activity-"
	TektonTTL string = "tekton.dev/ttl"
)

type exceptionResult struct {
	TaskrunName string
	Result      []v1beta1.TaskRunResult
}

func init() {
}

func main() {
	var namespace string
	var name string

	flag.StringVar(&namespace, "namespace", "", "The namespace of the pipelinerun.")
	flag.StringVar(&name, "name", "", "The name of the pipelinerun.")
	flag.Parse()

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("Get config of the cluster failed: %+v", err)
		os.Exit(1)
	}

	tektonClient, err := tektoncdclientset.NewForConfig(config)
	if err != nil {
		log.Errorf("Get client of tekton failed: %+v", err)
		os.Exit(1)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: TektonTTL,
	}

	pr, err := tektonClient.TektonV1beta1().PipelineRuns(namespace).List(context.TODO(), name, listOptions)
	if err != nil {
		fmt.Printf("Get Pipelinerun: %s failed: %+v", namespace+"/"+name, err)
		os.Exit(1)
	}

	fmt.Println(pr.Name)

	/*
		loop the Pipelinerun.Status.TaskRuns
		1. If the Pipelinerun.Status.TaskRuns[x].PipelineTaskName not prefix as "job-activity", ignore, that's means the task is not a Job Avtivity.
		2. Pipelinerun.Status.TaskRuns[x].Status.Conditions[0].type == Succeeded and Pipelinerun.Status.TaskRuns[x].Status.Conditions[0].status == True, means task success.
	*/
	results := []exceptionResult{}
	for _, tr := range pr.Status.TaskRuns {
		if !strings.HasPrefix(tr.PipelineTaskName, PREFIX) {
			continue
		}

		if tr.Status.GetCondition(apis.ConditionSucceeded).IsUnknown() {
			//the task failed, need write the failed reason to result
			result := exceptionResult{}
			result.TaskrunName = tr.PipelineTaskName
			result.Result = tr.Status.TaskRunResults
			results = append(results, result)
		}
	}
}

// func saveResult() error {

// }
