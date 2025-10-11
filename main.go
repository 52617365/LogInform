// Package main provides the entry point for the LogInform application.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/52617365/LogInform/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var explanations []internal.Explanation

var rootCmd = &cobra.Command{
	Use:   "loginform",
	Short: "LogInform - A tool for finding and explaining log identifiers",
	Long:  "LogInform helps you find and explain log identifiers in your content using predefined explanations.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var fileLookupError viper.ConfigFileNotFoundError
		if err := viper.ReadInConfig(); err != nil {
			if errors.As(err, &fileLookupError) {
				slog.Error("couldn't find a configuration file. looked in current path and user root.")
				// Indicates an explicitly set config file is not found (such as with
				// using `viper.SetConfigFile`) or that no config file was found in
				// any search path (such as when using `viper.AddConfigPath`)
			} else {
				slog.Error("found configuration file but the structure was invalid.", "config", viper.ConfigFileUsed())
				// Config file was found but another error was produced
			}
		}
		tmpExplanations := internal.FetchExplanations()
		explanations = tmpExplanations
		return nil
	},
}

var explainCmd = &cobra.Command{
	Use:   "explain [file]",
	Short: "Find and print lines containing identifiers with explanations",
	Long:  "Find and print lines containing identifiers with explanations from the loaded explanations file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		stdin, _ := cmd.Flags().GetBool("stdin")

		if stdin {
			content := strings.Join(args, " ")
			return internal.FindAndPrintMatches(strings.NewReader(content), explanations, os.Stdout)
		}

		if len(args) == 0 && !stdin {
			return fmt.Errorf("please provide a file path or use --stdin flag")
		}

		if _, err := os.Stat(args[0]); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s\nDid you mean to use --stdin to pass content directly?", args[0])
			}
			return fmt.Errorf("error checking file: %w", err)
		}

		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()

		return internal.FindAndPrintMatches(file, explanations, os.Stdout)
	},
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Print all loaded explanations",
	Long:  "Print all loaded explanations from the explanations file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.InspectExplanations(explanations, os.Stdout)
	},
}

func init() {
	explainCmd.Flags().Bool("inline", false, "Pass content directly as arguments instead of reading from file")

	viper.AddConfigPath("/etc/.explanations/")
	viper.AddConfigPath("$HOME/.explanations")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(inspectCmd)
}

func main() {
	rootCmd.Execute()
}
