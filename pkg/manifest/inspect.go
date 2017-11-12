package manifest

import (
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
		ms = append(ms, m)
	case *manifestlist.DeserializedManifestList:
		imgs = append(imgs, &image.Image{})
		ms = append(ms, v)
		for _, m := range v.Manifests {
			log.Debugf("ml digest %s", m.Digest)
			manifest, err := r.FetchManifest(repository, m.Digest.String())
			if err != nil {
				return nil, err
			}
			switch v := manifest.(type) {
			case *schema2.DeserializedManifest:
				blob, err := r.PullBlob(repository, v.Config.Digest.String())
				if err != nil {
					return nil, err
				}
				img, err := image.NewFromJSON(blob)
				if err != nil {
					return nil, err
				}
				imgs = append(imgs, img)
				ms = append(ms, manifest)
			}
		}

		log.Debugf("%#v", v)
	}

	return populate(imgs, tag, ms)
}

func populate(imgs []*image.Image, tag string, m []distribution.Manifest) ([]ImageInspect, error) {
	imgInspect := make([]ImageInspect, len(imgs))

	for i, img := range imgs {
		mediaType, payload, err := m[i].Payload()
		if err != nil {
			return nil, err
		}
		platform := manifestlist.PlatformSpec{
			Architecture: img.Architecture,
			OS:           img.OS,
			OSVersion:    img.OSVersion,
			OSFeatures:   img.OSFeatures,
			Features:     img.OSFeatures,
		}
		imgInspect[i] = ImageInspect{
			Size:      int64(len(payload)),
			MediaType: mediaType,
			Tag:       tag,
			Digest:    digest.FromBytes(payload),
			Platform:  platform,
			Manifest:  m[i],
		}
	}

	return imgInspect, nil
}
