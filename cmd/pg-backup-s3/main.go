package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ashwanthkumar/slack-go-webhook"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	defaultConfigFilename = "pg-backup-s3"
	envPrefix             = "PGBACKUPS3"

	rootCmd = &cobra.Command{
		Use:   "pg-backup-s3",
		Short: "pg-backup-s3 is a utility to take encrypted PostgreSQL backups",
		Long:  "pg-backup-s3 is a utility to take encrypted PostgreSQL backups",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
	}
)

func main() {
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	err := rootCmd.Execute()
	if err != nil {
		notifyError(err)
		log.Fatal(err)
	}
}

// lifesaving work by Carolyn Van Slyck (https://carolynvanslyck.com/blog/2020/08/sting-of-the-viper/)

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(defaultConfigFilename)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(".")

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

// notifyError allows to send the error to a Slack or compatible (like Discord's /slack) webhook
func notifyError(err error) {
	// keeping this outside if cobra
	url := os.Getenv("ERROR_WEBHOOK")
	if url == "" {
		return
	}

	log.Println("Notifying webhook")
	attachment := slack.Attachment{}
	attachment.AddField(slack.Field{Title: "Error", Value: err.Error()})
	payload := slack.Payload{
		Text:        "pg-backup-s3 returned an error",
		Attachments: []slack.Attachment{attachment},
	}
	errs := slack.Send(url, "", payload)
	if len(errs) > 0 {
		fmt.Printf("error: %s\n", err)
	}
}
