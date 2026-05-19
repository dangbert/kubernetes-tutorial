# Kubernetes Crash Course

**Scope:**
We'll be learning the basics of Kubernetes by locally hosting a toy Go server.

[./backend/server.go](./backend/server.go) implements a few API endpoints. To start, go ahead and run it locally like so:

````bash
cd backend/
go run server.go

# in another terminal hit the API endpoint
curl http://localhost:8000/fruit
````

Or visit http://localhost:8000/fruit in your browser.


## Part 1: Kubernetes Basics


### 1. Setup [Minikube](https://minikube.sigs.k8s.io/docs/start/)

Minikube provides a pre-configured host ready to run kubernetes.

````bash
# start a kubernetes single node "cluster" (inside Docker)
minikube start --driver docker

# now ensure kubectl is pointed to minikube
kubectl config get-contexts

# you should see the following:
# CURRENT   NAME       CLUSTER    AUTHINFO   NAMESPACE
# *         minikube   minikube   minikube   default
````

### 2. Build Docker image

````bash
docker build ./backend -t backend:latest

# now make the image available inside minikube
minikube image load backend:latest

# optionally you can test running the image locally
docker run -p 8000:8000 -t backend
````

### 3. Launching your first pod

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

Note if you're having any trouble, you can see the solution for this step [here](https://github.com/dangbert/k8s-tutorial/blob/solution/backend5.yaml).

### 4. Launching your first service

So you've managed to run the application inside Kubernetes! But what if we wanted to have a DNS record making it easier to communicate with your pod?

For this we'll define a new Kubernetes resource called a `service`. A service groups one or more pods under a single static IP address (with a DNS name), which is also helpful for load balancing traffic.

* You can optionally [see the docs here](https://kubernetes.io/docs/concepts/services-networking/service/) for more info.


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
# and now apply the changes
kubectl apply -f backend.yaml

# now check what IPs have been added to the service
# (the IP listed should match your pod from earlier)
kubectl get endpoints backend-service
#NAME              ENDPOINTS         AGE
#backend-service   10.244.0.9:8000   12m
````

And now let's try to hit the service under the DNS name "backend-service". To do this we'll launch a temporary pod named "debug", containing tools like the curl command.

````
kubectl run debug --image=nicolaka/netshoot -it --rm --restart=Never -- sh

# ^this may take a while to pull the image, but you can always monitor the status in a new terminal:
watch -n 1 kubectl describe pods debug

# once the "debug" pod is launched run inside:
curl http://backend-service/fruit
# then use <ctrl>d to kill the pod
````

Note if you're having any trouble, you can see the solution for this step [here](https://github.com/dangbert/k8s-tutorial/blob/solution/backend4.yaml).

### 5. Launching your first deployment

Now what if you wanted to scale to 3 instances of the backend pod? For this we'll use the `deployment` resource. A deployment handles spawning pods (from a template), ensuring the desired "replica" count (e.g. 3) is always met.
* See the [deployment docs](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) for more info.


Modify backend.yaml again, this time DELETE the entire `kind: pod` section, then add this in its place:

````yaml
apiVersion: apps/v1
kind: Deployment

metadata:
  name: backend
  labels:
    app: backend

spec:
  # indicate we'd prefer 3 instances of the pod
  replicas: 3
  selector:
    matchLabels:
      app: backend

  # here we describe how our pods should be created
  template:
    # this part is identical to the pod we defined in step 3:
    metadata:
      name: backend

      labels:
        app: backend

    spec:
      containers:
        - name: backend
          image: backend:latest
          imagePullPolicy: Never # this is a local image
          ports:
          # indicate this pod shall expose port 8000
          - containerPort: 8000

# (definition of our service from step 4 continues below)
---
# ...
````

````bash
# let's check on our existing pod
kubectl get pod

# and delete it
kubectl delete pod backend

# now apply, creating the "deployment" resource
kubectl apply -f backend.yaml
````

Our deployment should immediately create three backend pods. Let's check:

````bash
kubectl get pods
# NAME                      READY   STATUS    RESTARTS   AGE
# backend-5b664f9c6-284bv   1/1     Running   0          26s
# backend-5b664f9c6-fh856   1/1     Running   0          26s
# backend-5b664f9c6-xc7bk   1/1     Running   0          26s

# and we can stream the logs from all our backend pods at once
# (note that we can filter by the label app=backend here)
kubectl logs -l app=backend --prefix=true -f

# now in another terminal try to hit the backend-service
kubectl run debug --image=nicolaka/netshoot -it --rm --restart=Never -- curl http://backend-service/fruit
````

Observe the logs as you hit the backend, and try hitting it a few times. You'll notice that a different pod may respond to the request each time.

Note if you're having any trouble, you can see the solution for this step [here](https://github.com/dangbert/k8s-tutorial/blob/solution/backend5.yaml).

## Part 2: Helm Basics

Consider the yaml configs (a.k.a. "manifests") we created in Part 1. How would we distribute them to the world if we wanted to let anyone run our backend in the same way? While we could just pass around the backend.yaml file, if our system was more complex with many microservices across many yaml files this could become less portable. Also what if we want to deploy our manifests slightly differently in our staging cluster vs production?


Enter Helm. Helm is a package manager for Kubernetes manifests, allowing pre-packaged resources called `Charts` (i.e. a set of deployments, services, etc) to be shared and quickly installed in any kubernetes cluster. It also let's you use templating to dynamically define your manifests, giving you levers to customize the behavior in different environents.

In this section we'll learn the basics of Helm by refactoring our work above into a simple helm chart.

* [Explore the Helm docs](https://helm.sh/docs/intro/using_helm) for more information.


### 1. Create Helm Chart

````bash
helm create demo
````

Now you'll have the demo folder:
````txt
demo/
├── Chart.yaml
├── charts
├── templates
│   ├── NOTES.txt
│   ├── _helpers.tpl
│   ├── deployment.yaml
│   ├── hpa.yaml
│   ├── httproute.yaml
│   ├── ingress.yaml
│   ├── service.yaml
│   ├── serviceaccount.yaml
│   └── tests
│       └── test-connection.yaml
````

But let's refactor:
````bash
rm -rf demo/templates/*

# and move our backend.yaml into templates
mv backend.yaml demo/templates/
````

Now you're folder structure will look like this
````txt
$ tree demo/
demo/
├── Chart.yaml
├── charts
├── templates
│   └── backend.yaml
└── values.yaml
````

### 2. Install Helm Chart

Now let's install our Helm "chart" inside Kubernetes.

````bash
# start by deleting your existing resources
kubectl delete deployment/backend service/backend-service

# and verifying they're gone
kubectl get all
# you should only see:
#NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
#service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   8m59s

# and let's install your helm chart
helm install demo ./demo/

# and verify your deployment and service were recreated:
kubectl get all
````

You've successfully refactored your Kubernetes manifests into a Helm chart!

### 3. Try Helm Templating

Now we'll do a small refactor to try to take advantage of Helm templating. In this case we'll make the port exposed by the backend configurable via a boolean in [./demo/values.yaml](./demo/values.yaml) file.


