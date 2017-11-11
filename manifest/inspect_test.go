package manifest

import (
	"os"
	"testing"
	// log "github.com/Sirupsen/logrus"
)

func TestImage(t *testing.T) {
	r := NewClient("daocloud.io", os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"))
	_, err := r.Inspect("daocloud/gosample", "windows")
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// t.Errorf("%#v", imgs)
	// t.Errorf("%s", imgs[0].CanonicalJSON)

	// time.Sleep(time.Second)
}
