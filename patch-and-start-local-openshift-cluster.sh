set -x
cd $GOPATH/src/github.com/openshift/
rm -rf origin
git clone https://github.com/openshift/origin --branch release-4.1
cd $GOPATH/src/github.com/openshift/origin
cd -
 patchsource --source-dir=$GOPATH/src/github.com/openshift/origin --code-config-yaml=$PATCHDIR/patch.yaml --include-vendor
cd $GOPATH/src/github.com/openshift/origin
echo "*******************************************************"
sudo cp -rf $GOPATH/src/github.com/fabric8-analytics /usr/local/go/src/github.com/
make build-all 
sed -i 's/hyperkube controller-manager/hyperkube kube-controller-manager/g' $GOPATH/src/github.com/openshift/origin/hack/local-up-master/lib.sh
hack/install-etcd.sh
cp $GOPATH/src/github.com/openshift/origin/_output/local/bin/linux/amd64/hypershift $GOPATH/bin
cp $GOPATH/src/github.com/openshift/origin/_output/local/bin/linux/amd64/hyperkube $GOPATH/bin
cp $GOPATH/src/github.com/openshift/origin/_output/local/bin/linux/amd64/oc $GOPATH/bin
export PATH=$GOPATH/src/github.com/openshift/origin/_output/tools/etcd/bin:$PATH
hack/local-up-master/master.sh
