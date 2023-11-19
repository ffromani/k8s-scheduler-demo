kind create cluster --config=hack/kind-config.yaml
kubectl label node kind-worker node-role.kubernetes.io/worker=''
kind load docker-image ${IMAGE}
