// Package main provides the entry point for the LogInform application.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/52617365/LogInform/internal"
	"github.com/spf13/cobra"
)

var explanations []internal.Explanation

var rootCmd = &cobra.Command{
	Use:   "loginform",
	Short: "LogInform - A tool for finding and explaining log identifiers",
	Long:  "LogInform helps you find and explain log identifiers in your content using predefined explanations.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open("explanations.yml")
		if err != nil {
			return fmt.Errorf("error reading explanations.yml")
		}
		defer f.Close()
		tmpExplanations, err := internal.LoadExplanationsFromReader(f)
		if err != nil {
			return err
		}
		explanations = tmpExplanations
		return nil
	},
}

var explainCmd = &cobra.Command{
	Use:   "explain [content...]",
	Short: "Find and print lines containing identifiers with explanations",
	Long:  "Find and print lines containing identifiers with explanations from the loaded explanations file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		content := strings.Join(args, " ")
		return internal.FindAndPrintMatches(content, explanations, os.Stdout)
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
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(inspectCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
