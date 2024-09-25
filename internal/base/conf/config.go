package conf

const DefaultLogLevel = "info"
const DefaultMasterLogName = "service.log"
const DefaultSlaveLogName = "slave.log"

var RootPassword = "1234"
var ResetPassword = false
var Http3Enabled = false
var Http3Port = 2091
var IgnoredPaths = []string{
	"/api/v1/ping",
	"/api/v1/unlock",
	"/api/v1/login",
	"/admin/api/v1/ping",
	"/admin/api/v1/unlock",
	"/admin/api/v1/login",
	"/swagger/*",
}
