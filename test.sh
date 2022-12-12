#!/usr/bin/env bash

clean_up() {
    set +e
    echo "Cleaning testing environment..."
    docker-compose down -v
}
trap clean_up EXIT

# Export variables define in .env file
export $(grep -v '^#' default.env | xargs)

# Set the kubeconfig for k3s and the runner
export KUBECONFIG="$PWD/kubeconfig.yaml"
export K3S_TOKEN=${RANDOM}${RANDOM}${RANDOM}

# Start services
docker-compose up -d

echo Waiting for Kubernetes API...
until [[ $(kubectl get endpoints/kubernetes -o=jsonpath='{.subsets[*].addresses[*].ip}' ) ]]; do sleep 2; echo -n "--- " ; done
echo Kubernetes Connection is UP!

# Run sshd on slurm frontend
docker exec slurmctld /usr/sbin/sshd

go test .
