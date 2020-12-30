package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	v1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektoncdclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	//TektonTTL set by user express the TTL with seconds
	TektonTTL string = "tekton.dev/ttl"
)

type exceptionResult struct {
	TaskrunName string
	Result      []v1beta1.TaskRunResult
}

func init() {
}

func main() {
	log.Info("Start to clean the Pipelinerun based on the TTL.")
	var namespace string

	flag.StringVar(&namespace, "namespace", "", "The namespace of the pipelinerun.")
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

	prs, err := tektonClient.TektonV1beta1().PipelineRuns(namespace).List(context.TODO(), listOptions)
	if err != nil {
		log.Errorf("List Pipelinerun failed: %+v", err)
		os.Exit(1)
	}

	for _, pr := range prs.Items {
		ttl, err := strconv.ParseInt(pr.ObjectMeta.Labels[TektonTTL], 10, 64)
		if err != nil {
			log.Errorf("Value of ttl is not correct: %+v", err)
			continue
		}

		if pr.IsDone() || pr.IsCancelled() {
			startTime := pr.Status.CompletionTime
			existTime := time.Since(startTime.Time)

			if existTime.Seconds() > float64(ttl) {
				err = tektonClient.TektonV1beta1().PipelineRuns(namespace).Delete(context.TODO(), pr.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Errorf("Delete pr: %s failed: %+v", pr.Name, err)
					continue
				}
			}

		}

	}

	log.Info("Clean Pipelinerun complete.")
	os.Exit(0)
}
