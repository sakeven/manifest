package registry

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
)

// PushManifest pushs manifest to distrubiton.
func (r *Client) PushManifest(repository, tag string, m distribution.Manifest) (digest.Digest, error) {
	mediaType, p, err := m.Payload()
	req, err := r.newRequest("PUT", fmt.Sprintf("/v2/%s/manifests/%s", repository, tag), bytes.NewReader(p))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mediaType)

	resp, err := r.do(req, nil)
	if err != nil {
		return "", err
	}

	dgstHeader := resp.Header.Get("Docker-Content-Digest")
	return digest.Parse(dgstHeader)
}

// FetchManifest gets manifest from distrubiton
func (r *Client) FetchManifest(repository, tag string) (distribution.Manifest, error) {
	req, err := r.newRequest("GET", fmt.Sprintf("/v2/%s/manifests/%s", repository, tag), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", manifestlist.MediaTypeManifestList)
	req.Header.Add("Accept", schema2.MediaTypeManifest)

	bf := new(bytes.Buffer)
	resp, err := r.do(req, bf)
	if err != nil {
		return nil, err
	}

	var m distribution.Manifest

	contentType := resp.Header.Get("Content-Type")
	switch contentType {
	case schema1.MediaTypeManifest, schema1.MediaTypeSignedManifest:
		m = &schema1.SignedManifest{}
	case schema2.MediaTypeManifest:
		m = &schema2.DeserializedManifest{}
	case manifestlist.MediaTypeManifestList:
		m = &manifestlist.DeserializedManifestList{}
	}

	err = json.Unmarshal(bf.Bytes(), m)
	return m, err
}

// PullBlob pulls blob
func (r *Client) PullBlob(repository, sha string) ([]byte, error) {
	req, err := r.newRequest("GET", fmt.Sprintf("/v2/%s/blobs/%s", repository, sha), nil)
	if err != nil {
		return nil, err
	}

	bf := new(bytes.Buffer)
	_, err = r.do(req, bf)
	if err != nil {
		return nil, err
	}

	return bf.Bytes(), err
}

// MountBlob mounts a blob from other repository.
func (r *Client) MountBlob(repository string, digest string, fromRepo string) (string, error) {
	url := fmt.Sprintf("/v2/%s/blobs/uploads/?mount=%s&from=%s", repository, digest, fromRepo)
	req, err := r.newRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Length", "0")

	resp, err := r.do(req, nil)
	if err != nil {
		return "", err
	}
	return resp.Header.Get("Location"), nil
}
