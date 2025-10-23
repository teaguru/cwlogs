# AWS Region Support

## Overview

Added comprehensive AWS region support to the CloudWatch Log Viewer, allowing users to override the default region from their AWS profile and access CloudWatch logs in any AWS region.

## Why Region Support Matters

### CloudWatch Logs are Region-Specific
- **Log groups exist in specific regions** - A log group in `us-east-1` won't appear in `us-west-2`
- **Applications deploy to multiple regions** - Need to check logs where your app is running
- **Profile defaults might not match** - Your profile might default to `us-east-1` but your logs are in `eu-west-1`

### Common Use Cases
- **Multi-region deployments** - Check logs across different regions
- **Disaster recovery** - Monitor failover regions
- **Compliance requirements** - Access logs in specific geographic regions
- **Cost optimization** - Logs might be in cheaper regions

## Usage

### Command-Line Syntax

**Positional arguments (recommended):**
```bash
./cwlogs <profile> <region>
```

**Flag syntax (explicit):**
```bash
./cwlogs --profile <profile> --region <region>
```

**Mixed syntax:**
```bash
./cwlogs <profile> --region <region>
./cwlogs --profile <profile> <region>
```

### Examples

**Basic region override:**
```bash
# Use dev profile in us-west-2 region
./cwlogs dev us-west-2

# Use production profile in eu-central-1
./cwlogs production eu-central-1
```

**Flag syntax:**
```bash
# Explicit flag syntax
./cwlogs --profile dev --region us-west-2

# Mixed syntax
./cwlogs dev --region us-west-2
```

**Interactive with region override:**
```bash
# Interactive profile selection, specific region
./cwlogs --region eu-west-1
```

## Real-World Scenarios

### 1. Multi-Region Application Monitoring

```bash
#!/bin/bash
# monitor-all-regions.sh
PROFILE="production"
REGIONS=("us-east-1" "us-west-2" "eu-west-1" "ap-southeast-1")

echo "Select region to monitor:"
select region in "${REGIONS[@]}"; do
    echo "Monitoring $PROFILE in $region..."
    ./cwlogs $PROFILE $region
    break
done
```

### 2. Disaster Recovery Monitoring

```bash
#!/bin/bash
# dr-monitor.sh
PRIMARY_REGION="us-east-1"
DR_REGION="us-west-2"

echo "Checking primary region ($PRIMARY_REGION)..."
timeout 10s ./cwlogs production $PRIMARY_REGION

echo "Checking DR region ($DR_REGION)..."
./cwlogs production $DR_REGION
```

### 3. Development Workflow

```bash
#!/bin/bash
# dev-regions.sh
case $1 in
  "local")   ./cwlogs dev us-east-1 ;;      # Local development
  "staging") ./cwlogs staging eu-west-1 ;;  # European staging
  "prod")    ./cwlogs prod us-west-2 ;;     # Production in west coast
  *)         echo "Usage: $0 {local|staging|prod}" ;;
esac
```

### 4. Compliance and Data Residency

```bash
#!/bin/bash
# compliance-monitor.sh
# Monitor logs in specific regions for compliance

# EU data must stay in EU
./cwlogs eu-prod eu-central-1

# US data in US regions
./cwlogs us-prod us-east-1
```

## Technical Implementation

### AWS Client Configuration

```go
// createCloudWatchClient with optional region override
func createCloudWatchClient(profile string, region ...string) (*cloudwatchlogs.Client, error) {
    ctx := context.Background()
    
    // Start with profile configuration
    configOptions := []func(*config.LoadOptions) error{
        config.WithSharedConfigProfile(profile),
    }
    
    // Override region if provided
    if len(region) > 0 && region[0] != "" {
        configOptions = append(configOptions, config.WithRegion(region[0]))
    }
    
    cfg, err := config.LoadDefaultConfig(ctx, configOptions...)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS configuration for profile '%s': %w", profile, err)
    }
    
    return cloudwatchlogs.NewFromConfig(cfg), nil
}
```

