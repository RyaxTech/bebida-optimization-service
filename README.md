# BeBiDa Optimization Service

This repository provides the BeBiDa optimizations to improve Big Data jobs turnaround time through adapted mechanisms which offer better guarantees. Our goal is to bring more elasticity in HPC platforms, in order to allow Big Data jobs (Spark) to be executed dynamically, but without altering its resource manager’s internal aspects or losing in scheduling efficiency. We are convinced that each scheduling mode (HPC and Big Data) have their own advantages and disadvantages and they fit better to serve the needs of their typical use cases hence we do not want to change internals of any of them. For that, we focus in extending the BeBiDa techniques that enable the HPC and Big Data resource and job management systems to collocate with minimal interference on the HPC side, with **acceptable and high guarantees** for the Big Data jobs executions.

We use two mechanisms to improve BeBiDa guarantees: 1) deadline-aware and 2) time-critical. These two approaches are complementary could be combined.

## Overview 

The following figure sketches the design of executing jobs using the new BeBiDa deadline-aware and time-critical techniques through the usage of RYAX workflow engine.

<!---
![BeBiDa optimizations 1{caption=High-level view of the deadline-aware and time-critical BeBiDa mechanisms.}](./figureBOS.png?raw=true)
-->

<figure>
  <img
  src="./figureBOS.png">
  <figcaption>High-level view of the deadline-aware and time-critical BeBiDa mechanisms.</figcaption>
</figure>


## Deadline-aware (Punch)

In this technique we create empty jobs, called Punch jobs, which do not trigger the prolog/epilog to leave room for applications. Hence we prepare holes on the HPC schedule plan to guarantee a fixed pool of resources for the Big Data workload. The main issue is when to trigger these jobs and with how many resources and time. 

To cope with this problem, user can provide information on job duration and resource needs using Kubernetes annotations on the application Pod. See the configuration section below for more details.

If no deadline is provided the Punch job with be submitted immediately. 
If no resources hints are provided the default value will be used.

## Time-critical

In this technique we will use a dynamic set of resources to serve applications immediately and scale them out and in (grow and shrink) when necessary. Again, the main issue is when to add or remove nodes from the on-demand pool. For this we will make use of advanced reservations.

## Spark Application use Bebida Operator

A Spark application on Kubernetes is deployed with a Driver that coordinate the application and get resources from Kubernetes by creating pods. So, a Spark application is a set of pods with a static driver and a set of dynamic executors. To properly handle this with Bebida, we introduce the notion of Bebida Operator. the Optimizer get the annotation from the Driver and ignore the executors pod in the optimization.
The be excluded from the Bebida default Punch mechanism the Spark driver need to have the following annotation: `ryax.tech/bebidaOperator`


## Usage

## Roadmap

