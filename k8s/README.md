# Steps to Deploy to a Minikube K8s Cluster

1. Build and publish new Docker images

```sh
make build
make push-docker-images
```

2. Start Minikube with an ingress and hpa enabled

```sh
minikube start --vm=true --addons metrics-server
minikube addons enable ingress
```

3. Add hostnames to host machine's /etc/hosts as well as minikube's because I
   haven't figured out proper DNS on localhost with minikube.

```sh
MKIP=$(minikube ip)
echo "$MKIP shipyard.tech" | sudo tee -a /etc/hosts
echo "$MKIP api.shipyard.tech" | sudo tee -a /etc/hosts
echo "$MKIP idp.shipyard.tech" | sudo tee -a /etc/hosts
echo "$MKIP prom.shipyard.tech" | sudo tee -a /etc/hosts
echo "$MKIP grafana.shipyard.tech" | sudo tee -a /etc/hosts

echo "$MKIP api.shipyard.tech" | minikube ssh -- sudo tee -a /etc/hosts
echo "$MKIP idp.shipyard.tech" | minikube ssh -- sudo tee -a /etc/hosts
echo "$MKIP prom.shipyard.tech" | minikube ssh -- sudo tee -a /etc/hosts
```

4. Test apply the K8s configurations

```sh
kubectl apply -f k8s/dev --dry-run --validate=true
```

5. If everything looks good, actually apply configs

```sh
cd k8s/dev/scripts && ./apply_dev && cd ../../..
```

6. Navigate to `http://shipyard.tech` in a browser

7. Delete cluster and all persistent volumes

```sh
# persitent volumes (pv) aren't namespace scoped. delete manually
kubectl delete namespaces dev
kubectl delete pv shipyard-db-persistentvolume
kubectl delete pv shipyard-grafana-persistentvolume
```

8. Delete statefully persistent db data on host node

```sh
minikube ssh -- sudo rm -rf /data/shipyard-db-data /data/shipyard-grafana-data
```

9. Cleanup Hosts

```sh
vi /etc/hosts
# remove shipyard.tech references
minikube ssh
vi /etc/hosts
# remove shipyard.tech references
```


### Grafana

To hook up Grafana to the Prometheus Server, add a new prometheus
datasource in the Grafana dashboard (at http://grafana.shipyard.tech). The first time
you sign in, the credentials are admin:admin.

Add http://prom.shipyard.tech as the URL, and leave access as "Server". Then scroll down and click "Save and Test".


### Other Useful Commands

Get shell access to the pods

```sh
kubectl --namespace=dev exec --stdin --tty <full_pod_name> -- /bin/bash
```




### Helpful Links
- https://kubernetes.io/docs/reference/kubectl/cheatsheet
