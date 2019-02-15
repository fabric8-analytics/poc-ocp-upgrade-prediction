mgmt = graph.openManagement();

// for cluster version
cluster_version = mgmt.getPropertyKey('cluster_version');
if(cluster_version == null) {
    cluster_version = mgmt.makePropertyKey('cluster_version').dataType(String.class).make();
}

// for service
version = mgmt.getPropertyKey('version');
if(version == null) {
    version = mgmt.makePropertyKey('version').dataType(String.class).make();
}

name = mgmt.getPropertyKey('name');
if(name == null) {
    name = mgmt.makePropertyKey('name').dataType(String.class).make();
}

List<String> allKeys = [
        'name',
        'version',
        'cluster_version'
]

allKeys.each { k ->
    keyRef = mgmt.getPropertyKey(k);
    index_key = 'index_prop_key_'+k;
    if(null == mgmt.getGraphIndex(index_key)) {
        mgmt.buildIndex(index_key, Vertex.class).addKey(keyRef).buildCompositeIndex()
    }
}

mgmt.commit();
