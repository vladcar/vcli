package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vcli",
	Short: "Vlad's personal CLI utils",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", fmt.Sprintf("%v/vcli.yaml", homeDir), "config file (default is $HOME/vcli.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".vcli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("vcli.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match
}
