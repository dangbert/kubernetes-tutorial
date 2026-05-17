# Kubernetes Crash Course

### Scope
We'll be learning the basics of Kubernetes by locally hosting a toy Go server alongside Nginx.

[./backend/server.go](./backend/server.go) implements a single API endpoint. To start, go ahead and run it locally like so:

````bash
cd backend/
go run server.go

# in another terminal hit the API endpoint
curl http://localhost:8000/fruit
````

Or visit http://localhost:8000/fruit in your browser.


### Getting Started

With that out of the way let's set up your Kubernetes dev environment.

1. Setup [Minikube](https://minikube.sigs.k8s.io/docs/start/)

````bash
# start a kubernetes single node "cluster" (inside Docker)
minikube start --driver docker

# now ensure kubectl is pointed to minikube
kubectl config get-contexts

# you should see the following:
# CURRENT   NAME       CLUSTER    AUTHINFO   NAMESPACE
# *         minikube   minikube   minikube   default
````

2. Build Docker image

````bash
docker build ./backend -t backend:latest

# now make the image available inside minikube
minikube image load backend:latest

# optionally you can test running the image locally
docker run -p 8000:8000 -t backend
````

3. Launching your first pod

Our goal is to run server.go inside a `pod`, which is a thin abstraction around a container in Kubernetes.

Using the [Kubernetes docs](https://kubernetes.io/docs/concepts/workloads/pods/) as a reference, we can describe our pod (in yaml) like so:


````yaml
apiVersion: v1
kind: Pod

metadata:
  name: backend

spec:
  containers:
    - name: backend
      image: backend:latest
      imagePullPolicy: Never # this is a local image
      ports:
      # indicate this pod shall expose port 8000
      - containerPort: 8000
````

Copy that into ./backend.yaml (in the root of this repository) then run:

````bash
# load our pod's definition into kubernetes
kubectl apply -f backend.yaml

# now you can check the status of the pod
kubetl get pods # repeat until status show "Running"

# and stream the pod's logs
kubectl logs -f pods/backend
````

The pod is currently only accessible inside the minikube host...

````bash
# so we can port forward to expose on localhost:8000
kubectl port-forward pods/backend 8000:8000

# OR alternatively get the IP of your pod (within minikube)
kubectl get pods/backend -o wide

# and curl the IP directly
minikube ssh curl http://10.XX.XX.XX9:8000/fruit

# for example
minikube ssh curl http://10.244.0.9:8000/fruit
````

4. Launching your first service

So you've managed to run the application inside Kubernetes! But what if we wanted to have a DNS record making it easier to communicate with your pod?

For this we'll define a new Kubernetes resource called a `service`. A service groups one or more pods under a single static IP address (with a DNS name), which is also helpful for load balancing traffic. [docs here if you're inter

You can optionally [see the docs here](https://kubernetes.io/docs/concepts/services-networking/service/) for more info.


Now update backend.yaml, appending to the bottom:

````yaml
# (yaml from step 3 remains here)

---
apiVersion: v1
kind: Service

metadata:
  name: backend-service
  labels:
    app: backend # we'll use this later

spec:
  selector:
    # find all pods with the label app=backend, and group under this service
    app: backend
    ports:

  ports:
    - protocol: TCP
      port: 80 # port for the service itself to expose on its IP
      targetPort: 8000 # must match the pod's `containerPort`
````


````bash
# and apply the changes
kubectl apply -f backend.yaml

# now check what IPs have been added to the service
# (the IP listed should match your pod from earlier)
kubectl get endpoings backend-service
#NAME              ENDPOINTS         AGE
#backend-service   10.244.0.9:8000   12m
````

