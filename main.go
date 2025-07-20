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
	Use:   "explain [file]",
	Short: "Find and print lines containing identifiers with explanations",
	Long:  "Find and print lines containing identifiers with explanations from the loaded explanations file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		inline, _ := cmd.Flags().GetBool("inline")

		if inline {
			content := strings.Join(args, " ")
			return internal.FindAndPrintMatches(strings.NewReader(content), explanations, os.Stdout)
		}

		if len(args) == 0 {
			return fmt.Errorf("please provide a file path or use --inline flag")
		}

		if _, err := os.Stat(args[0]); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s\nDid you mean to use --inline to pass content directly?\n", args[0])
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
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(inspectCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
