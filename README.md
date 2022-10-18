# K8s Pod discovery

This small project provides a Pod Autodiscovery example based on
`github.com/elastic/elastic-agent-autodiscover` library. 


## Building
```
docker build -t k8sdiscovery .
docker tag k8sdiscovery:latest chrismark/k8sdiscovery:v0.0.1
```

## Uploading to a registry
```
docker tag k8sdiscovery:latest chrismark/k8sdiscovery:v0.0.1
docker push chrismark/k8sdiscovery:v0.0.1
```

## Load in a kind cluster to use it locally
```
kind load docker-image chrismark/k8sdiscovery:v0.0.1
```
And uncomment the `imagePullPolicy: Never` inside the manifest


## Run on Kubernetes

```
kubectl apply -f k8sdiscovery-kubernetes.yml
kubectl -n kube-system port-forward pod/k8sdiscovery 6060:6060
```

## Profiling the program

After having deployed everything on the cluster, run the following command to
take the heap profile of the program:
1`go tool pprof -png http://localhost:6060/debug/pprof/heap > out.png`