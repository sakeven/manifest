package app

import (
	"fmt"
	"strings"

	"github.com/sakeven/manifest/pkg/manifest"
	"github.com/sakeven/manifest/pkg/reference"

	"github.com/docker/docker/cli/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Run executes commands
func Run() {
	rootCmd.PersistentFlags().Bool("debug", false, "debugmode")
	rootCmd.PersistentFlags().String("username", "", "Username to access docker repository")
	rootCmd.PersistentFlags().String("password", "", "Password to access docker repository")
	rootCmd.PersistentFlags().String("cfg", config.Dir(), "docker configure file to access docker repository")
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(inspectCmd)
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

var pushCmd = &cobra.Command{
	Use:   "push <target repository> <source repositories ...>",
	Short: "create and push a manifest list",
	Long:  `Create a manifest list named as target repository from source repositories, then push to registry`,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		auth := getAuth(cmd.Flags())
		targetRepo := args[0]
		srcRepo := args[1:]
		digest, err := manifest.PutManifestList(auth, targetRepo, srcRepo...)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.Infof("Target image %s is digest %s", targetRepo, digest)
	},
}

var inspectCmd = &cobra.Command{
	Use:   "inspect <repository>",
	Short: "inspect an image repository",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		namedRef, err := reference.ParseNamed(args[0])
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

		for i, img := range imgs {
			fmt.Printf("%d    Manifest Type: %s\n", i+1, img.MediaType)
			fmt.Printf("%d           Digest: %s\n", i+1, img.Digest)
			fmt.Printf("%d  Manifest Length: %d\n", i+1, img.Size)
			fmt.Printf("%d         Platform:\n", i+1)
			fmt.Printf("%d           -      OS: %s\n", i+1, img.Platform.OS)
			fmt.Printf("%d           -    Arch: %s\n", i+1, img.Platform.Architecture)
			fmt.Printf("%d           - OS Vers: %s\n", i+1, img.Platform.OSVersion)
			fmt.Printf("%d           - OS Feat: %s\n", i+1, img.Platform.OSFeatures)
			fmt.Printf("%d           - Variant: %s\n", i+1, img.Platform.Variant)
			fmt.Printf("%d           - Feature: %s\n", i+1, strings.Join(img.Platform.Features, ","))
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
