mgmt = graph.openManagement();

// for cluster version
cluster_version = mgmt.getPropertyKey('cluster_version');
if(cluster_version == null) {
    cluster_version = mgmt.makePropertyKey('cluster_version').dataType(String.class).make();
}

// for service version
service_version = mgmt.getPropertyKey('service_version');
if(service_version == null) {
    service_version = mgmt.makePropertyKey('service_version').dataType(String.class).make();
}


// for package
packagenode = mgmt.getPropertyKey('package');
if(packagenode == null) {
    packagenode = mgmt.makePropertyKey('package').dataType(String.class).make();
}

// for function
functionnode = mgmt.getPropertyKey('function');
if(functionnode == null) {
    functionnode = mgmt.makePropertyKey('function').dataType(String.class).make();
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

vertex_label = mgmt.getPropertyKey('vertex_label');
if(vertex_label == null) {
    vertex_label = mgmt.makePropertyKey('vertex_label').dataType(String.class).make();
}

local_name = mgmt.getPropertyKey('local_name');
if(local_name == null) {
    local_name = mgmt.makePropertyKey('local_name').dataType(String.class).make();
}

importpath = mgmt.getPropertyKey('importpath');
if(importpath == null) {
    importpath = mgmt.makePropertyKey('importpath').dataType(String.class).make();
}

serviceName = mgmt.getPropertyKey('service_name');
if (serviceName == null) {
    serviceName = mgmt.makePropertyKey('service_name').dataType(String.class).make();
}

List<String> allKeys = [
        'name',
        'version',
        'cluster_version',
	    'service_version',
        'package',
        'function',
        'local_name',
        'importpath'
]

allKeys.each { k ->
    keyRef = mgmt.getPropertyKey(k);
    index_key = 'index_prop_key_'+k;
    if(null == mgmt.getGraphIndex(index_key)) {
        mgmt.buildIndex(index_key, Vertex.class).addKey(keyRef).buildCompositeIndex()
        mgmt.buildIndex(index_key + '_labelled', Vertex.class).addKey(mgmt.getPropertyKey('vertex_label')).addKey(keyRef).buildCompositeIndex()
    }
}

// Create the edge indexes
edgeLabel = mgmt.makePropertyKey('edge_label').dataType(String.class).make();
mgmt.buildIndex('index_prop_key_edge_label', Edge.class).addKey(edgeLabel).buildCompositeIndex();
mgmt.buildIndex('index_composite_func_vertex', Vertex.class).addKey(mgmt.getPropertyKey('vertex_label')).addKey(packagenode).addKey(serviceName).addKey(name).buildCompositeIndex();
mgmt.commit();
