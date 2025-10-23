package main

import (
	"flag"
	"os"
	"testing"
)

// Test command-line flag parsing
func TestCommandLineFlags(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test version flag
	os.Args = []string{"cwlogs", "--version"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	flagVersion := flag.Bool("version", false, "show version")
	flagProfile := flag.String("profile", "", "AWS profile to use")
	flagRegion := flag.String("region", "", "AWS region to use")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	if !*flagVersion {
		t.Error("Expected version flag to be true")
	}
	
	if *flagProfile != "" {
		t.Errorf("Expected empty profile, got %q", *flagProfile)
	}
	
	if *flagRegion != "" {
		t.Errorf("Expected empty region, got %q", *flagRegion)
	}
}

// Test profile flag parsing
func TestProfileFlag(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test profile flag
	os.Args = []string{"cwlogs", "--profile", "myprofile"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	flagVersion := flag.Bool("version", false, "show version")
	flagProfile := flag.String("profile", "", "AWS profile to use")
	flagRegion := flag.String("region", "", "AWS region to use")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	if *flagVersion {
		t.Error("Expected version flag to be false")
	}
	
	if *flagProfile != "myprofile" {
		t.Errorf("Expected profile 'myprofile', got %q", *flagProfile)
	}
	
	if *flagRegion != "" {
		t.Errorf("Expected empty region, got %q", *flagRegion)
	}
}

// Test help flag parsing
func TestHelpFlag(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test help flag
	os.Args = []string{"cwlogs", "--help"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	flagHelp := flag.Bool("help", false, "show help")
	flagProfile := flag.String("profile", "", "AWS profile to use")
	flagRegion := flag.String("region", "", "AWS region to use")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	if !*flagHelp {
		t.Error("Expected help flag to be true")
	}
	
	if *flagProfile != "" {
		t.Errorf("Expected empty profile, got %q", *flagProfile)
	}
	
	if *flagRegion != "" {
		t.Errorf("Expected empty region, got %q", *flagRegion)
	}
}

// Test no flags (default behavior)
func TestNoFlags(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test no flags
	os.Args = []string{"cwlogs"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	flagVersion := flag.Bool("version", false, "show version")
	flagProfile := flag.String("profile", "", "AWS profile to use")
	flagRegion := flag.String("region", "", "AWS region to use")
	flagHelp := flag.Bool("help", false, "show help")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	if *flagVersion {
		t.Error("Expected version flag to be false")
	}
	
	if *flagHelp {
		t.Error("Expected help flag to be false")
	}
	
	if *flagProfile != "" {
		t.Errorf("Expected empty profile, got %q", *flagProfile)
	}
	
	if *flagRegion != "" {
		t.Errorf("Expected empty region, got %q", *flagRegion)
	}
}

// Test positional argument parsing
func TestPositionalArgument(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test positional argument
	os.Args = []string{"cwlogs", "myprofile"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	flagVersion := flag.Bool("version", false, "show version")
	flagProfile := flag.String("profile", "", "AWS profile to use")
	flagRegion := flag.String("region", "", "AWS region to use")
	flagHelp := flag.Bool("help", false, "show help")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	// Check that flags are not set
	if *flagVersion {
		t.Error("Expected version flag to be false")
	}
	
	if *flagHelp {
		t.Error("Expected help flag to be false")
	}
	
	if *flagProfile != "" {
		t.Errorf("Expected empty profile flag, got %q", *flagProfile)
	}
	
	if *flagRegion != "" {
		t.Errorf("Expected empty region flag, got %q", *flagRegion)
	}
	
	// Check positional argument
	args := flag.CommandLine.Args()
	if len(args) != 1 {
		t.Fatalf("Expected 1 positional argument, got %d", len(args))
	}
	
	if args[0] != "myprofile" {
		t.Errorf("Expected positional argument 'myprofile', got %q", args[0])
	}
}

// Test mixed flag and positional argument (flag takes precedence)
func TestMixedFlagAndPositional(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test flag with positional argument
	os.Args = []string{"cwlogs", "--profile", "flagprofile", "positionalprofile"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	_ = flag.Bool("version", false, "show version")
	flagProfile := flag.String("profile", "", "AWS profile to use")
	_ = flag.String("region", "", "AWS region to use")
	_ = flag.Bool("help", false, "show help")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	// Flag should be set
	if *flagProfile != "flagprofile" {
		t.Errorf("Expected profile flag 'flagprofile', got %q", *flagProfile)
	}
	
	// Positional argument should still be available
	args := flag.CommandLine.Args()
	if len(args) != 1 {
		t.Fatalf("Expected 1 positional argument, got %d", len(args))
	}
	
	if args[0] != "positionalprofile" {
		t.Errorf("Expected positional argument 'positionalprofile', got %q", args[0])
	}
}

// Test region flag parsing
func TestRegionFlag(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test region flag
	os.Args = []string{"cwlogs", "--region", "us-west-2"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	_ = flag.Bool("version", false, "show version")
	_ = flag.String("profile", "", "AWS profile to use")
	flagRegion := flag.String("region", "", "AWS region to use")
	_ = flag.Bool("help", false, "show help")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	if *flagRegion != "us-west-2" {
		t.Errorf("Expected region 'us-west-2', got %q", *flagRegion)
	}
}

// Test profile and region positional arguments
func TestProfileAndRegionPositional(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test profile and region positional arguments
	os.Args = []string{"cwlogs", "myprofile", "us-east-1"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	
	_ = flag.Bool("version", false, "show version")
	flagProfile := flag.String("profile", "", "AWS profile to use")
	flagRegion := flag.String("region", "", "AWS region to use")
	_ = flag.Bool("help", false, "show help")
	
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	// Flags should be empty
	if *flagProfile != "" {
		t.Errorf("Expected empty profile flag, got %q", *flagProfile)
	}
	
	if *flagRegion != "" {
		t.Errorf("Expected empty region flag, got %q", *flagRegion)
	}
	
	// Check positional arguments
	args := flag.CommandLine.Args()
	if len(args) != 2 {
		t.Fatalf("Expected 2 positional arguments, got %d", len(args))
	}
	
	if args[0] != "myprofile" {
		t.Errorf("Expected profile 'myprofile', got %q", args[0])
	}
	
	if args[1] != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %q", args[1])
	}
}
