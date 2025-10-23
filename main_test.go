package main

import (
	"flag"
	"testing"
)

func TestCommandLine(t *testing.T) {
	t.Run("FlagParsing", func(t *testing.T) {
		t.Run("VersionFlag", func(t *testing.T) {
			// Arrange
			cleanup := setupTestFlags([]string{"cwlogs", "--version"})
			defer cleanup()
			
			// Act
			flagVersion, flagProfile, flagRegion, err := parseTestFlags()
			
			// Assert
			assertNoError(t, err)
			assertBoolEqual(t, *flagVersion, true, "version flag")
			assertStringEqual(t, *flagProfile, "")
			assertStringEqual(t, *flagRegion, "")
		})
		
		t.Run("ProfileFlag", func(t *testing.T) {
			// Arrange
			cleanup := setupTestFlags([]string{"cwlogs", "--profile", "myprofile"})
			defer cleanup()
			
			// Act
			flagVersion, flagProfile, flagRegion, err := parseTestFlags()
			
			// Assert
			assertNoError(t, err)
			assertBoolEqual(t, *flagVersion, false, "version flag")
			assertStringEqual(t, *flagProfile, "myprofile")
			assertStringEqual(t, *flagRegion, "")
		})
		
		t.Run("RegionFlag", func(t *testing.T) {
			// Arrange
			cleanup := setupTestFlags([]string{"cwlogs", "--region", "us-west-2"})
			defer cleanup()
			
			// Act
			flagVersion, flagProfile, flagRegion, err := parseTestFlags()
			
			// Assert
			assertNoError(t, err)
			assertBoolEqual(t, *flagVersion, false, "version flag")
			assertStringEqual(t, *flagProfile, "")
			assertStringEqual(t, *flagRegion, "us-west-2")
		})
		
		t.Run("CombinedFlags", func(t *testing.T) {
			// Arrange
			cleanup := setupTestFlags([]string{"cwlogs", "--profile", "prod", "--region", "eu-west-1"})
			defer cleanup()
			
			// Act
			flagVersion, flagProfile, flagRegion, err := parseTestFlags()
			
			// Assert
			assertNoError(t, err)
			assertBoolEqual(t, *flagVersion, false, "version flag")
			assertStringEqual(t, *flagProfile, "prod")
			assertStringEqual(t, *flagRegion, "eu-west-1")
		})
		
		t.Run("NoFlags", func(t *testing.T) {
			// Arrange
			cleanup := setupTestFlags([]string{"cwlogs"})
			defer cleanup()
			
			// Act
			flagVersion, flagProfile, flagRegion, err := parseTestFlags()
			
			// Assert
			assertNoError(t, err)
			assertBoolEqual(t, *flagVersion, false, "version flag")
			assertStringEqual(t, *flagProfile, "")
			assertStringEqual(t, *flagRegion, "")
		})
	})
	
	t.Run("PositionalArguments", func(t *testing.T) {
		t.Run("ProfileOnly", func(t *testing.T) {
			// Arrange
			cleanup := setupTestFlags([]string{"cwlogs", "myprofile"})
			defer cleanup()
			
			// Act
			flagVersion, flagProfile, flagRegion, err := parseTestFlags()
			
			// Assert
			assertNoError(t, err)
			assertBoolEqual(t, *flagVersion, false, "version flag")
			assertStringEqual(t, *flagProfile, "")
			assertStringEqual(t, *flagRegion, "")
			
			// Check positional arguments
			args := flag.CommandLine.Args()
			assertSliceLength(t, args, 1, "positional arguments")
			assertStringEqual(t, args[0], "myprofile")
		})
		
		t.Run("ProfileAndRegion", func(t *testing.T) {
			// Arrange
			cleanup := setupTestFlags([]string{"cwlogs", "myprofile", "us-east-1"})
			defer cleanup()
			
			// Act
			flagVersion, flagProfile, flagRegion, err := parseTestFlags()
			
			// Assert
			assertNoError(t, err)
			assertBoolEqual(t, *flagVersion, false, "version flag")
			assertStringEqual(t, *flagProfile, "")
			assertStringEqual(t, *flagRegion, "")
			
			// Check positional arguments
			args := flag.CommandLine.Args()
			assertSliceLength(t, args, 2, "positional arguments")
			assertStringEqual(t, args[0], "myprofile")
			assertStringEqual(t, args[1], "us-east-1")
		})
		
		t.Run("MixedFlagAndPositional", func(t *testing.T) {
			// Arrange
			cleanup := setupTestFlags([]string{"cwlogs", "--profile", "flagprofile", "positionalprofile"})
			defer cleanup()
			
			// Act
			flagVersion, flagProfile, flagRegion, err := parseTestFlags()
			
			// Assert
			assertNoError(t, err)
			assertBoolEqual(t, *flagVersion, false, "version flag")
			assertStringEqual(t, *flagProfile, "flagprofile")
			assertStringEqual(t, *flagRegion, "")
			
			// Check positional arguments (flag takes precedence)
			args := flag.CommandLine.Args()
			assertSliceLength(t, args, 1, "positional arguments")
			assertStringEqual(t, args[0], "positionalprofile")
		})
	})
	
	t.Run("ArgumentValidation", func(t *testing.T) {
		t.Run("ValidArguments", func(t *testing.T) {
			tests := []struct {
				name string
				args []string
			}{
				{"no args", []string{"cwlogs"}},
				{"profile only", []string{"cwlogs", "dev"}},
				{"profile and region", []string{"cwlogs", "dev", "us-west-2"}},
				{"flags only", []string{"cwlogs", "--profile", "dev", "--region", "us-west-2"}},
			}
			
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					// Arrange
					cleanup := setupTestFlags(tt.args)
					defer cleanup()
					
					// Act
					_, _, _, err := parseTestFlags()
					
					// Assert
					assertNoError(t, err)
				})
			}
		})
	})
}
