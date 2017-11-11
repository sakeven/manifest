package main

import (
	"github.com/sakeven/manifest/manifest"

	"github.com/docker/docker/cli/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var RootCmd = &cobra.Command{
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
		log.Debugf("%#v", auth)
		for _, arg := range args {
			log.Debugf("%s", arg)
		}
		targetRepo := args[0]
		srcRepo := args[1:]
		digest, err := manifest.PutManifestList(auth, targetRepo, srcRepo...)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.Infof("Target image %s is digest %s", targetRepo, digest)
	},
}

func getAuth(flags *pflag.FlagSet) *manifest.AuthInfo {
	return &manifest.AuthInfo{
		Username:  getString(flags, "username"),
		Password:  getString(flags, "password"),
		DockerCfg: getString(flags, "cfg"),
	}
}

func main() {
	RootCmd.PersistentFlags().Bool("debug", false, "debugmode")
	RootCmd.PersistentFlags().String("username", "", "Username to access docker repository")
	RootCmd.PersistentFlags().String("password", "", "Password to access docker repository")
	RootCmd.PersistentFlags().String("cfg", config.Dir(), "docker configure file to access docker repository")
	RootCmd.AddCommand(pushCmd)
	RootCmd.Execute()
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
