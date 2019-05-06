#! /bin/bash
# These are the packages that are a major annoyance when patching origin,
# exclude these after patching.
git checkout -- vendor/k8s.io/kubernetes/pkg/util
git checkout -- vendor/k8s.io/kubernetes/pkg/util
git checkout -- vendor/k8s.io/kubernetes/pkg/volume/util
git checkout -- test
git checkout -- vendor/k8s.io/kubernetes/pkg/kubelet
git checkout -- vendor/k8s.io/kubernetes/cmd/kubelet/app
git checkout -- pkg/oc
git checkout -- vendor/k8s.io/kubernetes/pkg/volume
git checkout -- vendor/k8s.io/kubernetes/pkg/proxy/ipvs
git checkout -- vendor/k8s.io/kubernetes/pkg/cloudprovider/providers/cloudstack
git checkout -- vendor/k8s.io/kubernetes/pkg/kubectl
git checkout -- pkg/cmd/recycle
git checkout -- cmd/sdn-cni-plugin