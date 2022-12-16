# BeBiDa Optimization Service

This repository provides the BeBiDa optimizations to improve Big Data jobs turnaround time through adapted mechanisms which offer better guarantees. Our goal is to bring more elasticity in HPC platforms, in order to allow Big Data jobs (Spark) to be executed dynamically, but without altering its resource manager’s internal aspects or losing in scheduling efficiency. We are convinced that each scheduling mode (HPC and Big Data) have their own advantages and disadvantages and they fit better to serve the needs of their typical use cases hence we do not want to change internals of any of them. For that, we focus in extending the BeBiDa techniques that enable the HPC and Big Data resource and job management systems to collocate with minimal interference on the HPC side, with **acceptable and high guarantees** for the Big Data jobs executions.

We can use two mechanisms to improve BeBiDa guarantees: 1) deadline-aware and 2) time-critical. These two approaches are complementary will be combined.

## Deadline-aware
In this technique we create empty jobs which do not trigger the prolog/epilog to leave room for applications. Hence we prepare holes on the HPC schedule plan to guarantee a fixed pool of resources for the Big Data workload. The main issue is when to trigger these jobs and with how many resources and time.

## Time-critical
In this technique we will use a dynamic set of resources to serve applications immediately and scale them out and in (grow and shrink) when necessary. Again, the main issue is when to add or remove nodes from the on-demande pool. For this we will make use of advanced reservations.

The following figure sketches the design of executing jobs using the new BeBiDa deadline-aware and time-critical techniques through the usage of RYAX workflow engine.

<!---
![BeBiDa optimizations 1{caption=High-level view of the deadline-aware and time-critical BeBiDa mechanisms.}](./figureBOS.png?raw=true)
-->

<figure>
  <img
  src="./figureBOS.png">
  <figcaption>High-level view of the deadline-aware and time-critical BeBiDa mechanisms.</figcaption>
</figure>

## Roadmap

- [X] First implementation with simple heuristic (Ti'Punch):
    base on a threshold on the number of pending job in the BDA queue, create
    HPC jobs that will stay in the BDA resource pool.
- [X] support for K8s (BDA)
- [X] support Slurm over SSH (HPC)
- [X] Full testing environment
- [X] Handle BDA app early termination (cancel HPC job if not used anymore)
- [ ] Handle HPC job termination
- [ ] Support for OAR (HPC)
- [ ] Improve heuristic using BDA app time and resource requirements

# Usage

## Build

```sh
go build .
```

## Testing environment

**WARNING**: This environment is for testing only. It is not secure because
secrets are hard-coded to simplify development.

You can spawn a test cluster for Bebida using:
```sh
docker-compose up -d
```
It contains a Slurm master and a K3s master with 1 Kubernetes only worker node and 2 nodes Slurm+K3s nodes with Bebida enabeled.

You can now test check that the cluster is running by watching the nodes state
with:
```sh
docker-compose exec -ti k3s-server kubectl get nodes -w
```
You should have 4 nodes in Ready state.

In another terminal, check that all Pods are Running and watch them with:
```sh
docker-compose exec -ti k3s-server kubectl get pods -A -w -o wide
```

In a third terminal, you can see the available 2 nodes in the Slurm cluster in idle and then run a job with:
```sh
docker-compose exec -ti slurmctld sinfo
docker-compose exec -ti slurmctld srun -N 2 sleep 10
```

You can see that pods previously on the `c1` and `c2` nodes where removed
before the job starts and the nodes were in SchedulingDisabled during the job
and then come back to a Ready state.

## Run the Bebida Optimizer

Before running the optimizer you'll need SSH access to the Slurm frontend. In
the project root directory run:
```sh
export BEBIDA_SSH_PKEY=$(docker-compose exec -ti slurmctld cat /root/.ssh/id_rsa | base64)
export BEBIDA_SSH_USER="root"
export BEBIDA_SSH_HOSTNAME="127.0.0.1"
export BEBIDA_SSH_PORT="2222"
```

Get Kubernetes access from the k3s generated config:
```sh
export KUBECONFIG=$PWD/kubeconfig.yaml
```

Then, run the optimizer:
```sh
go run .
```

Put a job in the queue with:
```sh
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: busybox-1
spec:
  containers:
  - name: busybox
    image: busybox:1.28
    args:
    - sleep
    - "100"
---
apiVersion: v1
kind: Pod
metadata:
  name: busybox-2
spec:
  containers:
  - name: busybox
    image: busybox:1.28
    args:
    - sleep
    - "1000"
EOF
```

