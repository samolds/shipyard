# Democart

This is an exercise in building a modular and containerized Shopping Cart
service. It consists of 3 main components:

1. A backend REST API, written in Go.
2. An identity provider, written in Go.
3. A frontend marketplace, using ReactJS.

The backend REST API and the identity provider are both running on the same
host for simplicity, but they could easily be broken out. They are both located
in `go/src/democart`.  
The frontend marketplace is located in `web/democart`.


### Dependencies

- docker


### Building From Scratch and Running Locally

```sh
make new
```


### Running Locally

```sh
make up
```


### Start a Quick'n'Dirty Dev Server

```sh
make devup
```


### How to Use

- Go to [localhost:3000](http://localhost:3000) in a browser.
- "Sign Up" to create a new user
- "Make Dummy Address"
- "Make Dummy Item"
- See the item in the marketplace
- Add the item to your cart
- Add the item to your cart again
- Order the items in your cart
- Logout
- See the Dummy items in the marketplace


### TODOs

- More unit tests
- Better documentation
- More graceful error handling
- Stylize the frontend better
- Production-ify with nginx, letsencrypt, K8s, and managed RDS db

#### Now:
1. Use main.go style from demoapi in democart
2. Better configuration handling between psql secrets and server config file
   using config from demoapi
3. Finish off basic K8s stuff with working example. commit branch. merge. Start
   go backend cleanup
4. Add demoapi grafana/prometheus server
5. Use Dockerfile template from demoapi in democart
6. Break out idp to separate server entirely.
7. Essentially, clean up democart to use some of the more elogant demoapi stuff,
   but with all the functionality of democart
8. Rename Democart to API Server Stencil / Server, etc
  - Something like:
    - API Starter
    - Template
    - K8s Server Template
    - Prod Server
    - Server Build
    - Server Foundation
    - API Foundation
    - Prod API Base
    - Sam Base
    - API Base
    - Base Layer
    - API Thermal Layer
9. K8s configuration stuff:
  - Add namespacing metadata
  - Add rollingupdate strategy
  - Add better labels/selectors - app, tier, env, release, etc
10. Clean up fake idp stuff. See if swapping out for Auth0 works.
11. Use demoapi's `entrypoint.sh` instead of `wait_for_psql.sh`
12. Clean up docker-compose stuff so that works in addition to k8s stuff
13. Stop using DBX
14. Figure out "failed to sync configmap cache: timed out waiting for the
    condition" occasional error
15. Better pod DNS:
    https://kubernetes.io/docs/concepts/services-networking/dns-pod-service


### Notes

- docker-compose.yml is useful to build the project during development
- K8s is a more production ready solution.
  - K8s components of this project:
    - A deployment for the backend api
    - A deployment for the frontend app
    - A deployment for the database (normally, this is outside the cluster)
    - A service exposing the backend api publicly (load balanced) (NodePort)
    - A service exposing the frontend publicly (load balanced) (NodePort)
    - A service exposing the database to only the backend nodes (ClusterIP)
    - A persistent volume to persist database data
    - A horizontal pod autoscaler for the backend
    - A horizontal pod autoscaler for the frontend
    - An Ingress routing frontend and backend traffic??
    - A config map for backend
    - A config map for frontend


### Useful Links/Ideas/Notes:
- "Connecting a frontend to a backend"
  - https://kubernetes.io/docs/tasks/access-application-cluster/connecting-frontend-backend
  - It looks like it's using nginx within the frontend pods to find the backend service
- Kompose
  - https://github.com/kubernetes/kompose/blob/master/README.md#installation
  - A command line tool to convert docker-compose yaml to kubernetes templates
- A "How To" to convert a docker-compose to prod ready k8s
  - https://www.digitalocean.com/community/tutorials/how-to-migrate-a-docker-compose-workflow-to-kubernetes
- Does democart need to have hosted docker images?? or can they be local??
  - https://hub.docker.com/repository/docker/samolds/democart
- How to use local docker images with a minikube cluster
  - https://medium.com/bb-tutorials-and-thoughts/how-to-use-own-local-doker-images-with-minikube-2c1ed0b0968
- Connecting to a localhost-running db or service from docker
  - https://docs.docker.com/docker-for-mac/networking/#use-cases-and-workarounds
  - https://stackoverflow.com/q/49289009
- Mapping to external services, like communicating with an external db
  - https://cloud.google.com/blog/products/gcp/kubernetes-best-practices-mapping-external-services
