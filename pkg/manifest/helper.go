package manifest

import (
	"github.com/sakeven/manifest/pkg/reference"

	dreference "github.com/docker/distribution/reference"
)

// Parse returns repo name and tag or digest.
func Parse(targetRef reference.Named) (string, string) {
	tagOrDigest := "latest"
	switch v := targetRef.(type) {
	case dreference.Tagged:
		tagOrDigest = v.Tag()
	case dreference.Digested:
		tagOrDigest = v.Digest().String()
	}
	return targetRef.RemoteName(), tagOrDigest
}

func isSameRepo(a, b reference.Named) bool {
	return a.FullName() == b.FullName()
}

func isSameHub(a, b reference.Named) bool {
	return a.Hostname() == b.Hostname()
}
