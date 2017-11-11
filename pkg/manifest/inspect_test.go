package manifest

import (
	"os"
	"testing"

	"github.com/sakeven/manifest/pkg/registry"
	// log "github.com/Sirupsen/logrus"
)

func TestImage(t *testing.T) {
	r := registry.NewClient("daocloud.io", os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_PASSWORD"))
	_, err := Inspect(r, "daocloud/gosample", "windows")
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// t.Errorf("%#v", imgs)
	// t.Errorf("%s", imgs[0].CanonicalJSON)

	// time.Sleep(time.Second)
}
