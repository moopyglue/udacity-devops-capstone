# udacity-devops-capstone

## overview

Jenkins based pipeline with ROLLING deployment model for an application called 'eyes' with ...

- a jenkins pipeline that published changes triggered by changes on a github repo
- pipeline includes linting
- a simplistic cloudfoundry build which lays out the needed AWS structure to run things
- a kubernetes cluster to run 'carrier' nodes
- a stand alone 'hub' node to manage shared information bween 'carrier' nodes.

## application
The large scale version of 'eyes' used here runs multiple 'carrier' instances in a kubernettes cluster with a central backend 'hub' which acts like a memory database. The carrier docker containers can handle approximatly 200 simultainoulsy connected (websocket connections) end points each, and the central hub should be able to manage the shared data for up to 20 carriers.

A better design would have been to use a lamba function to pair up controller and listener websocket conections on the same hosts so that no central host is required to share data, but no time to redesign in during project.

## Background non-documented setup

- setting up ssh keys as needed in AWS
- runnig a jenkins server (which I did on my own home server) and configuring it to use a pipeline file from the repo and to be abe to recieve secure API inputs from git hub
- setting up secure API trigered events on commitment of changes to github to reach out to jenkins
- setting up credentials for github, aws, and docker.io 
- early testing was done using minikube
- installing golang

## cloudfoundry setup
The 'base' file creates the free network structure and the kubernetes control cluster (not free, but cheap). 
The 'hub' creates a single hub node, and the 'nodes' file creates the nodes to be used for the cluster - these were seperated from above so they can be comissioned/decomissioned as needed to educe costs during development.
```
$ cd cloudfoundry
$ ./stack create udcap-base
$ ./stack create udcap-hub
$ ./stack create udcap-nodes
```

## initial app docker container build
To get things running you will need to generate the first docker container so that kubernetetes cluster can be setup
```
$ cd eyes-server
$ make all upload
```

## kubenetes cluster setup
First creation is done using below command. This sets up the 
```
$ kubectl apply -f kubernetes/k8-config.yaml
```

