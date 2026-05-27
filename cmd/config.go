package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `View or modify GoPack user configurations inside ~/.gopack/config.toml.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		if !viper.IsSet(key) {
			fmt.Printf("Config key '%s' not found.\n", key)
			os.Exit(1)
		}
		fmt.Println(viper.Get(key))
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		val := args[1]
		
		viper.Set(key, val)
		err := viper.WriteConfig()
		if err != nil {
			fmt.Printf("Failed to write config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set '%s' = '%s'\n", key, val)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		for _, k := range viper.AllKeys() {
			fmt.Printf("%s = %v\n", k, viper.Get(k))
		}
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}