- [X] First implementation with simple heuristic (Ti'Punch):
    base on a threshold on the number of pending job in the BDA queue, create
    HPC jobs that will stay in the BDA resource pool.
- [X] support for K8s (BDA)
- [X] support Slurm over SSH (HPC)
- [X] Handle BDA app early termination (cancel HPC job if not used anymore)
- [X] Full testing environment with SLURM and OAR [nixos-compose compositions](https://github.com/oar-team/regale-nixos-compose/tree/main/bebida)
- [X] Support for OAR over SSH (HPC)
- [ ] Add deadline support using Kubernetes annotations
- [ ] Implement the TimeCritical app support with dynamic partitioning
  - [ ] OAR Quotas
  - [ ] Slurm partition Limits
- [ ] Improve heuristic using BDA app time and resource requirements

# Usage

## Configure

BeBiDa uses annotation to gather information about job types and resources requirements. Annotations for BeBiDa are:

* `ryax.tech/timeCritical`: set to `true` to prioritize the job as time critical (defaults to: `false`)
* `ryax.tech/deadline`: date of the deadline in the RFC3339 format, e.g. "2006-01-02T15:04:05Z07:00"
* `ryax.tech/duration`: walltime in seconds​
* `ryax.tech/resources.cores`: number of cores needed
* `ryax.tech/resources.memory`​: memory in megabytes

## Setup a testing environment

**WARNING**: This environment is for testing only. It is not secure because
secrets are hard-coded to simplify development.

You can use either VMs for local development (be aware that you will need at least 16GB of memory), or the grid5000 for larger deployment.

### On you local machine with VMs

You can spawn a test cluster for Bebida using nixos-compose. First, you will need to install [Nix](https://github.com/DeterminateSystems/nix-installer) on your machine. Now you can get the nixos-compose and the needed derivation with:
```sh
nix develop "github:oar-team/regale-nixos-compose?dir=bebida#devShell.x86_64-linux"
```

In the provide shell run  (for slurm just replace oar by slurm in the next command):
```sh
nxc build -C oar::vm
export MEM=2048  # Needed to set the VM memory size
nxc start
```

In order to expose OAR the services your machine run:
```sh
ssh -f -N -L 8080:localhost:80 root@localhost -p 22022
ssh -f -N -L 8081:localhost:80 root@localhost -p 22023
```

### On Grid5000

[Grid5000](https://www.grid5000.fr) is research testbed that allows you to simply deploy the testing environment on bare metal servers.

To do so, connect to the frontend of a site and install NixOS-compose (a.k.a `nxc`) to be able to run the composition build:
```sh
pip install git+https://github.com/oar-team/nixos-compose.git

# You might want to add this on your .bashrc
cat >> ~/.bashrc <<EOF
export PATH=$PATH:$HOME/.local/bin
EOF

# Make the exeutable available
source ~/.bashrc

nxc helper install-nix

# Add some nix configuration
mkdir -p ~/.config/nix
cat > ~/.config/nix/nix.conf <<EOF
experimental-features = nix-command flakes
EOF

nix --version
```

> NOTE:
> Because building the environment might use a lot of resources it is advised to run this build inside an interactive job using:
> `oarsub -I`

Now that you have Nix installed, and the `nxc` available and you can build environment:
```sh
git clone https://github.com/oar-team/regale-nixos-compose.git
cd regale-nixos-compose/bebida/
nxc build -C oar::g5k-nfs-store
```

Finally, you can reserve some resources and deploy:
```sh
# Get some resource and capture the Job ID
export $(oarsub -l cluster=1/nodes=4,walltime=1:0:0 "$(nxc helper g5k_script) 1h" | grep OAR_JOB_ID)
# WAit until the job starts...
oarstat -j $OAR_JOB_ID -J | jq --raw-output 'to_entries | .[0].value.assigned_network_address | .[]' > machines
nxc start -C oar::g5k-nfs-store -m machines
```

In order to expose OAR services and Ryax on your machine, you can create a port forward. Run this on your machine but replace the nodes and the site. To the node IP by name, for example for server use `nxc helper ip server`:
```sh
ssh -f -N <SITE>.g5k -L 8080:<FRONTEND MACHINE>:80
ssh -f -N <SITE>.g5k -L 8081:<SERVER MACHINE>:80
```

## Tests

Run the integrated tests with:
```sh
nxc driver -t
```

You can access the OAR API, Drawgantt and Monika interface in your browser with: [http://localhost:8080/drawgantt]()

The Ryax interface is available at: [http://localhost:8081/app]()


You can now check that the cluster is running by watching the nodes state with:
```sh
nxc connect server
kubectl get nodes
```
You should have 3 nodes in Ready state.

In another terminal, check that all Pods are Running and watch them with:
```sh
nxc connect server
kubectl get pods -A -w -o wide
```

In a third terminal, you can see the idle nodes in the OAR/Slurm cluster and run a job to trigger the BeBiDa mechanism.
On OAR:
```sh
nxc connect frontend
su - user1
oarstat
oarsub -l nodes=2 sleep 10
```

Or on Slurm:
```sh
nxc connect frontend
sinfo
srun -N 2 sleep 10
```

You can see that pods previously on the `node1` and `node2` nodes where removed before the job starts and the nodes were in SchedulingDisabled during the job and then come back to a Ready state.


## Test the Bebida Optimizer

The Bebida optimization process called `bebida-shaker` is available in the test
environment and runs on the `server` node. You can watch the logs using:
```sh
journalctl -u bebida-shaker -f
```

To see it in action, create pods with:
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

## Full Scenario Demo

In order to demonstrate how the Bebida optimization works, here is a complete scenario that showcase the deadline aware feature of Bebida Shaker.

In the testing environment, we will play the following scenario:

1. The HPC cluster runs some regular HPC jobs and some other are in the queue
2. A Spark application with a deadline is submitted through Ryax to the Kubernetes cluster, Bebida Shaker creates a Punch job before the deadline using annotations to get resources needs
3. The application starts with the driver scheduled in the Kubernetes only safe-node and an executor starting on available HPC nodes
4. An HPC job finishes which leave room for some pending executors starts
5. An HPC job starts which kills some running executors
6. The Punch job starts, and is used to deploy pending executors
7. The Spark application finishes and the results are available in Ryax

TODO: Add a screen capture here!
