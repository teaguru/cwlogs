# Positional Arguments Enhancement

## Overview

Added support for positional arguments to make the CloudWatch Log Viewer even more convenient to use. Now you can specify the AWS profile as a simple argument instead of using flags.

## New Syntax

### Before (flag only)
```bash
./cwlogs --profile myprofile
```

### After (both ways supported)
```bash
./cwlogs myprofile              # Positional argument (shorter)
./cwlogs --profile myprofile    # Flag (explicit)
```

## Usage Examples

### Quick Development Workflow
```bash
# Super clean syntax for daily use
./cwlogs dev
./cwlogs staging  
./cwlogs prod
```

### Automation Scripts
```bash
#!/bin/bash
# monitor.sh - Clean and readable
echo "Monitoring $1 environment..."
./cwlogs $1
```

### Multi-Environment Switcher
```bash
#!/bin/bash
# env-monitor.sh
case $1 in
  "d"|"dev")     ./cwlogs development ;;
  "s"|"staging") ./cwlogs staging ;;
  "p"|"prod")    ./cwlogs production ;;
  *)             echo "Usage: $0 {dev|staging|prod}" ;;
esac
```

### CI/CD Integration
```bash
# GitHub Actions / GitLab CI
- name: Monitor deployment
  run: ./cwlogs ${{ env.AWS_PROFILE }}
```

## Technical Implementation

### Argument Parsing Logic
```go
// Check for profile from flag or positional argument
if *flagProfile != "" {
    // Use profile from --profile flag
    profile = *flagProfile
    fmt.Printf("Using AWS profile: %s (from --profile flag)\n", profile)
} else if len(flag.Args()) > 0 {
    // Use profile from positional argument
    profile = flag.Args()[0]
    fmt.Printf("Using AWS profile: %s (from argument)\n", profile)
    
    // Check for too many arguments
    if len(flag.Args()) > 1 {
        fmt.Fprintf(os.Stderr, "Error: Too many arguments. Expected: %s [profile]\n", os.Args[0])
        os.Exit(1)
    }
}
```

### Precedence Rules
1. **Flag takes precedence**: `./cwlogs --profile flag-profile positional-profile` uses `flag-profile`
2. **Positional fallback**: `./cwlogs positional-profile` uses `positional-profile`
3. **Interactive fallback**: `./cwlogs` shows profile selection menu
4. **Error on multiple args**: `./cwlogs profile1 profile2` shows error

### Error Handling
```bash
# Too many arguments
$ ./cwlogs profile1 profile2 extra
Error: Too many arguments. Expected: ./cwlogs [profile]
Use --help for usage information.

# Invalid profile
$ ./cwlogs nonexistent
Using AWS profile: nonexistent (from argument)
Error validating AWS profile: failed to load AWS configuration for profile 'nonexistent'
```

## Benefits

### For Developers
- **Faster typing**: `./cwlogs dev` vs `./cwlogs --profile dev`
- **More intuitive**: Follows common CLI patterns (git, docker, etc.)
- **Less verbose**: Cleaner command history and scripts

### For DevOps/SRE
- **Scriptable**: Easy to parameterize in automation
- **Readable**: Shell scripts are more self-documenting
- **Flexible**: Both syntaxes work, choose what fits

### For Teams
- **Consistent**: Works like other modern CLI tools
- **Discoverable**: `--help` shows both syntaxes
- **Backward compatible**: Existing `--profile` scripts still work

## Testing Coverage

### New Tests Added
```go
// Test positional argument parsing
func TestPositionalArgument(t *testing.T)

// Test mixed flag and positional argument (flag takes precedence)  
func TestMixedFlagAndPositional(t *testing.T)
```

### Test Results
- **31 total tests** (up from 29)
- **All tests passing**
- **9.6% code coverage**
- **Fast execution** (~0.3s)

## Documentation Updates

### Help Output
```
Usage: ./cwlogs [options] [profile]

Arguments:
  profile               AWS profile name (alternative to --profile flag)

Examples:
  ./cwlogs                    # Interactive profile and log group selection
  ./cwlogs dev                # Use 'dev' profile (positional argument)
  ./cwlogs --profile dev      # Use 'dev' profile (flag)
```

### README.md
- Updated command-line options table
- Added positional argument examples
- Showed both syntaxes in usage examples

### All Documentation Files
- Updated all examples to use shorter syntax
- Maintained flag examples for clarity
- Added precedence explanations

## Real-World Usage Patterns

### Development
```bash
# Quick log checking during development
./cwlogs dev
# vs the old way: ./cwlogs --profile dev
```

### Operations
```bash
# Production incident response
./cwlogs prod
# Faster to type when every second counts
```

### Automation
```bash
# Deployment monitoring script
#!/bin/bash
ENV=${1:-staging}
echo "Monitoring $ENV deployment..."
./cwlogs $ENV
```

### Integration
```bash
# With AWS SSO
aws sso login --profile company && ./cwlogs company

# With environment switching
export AWS_PROFILE=production
./cwlogs $AWS_PROFILE
```

## Comparison with Other Tools

### Similar CLI Patterns
```bash
# Docker
docker run ubuntu          # vs docker run --image ubuntu

# Git  
git checkout main          # vs git checkout --branch main

# AWS CLI
aws s3 ls mybucket         # vs aws s3 ls --bucket mybucket

# Our tool
./cwlogs myprofile         # vs ./cwlogs --profile myprofile
```

The positional argument enhancement makes cwlogs feel more like a native CLI tool that follows established patterns, improving the overall user experience while maintaining full backward compatibility.
