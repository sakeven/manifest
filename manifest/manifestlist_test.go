package manifest

import (
	"os"
	"testing"
)

func TestPutManifestList(t *testing.T) {
	auth := &AuthInfo{
		Username: os.Getenv("DOCKER_USERNAME"),
		Password: os.Getenv("DOCKER_PASSWORD"),
	}

	_, err := PutManifestList(auth, "daocloud.io/daocloud/gosample:latest", "daocloud.io/daocloud/gosample:linux", "daocloud.io/daocloud/gosample:windows")
	if err != nil {
		t.Errorf("%s", err)
	}

	// t.Errorf("%s", digest)
}
