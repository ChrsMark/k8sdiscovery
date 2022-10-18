package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

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
}

// OnUpdate handles events for pods that have been updated.
func (p *pod) OnUpdate(obj interface{}) {
	o := obj.(*kubernetes.Pod)
	klog.Infof("Watcher Pod update: %+v", o.Name)
}

// OnDelete stops pod objects that are deleted.
func (p *pod) OnDelete(obj interface{}) {
	o := obj.(*kubernetes.Pod)
	klog.Infof("Watcher Pod delete: %+v", o.Name)
}

// Start starts the eventer
func (p *pod) Start() error {
	return p.watcher.Start()
}

// Stop stops the eventer
func (p *pod) Stop() {
	p.watcher.Stop()
}
