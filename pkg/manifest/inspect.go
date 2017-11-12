package manifest

import (
	"encoding/json"
	"os"

	"github.com/sakeven/manifest/pkg/registry"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/image"
	"github.com/opencontainers/go-digest"
)

// ImageInspect stores image inspect information
type ImageInspect struct {
	Size       int64
	MediaType  string
	Tag        string
	Digest     digest.Digest
	Platform   manifestlist.PlatformSpec
	References []string
	Manifest   distribution.Manifest
}

// Inspect get images inspect information
func Inspect(r *registry.Client, repository, tag string) ([]ImageInspect, error) {
	m, err := r.FetchManifest(repository, tag)
	if err != nil {
		return nil, err
	}

	var imgs []*image.Image
	var ms []distribution.Manifest
	var platforms []manifestlist.PlatformSpec

	switch v := m.(type) {
	case *schema1.SignedManifest:
	case *schema2.DeserializedManifest:
		log.Debugf("%#v", v)
		blob, err := r.PullBlob(repository, v.Config.Digest.String())
		if err != nil {
			return nil, err
		}
		img, err := image.NewFromJSON(blob)
		if err != nil {
			return nil, err
		}
		imgs = append(imgs, img)
		platform := manifestlist.PlatformSpec{
			Architecture: img.Architecture,
			OS:           img.OS,
			OSVersion:    img.OSVersion,
			OSFeatures:   img.OSFeatures,
		}
		platforms = append(platforms, platform)
		ms = append(ms, m)
	case *manifestlist.DeserializedManifestList:
		json.NewEncoder(os.Stdout).Encode(v)
		ms = append(ms, v)
		platforms = append(platforms, manifestlist.PlatformSpec{})
		for _, m := range v.Manifests {
			log.Debugf("ml digest %s", m.Digest)
			manifest, err := r.FetchManifest(repository, m.Digest.String())
			if err != nil {
				return nil, err
			}
			ms = append(ms, manifest)
			platforms = append(platforms, m.Platform)
			// switch v := manifest.(type) {
			// case *schema2.DeserializedManifest:
			// 	blob, err := r.PullBlob(repository, v.Config.Digest.String())
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// 	img, err := image.NewFromJSON(blob)
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// 	imgs = append(imgs, img)
			// }
		}
		log.Debugf("%#v", v)
	}

	return populate(platforms, tag, ms)
}

func populate(platforms []manifestlist.PlatformSpec, tag string, ms []distribution.Manifest) ([]ImageInspect, error) {
	imgInspect := make([]ImageInspect, len(ms))
	for i, m := range ms {
		mediaType, payload, err := m.Payload()
		if err != nil {
			return nil, err
		}
		imgInspect[i] = ImageInspect{
			Size:      int64(len(payload)),
			MediaType: mediaType,
			Tag:       tag,
			Digest:    digest.FromBytes(payload),
			Platform:  platforms[i],
			Manifest:  m,
		}
	}

	return imgInspect, nil
}
