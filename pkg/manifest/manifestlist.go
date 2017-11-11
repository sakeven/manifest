package manifest

import (
	"fmt"

	"github.com/sakeven/manifest/pkg/reference"
	"github.com/sakeven/manifest/pkg/registry"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	engineTypes "github.com/docker/docker/api/types"
	registryTypes "github.com/docker/docker/api/types/registry"
	"github.com/opencontainers/go-digest"
	log "github.com/sirupsen/logrus"
)

// we will store up a list of blobs we must ask the registry
// to cross-mount into our target namespace
type blobMount struct {
	FromRepo string
	Digest   string
}

// PutManifestList takes an authentication variable and pushes an image list based on the spec
func PutManifestList(a *AuthInfo, dstImage string, srcImages ...string) (string, error) {
	var (
		manifestList      manifestlist.ManifestList
		blobMountRequests []blobMount
		manifestRequests  []distribution.Manifest
	)

	// process the target image name reference
	targetRef, err := reference.ParseNamed(dstImage)
	if err != nil {
		return "", fmt.Errorf("error parsing name for %s: %s", dstImage, err)
	}

	// Now create the manifest list payload by looking up the manifest schemas
	// for the constituent images:
	log.Info("Retrieving digests of images...")
	for _, img := range srcImages {
		namedRef, err := reference.ParseNamed(img)
		if err != nil {
			return "", err
		}

		r, err := GetHTTPClient(a, namedRef.Hostname())
		if err != nil {
			return "", err
		}

		repo, tagOrDigest := Parse(namedRef)
		log.Debugf("%s %s", repo, tagOrDigest)
		mfstData, err := Inspect(r, repo, tagOrDigest)
		if err != nil {
			return "", fmt.Errorf("inspect of image %s failed with error: %v", img, err)
		}

		// TODO support different repositroy.
		if namedRef.Hostname() != targetRef.Hostname() {
			return "", fmt.Errorf("cannot use source images from a different registry than the target image: %s != %s", namedRef.Hostname(), targetRef.Hostname())
		}

		if len(mfstData) > 1 {
			// too many responses--can only happen if a manifest list was returned for the name lookup
			return "", fmt.Errorf("manifest lists do not allow recursion")
		}

		// the non-manifest list case will always have exactly one manifest response
		imgMfst := mfstData[0]
		manifest := manifestlist.ManifestDescriptor{
			Platform: imgMfst.Platform,
			Descriptor: distribution.Descriptor{
				Digest:    imgMfst.Digest,
				Size:      imgMfst.Size,
				MediaType: imgMfst.MediaType,
			},
		}

		log.Infof("Image %s is digest %s; size: %d", img, imgMfst.Digest, imgMfst.Size)

		// if this image is in a different repo, we need to add the layer & config digests to the list of
		// requested blob mounts (cross-repository push) before pushing the manifest list
		if targetRef.FullName() != namedRef.FullName() {
			log.Debugf("Adding manifest references of %s to blob mount requests", img)
			for _, layer := range imgMfst.References {
				blobMountRequests = append(blobMountRequests, blobMount{FromRepo: namedRef.FullName(), Digest: layer})
			}
			// also must add the manifest to be pushed in the target namespace
			log.Debugf("Adding manifest %s -> to be pushed to %s as a manifest reference", namedRef.FullName(), namedRef.FullName())
			manifestRequests = append(manifestRequests, imgMfst.Manifest)
		}
		manifestList.Manifests = append(manifestList.Manifests, manifest)
	}

	deserializedManifestList, err := manifestlist.FromDescriptors(manifestList.Manifests)
	if err != nil {
		return "", fmt.Errorf("cannot deserialize manifest list: %s", err)
	}

	httpClient, err := GetHTTPClient(a, targetRef.Hostname())
	if err != nil {
		return "", fmt.Errorf("failed to setup HTTP client to repository: %s", err)
	}

	// before we push the manifest list, if we have any blob mount requests, we need
	// to ask the registry to mount those blobs in our target so they are available
	// as references
	if err := mountBlobs(httpClient, targetRef, blobMountRequests); err != nil {
		return "", fmt.Errorf("failed to mount blobs for cross-repository push: %s", err)
	}

	// we also must push any manifests that are referenced in the manifest list into
	// the target namespace
	if err := pushReferences(httpClient, targetRef, manifestRequests); err != nil {
		return "", fmt.Errorf("failed to push manifests referenced: %s", err)
	}

	// push final manifest
	repo, tag := Parse(targetRef)
	finalDigest, err := httpClient.PushManifest(repo, tag, deserializedManifestList)
	if err != nil {
		return "", fmt.Errorf("push manifest list failed: %s", err)
	}

	return string(finalDigest), nil
}

// GetHTTPClient gets registry cleint
func GetHTTPClient(a *AuthInfo, endpoint string) (*registry.Client, error) {
	authConfig, err := getAuthConfig(a, nil) // TODO
	if err != nil {
		return nil, fmt.Errorf("Cannot retrieve authconfig: %v", err)
	}

	return registry.NewClient(endpoint, authConfig.Username, authConfig.Password), nil
}

func pushReferences(httpClient *registry.Client, ref reference.Named, manifests []distribution.Manifest) error {
	// for each referenced manifest object in the manifest list (that is outside of our current repo/name)
	// we need to push by digest the manifest so that it is added as a valid reference in the current
	// repo. This will allow us to push the manifest list properly later and have all valid references.

	// first get rid of possible hostname so the target URL is constructed properly
	name := ref.String()
	ref, err := reference.ParseNamed(name)
	if err != nil {
		return fmt.Errorf("Error parsing repo/name portion of reference without hostname: %s: %v", name, err)
	}
	for _, manifest := range manifests {
		_, p, err := manifest.Payload()
		if err != nil {
			return err
		}

		dgst := digest.FromBytes(p)
		dgstResult, err := httpClient.PushManifest(ref.Name(), dgst.String(), manifest)
		if err != nil {
			return fmt.Errorf("couldn't push manifest: %v", err)
		}
		if dgstResult != dgst {
			return fmt.Errorf("Pushed referenced manifest received a different digest: expected %s, got %s", dgst, dgstResult)
		}
	}
	return nil
}

func mountBlobs(httpClient *registry.Client, ref reference.Named, blobsRequested []blobMount) error {
	for _, blob := range blobsRequested {
		location, err := httpClient.MountBlob(ref.RemoteName(), blob.Digest, blob.FromRepo)
		if err != nil {
			log.Errorf("Mount failed %s", err)
			return err
		}
		log.Debugf("Mount of blob %s succeeded, location: %q", blob.Digest, location)
	}
	return nil
}

// getAuthConfig gets auth config for specific registry.
func getAuthConfig(a *AuthInfo, index *registryTypes.IndexInfo) (engineTypes.AuthConfig, error) {
	var (
		username = a.Username
		password = a.Password
		// cfg           = a.DockerCfg
		defAuthConfig = engineTypes.AuthConfig{
			Username: a.Username,
			Password: a.Password,
			Email:    "stub@example.com",
		}
	)

	if username != "" && password != "" {
		return defAuthConfig, nil
	}

	// confFile, err := config.Load(cfg)
	// if err != nil {
	// 	return engineTypes.AuthConfig{}, err
	// }

	// authConfig := registry.ResolveAuthConfig(confFile.AuthConfigs, index)
	// logrus.Debugf("authConfig for %s: %v", index.Name, authConfig.Username)

	return defAuthConfig, nil
}
