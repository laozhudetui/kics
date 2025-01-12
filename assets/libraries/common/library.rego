package generic.common

getDocument = document { # is this really fully generic?
	document := input.document
}

json_unmarshal(s) = result {
	s == null
	result := json.unmarshal("{}")
}

json_unmarshal(s) = result {
	s != null
	result := json.unmarshal(s)
}

calc_IP_value(ip) = result {
	ips := split(ip, ".")

	#calculate the value of an ip
	#a.b.c.d
	#a*16777216 + b*65536 + c*256 + d
	result = (((to_number(ips[0]) * 16777216) + (to_number(ips[1]) * 65536)) + (to_number(ips[2]) * 256)) + to_number(ips[3])
}

isDefaultPassword(p) {
	ar = {
		"!@",
		"root",
		"wubao",
		"password",
		"123456",
		"admin",
		"12345",
		"1234",
		"p@ssw0rd",
		"123",
		"1",
		"jiamima",
		"test",
		"root123",
		"!",
		"!q@w",
		"!qaz@wsx",
		"idc!@",
		"admin!@",
		"",
		"alpine",
		"qwerty",
		"12345678",
		"111111",
		"123456789",
		"1q2w3e4r",
		"123123",
		"default",
		"1234567",
		"qwe123",
		"1qaz2wsx",
		"1234567890",
		"abcd1234",
		"000000",
		"user",
		"toor",
		"qwer1234",
		"1q2w3e",
		"asdf1234",
		"redhat",
		"1234qwer",
		"cisco",
		"12qwaszx",
		"test123",
		"1q2w3e4r5t",
		"admin123",
		"changeme",
		"1qazxsw2",
		"123qweasd",
		"q1w2e3r4",
		"letmein",
		"server",
		"root1234",
		"master",
		"abc123",
		"rootroot",
		"a",
		"system",
		"pass",
		"1qaz2wsx3edc",
		"p@$$w0rd",
		"112233",
		"welcome",
		"!QAZ2wsx",
		"linux",
		"123321",
		"manager",
		"1qazXSW@",
		"q1w2e3r4t5",
		"oracle",
		"asd123",
		"admin123456",
		"ubnt",
		"123qwe",
		"qazwsxedc",
		"administrator",
		"superuser",
		"zaq12wsx",
		"121212",
		"654321",
		"ubuntu",
		"0000",
		"zxcvbnm",
		"root@123",
		"1111",
		"vmware",
		"q1w2e3",
		"qwerty123",
		"cisco123",
		"11111111",
		"pa55w0rd",
		"asdfgh",
		"11111",
		"123abc",
		"asdf",
		"centos",
		"888888",
		"54321",
		"password123",
	}

	ar[p]
}

isCommonValue(p) {
	bl = {
		"RESOURCE",
		"GROUP",
		"SUBNET",
		"S3",
		"SERVICE",
		"AZURE",
		"BUCKET",
		"VIRTUAL",
		"NETWORK",
		"POLICY",
		"AWS",
		"PROTOCOL",
		"CLOUD",
		"MINUTE",
		"TLS",
		"EC2",
		"VPC",
		"INTERNET",
		"ROUTE",
		"EFS",
		"INSTANCE",
		"VPN",
		"MOUNT",
		"MYSQL",
		"APACHE",
		"ETHERNET",
		"TERRAFORM",
		"TARGET",
		"ENVIRONMENT",
		"MEMORY",
		"PACKAGE",
		"STATEMENT",
		"REGION",
		"INGRESS",
		"CHECKPOINT",
		"MODULE",
		"BASIC",
		"NUMBER",
		"MASLEN",
	}

	black := bl[_]
	contains(upper(p), black)
}

isCommonKey(p) {
	bl = {
		"namespace",
		"bypass",
		"name",
		"ref",
		"base64",
		"pattern",
		"author",
		"group",
		"image",
		"host",
		"interface",
		"service",
		"src",
		"value",
		"default",
		"sku",
		"condition",
		"status",
		"size",
		"runtime",
		"id",
		"chdir",
		"env",
		"person",
		"kind",
		"content",
		"age",
		"length",
		"prevention",
		"change",
		"attribute",
		"stage",
		"version",
		"tag",
		"alert",
		"device",
		"type",
		"java",
		"metadata",
		"child",
		"sc1",
		"task",
		"memory",
		"storage",
		"bundle",
		"label",
		"origin",
		"upstream",
		"time",
		"project",
		"from",
		"maven",
		"destination",
		"shape",
		"local",
		"target",
		"exported",
		"zone",
	}

	black := bl[_]
	contains(lower(p), black)
}

# verifies if the resource(statement.Principal.AWS) contains an ARN that points to a specific IAM user
allowsAllPrincipalsToAssume(resource, statement) {
	is_string(resource) == true
	contains(resource, "arn:aws:iam::")
	contains(resource, ":root")
	not contains(statement.Effect, "Deny")
}

allowsAllPrincipalsToAssume(resource, statement) {
	is_array(resource) == true
	contains(resource[x], "arn:aws:iam::")
	contains(resource[x], ":root")
	not contains(statement.Effect, "Deny")
}
