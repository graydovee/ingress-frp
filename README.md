# ingress-frp

[![GitHub license](https://img.shields.io/github/license/kubernetes/ingress-nginx.svg)](https://github.com/graydovee/ingress-frp/blob/main/LICENSE)

## Description

Ingress-Frp is a Ingress controller for Kubernetes. It use [Frp](https://github.com/fatedier/frp) to help you to visit your private network's Kubernetes services from the Internet.

## Getting Started

### Install

#### install with helm

```sh
helm upgrade ingress-frp ./deploy/ingress-frp -n kube-system \
--set frp.token=<SAME_TO_YOUER_SERVER_TOKEN> \
--set frp.frpc.nodeSelector."ingress\.graydove\.cn/frp"=frpc \ 
--set frp.frps.addr=<YOUR_FRP_SERVER_ADDRESS> \
--set frp.frps.port=<YOUR_FRP_SERVER_PORT> \
--set manager.image.tag=v0.0.2
```

#### install with Kustomize

1. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/ingress-frp:tag
```

2. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/ingress-frp:tag
```

##### UnDeploy the controller to the cluster:

```sh
make undeploy
```

## How to use

Should set ingress.spec.ingressClassName=frp to enable fro ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myproject-ingress
  namespace: default
spec:
  ingressClassName: frp
  tls:
  - hosts:
    - myproject.example.com
    secretName: myproject-tls
  rules:
  - host: myproject.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: myproject
            port:
              number: 3000
---
apiVersion: v1
kind: Service
metadata:
  name: myproject
  namespace: default
spec:
  selector:
    app.kubernetes.io/name: myproject
  type: ClusterIP
  clusterIP: 127.0.0.1
  ports:
  - port: 3000
    targetPort: 3000
    protocol: TCP
    name: http
```

## Contributing

Welcome to contribute

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) and  [Frp](https://github.com/fatedier/frp)
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster

https://book.kubebuilder.io/introduction.html)

