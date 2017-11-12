package app

import (
	"fmt"
	"strings"

	"github.com/docker/distribution/manifest/manifestlist"

	"github.com/sakeven/manifest/pkg/manifest"
	"github.com/sakeven/manifest/pkg/reference"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Run executes commands
func Run() {
	rootCmd.PersistentFlags().Bool("debug", false, "debugmode")
	rootCmd.PersistentFlags().String("username", "", "Username to access docker repository")
	rootCmd.PersistentFlags().String("password", "", "Password to access docker repository")
	rootCmd.PersistentFlags().String("cfg", config.Dir(), "docker configure file to access docker repository")
	rootCmd.AddCommand(createCmd, inspectCmd, annotateCmd)
	rootCmd.Execute()
}

// rootCmd root of cmd, shows manifest usage.
var rootCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Manifest is a tool for manager manifest of docker images",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if getBool(cmd.Flags(), "debug") {
			log.SetLevel(log.DebugLevel)
		}
	},
}

var createCmd = &cobra.Command{
	Use:   "create <target repository> <source repositories ...>",
	Short: "create and push a manifest list",
	Long:  `Create a manifest list named as target repository from source repositories, then push to registry`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		auth := getAuth(cmd.Flags())
		targetRepo := args[0]
		srcRepo := args[1:]
		digest, err := manifest.PutManifestList(auth, targetRepo, srcRepo...)
		if err != nil {
			log.Fatalf("%s", err)
		}
		fmt.Printf("Target image %s is digest %s\n", targetRepo, digest)
	},
}

var annotateCmd = &cobra.Command{
	Use:   "annotate <target repository>",
	Short: "annotate a manifest with platform spec",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// auth := getAuth(cmd.Flags())
		// targetRepo := args[0]
		// srcRepo := args[1:]
		// digest, err := manifest.PutManifestList(auth, targetRepo, srcRepo...)
		// if err != nil {
		// 	log.Fatalf("%s", err)
		// }
		// fmt.Printf("Target image %s is digest %s\n", targetRepo, digest)
	},
}

var inspectCmd = &cobra.Command{
	Use:   "inspect <repository>",
	Short: "inspect an image repository",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageName := args[0]
		namedRef, err := reference.ParseNamed(imageName)
		if err != nil {
			log.Fatalf("%s", err)
		}

		auth := getAuth(cmd.Flags())
		r, err := manifest.GetHTTPClient(auth, namedRef.Hostname())
		if err != nil {
			log.Fatalf("%s", err)
		}

		repo, id := manifest.Parse(namedRef)
		imgs, err := manifest.Inspect(r, repo, id)
		if err != nil {
			log.Fatalf("%s", err)
		}

		idx := 0
		for _, img := range imgs {
			if img.MediaType == manifestlist.MediaTypeManifestList {
				fmt.Printf("Name:   %s\n", imageName)
				fmt.Printf("Manifest Type: %s\n", img.MediaType)
				fmt.Printf("Digest: %s\n", img.Digest)
				fmt.Printf(" * Contains %d manifest references:\n", len(img.Manifest.(*manifestlist.DeserializedManifestList).Manifests))
				idx = 0
				continue
			}
			idx++
			fmt.Printf("%d    Manifest Type: %s\n", idx, img.MediaType)
			fmt.Printf("%d           Digest: %s\n", idx, img.Digest)
			fmt.Printf("%d  Manifest Length: %d\n", idx, img.Size)
			fmt.Printf("%d         Platform:\n", idx)
			fmt.Printf("%d           -      OS: %s\n", idx, img.Platform.OS)
			fmt.Printf("%d           -    Arch: %s\n", idx, img.Platform.Architecture)
			fmt.Printf("%d           - OS Vers: %s\n", idx, img.Platform.OSVersion)
			fmt.Printf("%d           - OS Feat: %s\n", idx, img.Platform.OSFeatures)
			fmt.Printf("%d           - Variant: %s\n", idx, img.Platform.Variant)
			fmt.Printf("%d           - Feature: %s\n", idx, strings.Join(img.Platform.Features, ","))
			fmt.Println()
		}
	},
}

func getAuth(flags *pflag.FlagSet) *manifest.AuthInfo {
	return &manifest.AuthInfo{
		Username:  getString(flags, "username"),
		Password:  getString(flags, "password"),
		DockerCfg: getString(flags, "cfg"),
	}
}

func getString(flags *pflag.FlagSet, flag string) string {
	val, err := flags.GetString(flag)
	if err != nil {
		log.Fatalf("%s", err)
	}
	return val
}

func getBool(flags *pflag.FlagSet, flag string) bool {
	val, err := flags.GetBool(flag)
	if err != nil {
		log.Fatalf("%s", err)
	}
	return val
}
