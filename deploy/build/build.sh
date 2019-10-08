#!/usr/bin/env bash

cd ${GOPATH}/src/k8s-billing
go build -o deploy/build/billing cmd/main.go || exit 0

cd ${GOPATH}/src/k8s-billing/deploy/build
sudo docker build -t k8s-billing:v1.0.1 .

cd ${GOPATH}/src/k8s-billing/deploy/deployment
kubectl delete po k8s-billing-0 -nkube-system
kubectl apply -f deployment.yaml

sleep 3
kubectl logs -f --tail=200 k8s-billing-0 -nkube-system