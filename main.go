package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "net/http/pprof"

	kuberntescli "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/elastic/elastic-agent-autodiscover/bus"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
)

type pod struct {
	publishFunc func([]bus.Event)
	watcher     kubernetes.Watcher
	client      kuberntescli.Interface
}

func main() {
	var kubeconfig string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	// Server for pprof
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Errorf("could not create client %v", err)
			return
		}
	}

	// creates the client
	client, err := kuberntescli.NewForConfig(config)
	if err != nil {
		klog.Errorf("could not create client %v", err)
		return
	}

	watcher, err := kubernetes.NewNamedWatcher("pod", client, &kubernetes.Pod{}, kubernetes.WatchOptions{
		SyncTimeout:  10 * time.Minute,
		Node:         "",
		Namespace:    "",
		HonorReSyncs: true,
	}, nil)
	if err != nil {
		klog.Errorf("could not create kubernetes watcher %v", err)
		return
	}

	p := &pod{
		watcher: watcher,
		client:  client,
	}

	watcher.AddEventHandler(p)

	klog.Infof("start watching for pods")
	go p.Start()

	// Wait forever
	select {}
}

// OnAdd ensures processing of pod objects that are newly added.
func (p *pod) OnAdd(obj interface{}) {
	o := obj.(*kubernetes.Pod)
	klog.Infof("Watcher Pod add: %+v", o.Name)

	// Get metadata of the object
	accessor, err := meta.Accessor(o)
	if err != nil {
		return
	}
	meta := map[string]string{}
	for _, ref := range accessor.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			switch ref.Kind {
			// grow this list as we keep adding more `state_*` metricsets
			case "Deployment",
				"ReplicaSet",
				"StatefulSet",
				"DaemonSet",
				"Job",
				"CronJob":
				meta[strings.ToLower(ref.Kind)+".name"] = ref.Name
			}
		}
	}
	if jobName, ok := meta["job.name"]; ok {
		dep := p.getCronjobOfJob(jobName, o.GetNamespace())
		if dep != "" {
			meta["cronjob.name"] = dep
		}
		klog.Infof("Watcher Pod of Job of Cronjob: %+v", meta)
	}
}

// getCronjobOfJob return the name of the Cronjob object that
// owns the Job with the given name under the given Namespace
func (p *pod) getCronjobOfJob(jobName string, ns string) string {
	if p.client == nil {
		return ""
	}
	cronjob, err := p.client.BatchV1().Jobs(ns).Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	for _, ref := range cronjob.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			switch ref.Kind {
			case "CronJob":
				return ref.Name
			}
		}
	}
	return ""
}

// OnUpdate handles events for pods that have been updated.
func (p *pod) OnUpdate(obj interface{}) {
	o := obj.(*kubernetes.Pod)
	klog.Infof("Watcher Pod update: %+v", o.Name)

	// Get metadata of the object
	accessor, err := meta.Accessor(o)
	if err != nil {
		return
	}
	meta := map[string]string{}
	for _, ref := range accessor.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			switch ref.Kind {
			// grow this list as we keep adding more `state_*` metricsets
			case "Deployment",
				"ReplicaSet",
				"StatefulSet",
				"DaemonSet",
				"Job",
				"CronJob":
				meta[strings.ToLower(ref.Kind)+".name"] = ref.Name
			}
		}
	}
	if jobName, ok := meta["job.name"]; ok {
		dep := p.getCronjobOfJob(jobName, o.GetNamespace())
		if dep != "" {
			meta["cronjob.name"] = dep
		}
		klog.Infof("Watcher Pod of Job of Cronjob: %+v", meta)
	}
}

// OnDelete stops pod objects that are deleted.
func (p *pod) OnDelete(obj interface{}) {
	o := obj.(*kubernetes.Pod)
	klog.Infof("Watcher Pod delete: %+v", o.Name)

	// Get metadata of the object
	accessor, err := meta.Accessor(o)
	if err != nil {
		return
	}
	meta := map[string]string{}
	for _, ref := range accessor.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			switch ref.Kind {
			// grow this list as we keep adding more `state_*` metricsets
			case "Deployment",
				"ReplicaSet",
				"StatefulSet",
				"DaemonSet",
				"Job",
				"CronJob":
				meta[strings.ToLower(ref.Kind)+".name"] = ref.Name
			}
		}
	}
	if jobName, ok := meta["job.name"]; ok {
		dep := p.getCronjobOfJob(jobName, o.GetNamespace())
		if dep != "" {
			meta["cronjob.name"] = dep
		}
		klog.Infof("Watcher Pod of Job of Cronjob: %+v", meta)
	}
}

// Start starts the eventer
func (p *pod) Start() error {
	return p.watcher.Start()
}

// Stop stops the eventer
func (p *pod) Stop() {
	p.watcher.Stop()
}