### Argument Parsing Logic

```go
// Parse region from flag or positional argument
if *flagRegion != "" {
    // Use region from --region flag
    region = *flagRegion
    fmt.Printf("Using AWS region: %s (from --region flag)\n", region)
} else if len(flag.Args()) > 1 {
    // Use region from second positional argument
    region = flag.Args()[1]
    fmt.Printf("Using AWS region: %s (from argument)\n", region)
}
```

### Precedence Rules

1. **Flag takes precedence**: `--region us-west-2` overrides positional region
2. **Positional fallback**: Second argument used as region if no flag
3. **Profile default**: Uses region from AWS profile if no override
4. **Error on too many args**: `./cwlogs profile region extra` shows error

## Error Handling

### Invalid Region
```bash
$ ./cwlogs dev invalid-region
Error validating AWS profile/region: failed to load AWS configuration...
```

### Region-Specific Issues
```bash
# No log groups in specified region
$ ./cwlogs prod eu-north-1
No log groups found

# Check if logs exist in that region
$ aws logs describe-log-groups --region eu-north-1 --profile prod
```

### Helpful Error Messages
- Clear indication when region override is active
- Suggestions to check different regions
- Commands to verify log group existence

## Benefits

### For Developers
- **Quick region switching** - `./cwlogs dev us-west-2` vs complex AWS CLI commands
- **No profile editing** - Override region without changing AWS config
- **Intuitive syntax** - Follows natural `profile region` order

### For DevOps/SRE
- **Multi-region monitoring** - Easy to check logs across regions
- **Incident response** - Quickly switch regions during outages
- **Automation friendly** - Script-friendly syntax for region switching

### For Organizations
- **Compliance support** - Access logs in required regions
- **Cost optimization** - Check logs in cheaper regions
- **Global operations** - Monitor applications worldwide

## Testing Coverage

### New Tests
```go
// Test region flag parsing
func TestRegionFlag(t *testing.T)

// Test profile and region positional arguments  
func TestProfileAndRegionPositional(t *testing.T)
```

### Updated Tests
- All existing tests updated to include region flag
- Comprehensive argument parsing validation
- Error handling for too many arguments

## Documentation Updates

### Help Output
```
Usage: ./cwlogs [options] [profile] [region]

Arguments:
  profile               AWS profile name
  region                AWS region (overrides profile default)

Examples:
  ./cwlogs dev us-west-2           # Use 'dev' profile in us-west-2
  ./cwlogs --profile dev --region us-east-1  # Use flags
```

### README.md
- Added region examples to all usage sections
- Updated command-line options table
- Added region-specific troubleshooting

## Common AWS Regions

### US Regions
```bash
./cwlogs prod us-east-1      # N. Virginia (most services)
./cwlogs prod us-east-2      # Ohio
./cwlogs prod us-west-1      # N. California  
./cwlogs prod us-west-2      # Oregon
```

### Europe Regions
```bash
./cwlogs prod eu-west-1      # Ireland
./cwlogs prod eu-west-2      # London
./cwlogs prod eu-central-1   # Frankfurt
./cwlogs prod eu-north-1     # Stockholm
```

### Asia Pacific Regions
```bash
./cwlogs prod ap-southeast-1 # Singapore
./cwlogs prod ap-southeast-2 # Sydney
./cwlogs prod ap-northeast-1 # Tokyo
./cwlogs prod ap-south-1     # Mumbai
```

## Migration Guide

### Before (Profile Region Only)
```bash
# Had to change AWS profile region or use AWS CLI
aws configure set region us-west-2 --profile dev
./cwlogs dev

# Or use AWS CLI directly
aws logs describe-log-groups --region us-west-2 --profile dev
```

### After (Region Override)
```bash
# Simple region override
./cwlogs dev us-west-2

# No profile modification needed
# No complex AWS CLI commands
# Works with any region instantly
```

The region support makes cwlogs a truly global tool, capable of accessing CloudWatch logs in any AWS region without complex configuration changes or AWS CLI gymnastics.
