set -x
cd $GOPATH/src/github.com/openshift/
rm -rf origin
git clone https://github.com/openshift/origin --branch release-4.1
cd $PATCHDIR
patchsource --source-dir=$GOPATH/src/github.com/openshift/origin --code-config-yaml=patch.yaml
cd $GOPATH/src/github.com/openshift/origin
cp -rf $GOPATH/src/github.com/maruel/ $GOPATH/src/github.com/openshift/origin/_output/local/go/src/github.com/openshift/origin/vendor/github.com/
hack/build-go.sh cmd/hypershift
cp -rf $GOPATH/src/github.com/maruel/ $GOPATH/src/github.com/openshift/origin/_output/local/go/src/github.com/openshift/origin/vendor/github.com/
hack/build-go.sh cmd/hypershift
sed -i 's/hyperkube controller-manager/hyperkube kube-controller-manager/g' $GOPATH/src/github.com/openshift/origin/hack/local-up-master/lib.sh
hack/install-etcd.sh
export PATH=$GOPATH/src/github.com/openshift/origin/_output/tools/etcd/bin:$PATH
hack/local-up-master/master.sh
