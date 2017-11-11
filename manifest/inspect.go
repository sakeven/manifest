package manifest

import (
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/image"
	"github.com/opencontainers/go-digest"
	"github.com/sakeven/manifest/registry"
	log "github.com/sirupsen/logrus"
)

// ImageInspect stores image inspect information
type ImageInspect struct {
	Size          int64
	MediaType     string
	Tag           string
	Digest        digest.Digest
	Platform      manifestlist.PlatformSpec
	References    []string
	CanonicalJSON []byte
	Manifest      distribution.Manifest
}

// Inspect get images inspect information
func Inspect(r *registry.Client, repository, tag string) ([]ImageInspect, error) {
	m, err := r.FetchManifest(repository, tag)
	if err != nil {
		log.Errorf("Failed to fetch manifest %s", err)
		return nil, err
	}

	var imgs []*image.Image

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
		log.Debugf("%#v\n %d", img, img.Size)
		imgs = append(imgs, img)
	case *manifestlist.DeserializedManifestList:
	}

	return populate(imgs, tag, m)
}

func populate(imgs []*image.Image, tag string, m distribution.Manifest) ([]ImageInspect, error) {
	mediaType, payload, err := m.Payload()
	if err != nil {
		return nil, err
	}

	imgInspect := make([]ImageInspect, len(imgs))

	for i, img := range imgs {
		platform := manifestlist.PlatformSpec{
			Architecture: img.Architecture,
			OS:           img.OS,
			OSVersion:    img.OSVersion,
			OSFeatures:   img.OSFeatures,
			Features:     img.OSFeatures,
		}
		imgInspect[i] = ImageInspect{
			Size:          int64(len(payload)),
			MediaType:     mediaType,
			Tag:           tag,
			Digest:        digest.FromBytes(payload),
			Platform:      platform,
			CanonicalJSON: payload,
			Manifest:      m,
		}
	}

	return imgInspect, nil
}
