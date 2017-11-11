package manifest

import (
	dreference "github.com/docker/distribution/reference"
	"github.com/sakeven/manifest/reference"
)

func parse(targetRef reference.Named) (string, string) {
	tagOrDigest := ""
	switch v := targetRef.(type) {
	case dreference.Tagged:
		tagOrDigest = v.Tag()
	case dreference.Digested:
		tagOrDigest = v.Digest().String()
	}
	return targetRef.RemoteName(), tagOrDigest
}
