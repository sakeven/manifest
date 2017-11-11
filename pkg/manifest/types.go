package manifest

// AuthInfo holds information about how manifest-tool should connect and authenticate to the docker registry
type AuthInfo struct {
	Username  string
	Password  string
	DockerCfg string
}
