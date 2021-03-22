package generic.cloudformation

hasSecretManager(str, document) {
	selectedSecret := strings.replace_n({"${": "", "}": ""}, regex.find_n(`\${\w+}`, str, 1)[0])
	document[selectedSecret].Type == "AWS::SecretsManager::Secret"
}

# Check if the type is ELB
isLoadBalancer(resource) {
	resource.Type == "AWS::ElasticLoadBalancing::LoadBalancer"
}

# Check if the type is ELB
isLoadBalancer(resource) {
	resource.Type == "AWS::ElasticLoadBalancingV2::LoadBalancer"
}

# Check if array contains an element
arrayContains(elem, array) {
	array[_] == elem
}

# Check if there is an action inside an array
checkAction(currentAction, actionToCompare) {
	is_string(currentAction)
    currentAction == "*"
    currentAction == actionToCompare
} else {
    is_string(currentAction)
	contains(lower(currentAction), actionToCompare)
}

# Dictionary of UDP ports
udpPortsMap = {
    53: "DNS",
    137: "NetBIOS Name Service",
    138: "NetBIOS Datagram Service",
    139: "NetBIOS Session Service",
    161: "SNMP",
    389: "LDAP",
    1434: "MSSQL Browser",
    2483: "Oracle DB SSL",
    2484: "Oracle DB SSL",
    5432: "PostgreSQL",
    11211: "Memcached",
    11214: "Memcached SSL",
    11215: "Memcached SSL",
}
