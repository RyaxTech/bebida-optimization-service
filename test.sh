#!/usr/bin/env bash

clean_up() {
    set +e
    echo "Cleaning testing environment..."
    docker-compose down -v
}
trap clean_up EXIT

# Set the kubeconfig for k3s and the runner
export KUBECONFIG="$PWD/kubeconfig.yaml"
export K3S_TOKEN=${RANDOM}${RANDOM}${RANDOM}

# Start services
docker-compose up -d

echo Waiting for Kubernetes API...
until [[ $(kubectl get endpoints/kubernetes -o=jsonpath='{.subsets[*].addresses[*].ip}' ) ]]; do sleep 2; echo -n "--- " ; done
echo Kubernetes Connection is UP!

# Export variables for ssh access
export BEBIDA_SSH_PKEY=$(docker-compose exec -ti slurmctld cat /root/.ssh/id_rsa | base64)
export BEBIDA_SSH_USER="root"
export BEBIDA_SSH_HOSTNAME="127.0.0.1"
export BEBIDA_SSH_PORT="2222"

go test .
