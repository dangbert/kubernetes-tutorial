# Kubernetes Crash Course

### Scope
We'll be running a toy Golang server in Kubernetes alongside Nginx to learn the basics.


[./backend/server.go](./backend/server.go) implements a single API endpoint. To start, go ahead and run it locally like so:

````bash
cd backend/
go run server.go

# in another terminal hit the API endpoint
curl http://localhost:8000/fruit
````

Or visit http://localhost:8000/fruit in your browser.


### Getting Started

**set up dev env**

````bash
# start a kubernetes single node "cluster" for dev purposes
# (launches inside docker)
minikube start --driver docker
````


