# AWS Setup Testing Guide

This guide shows how the CloudWatch Log Viewer handles different AWS credential configurations.

## Scenario 1: Basic `aws configure` Setup

**What happens:**
```bash
# User runs basic AWS setup
aws configure
# AWS Access Key ID: AKIA...
# AWS Secret Access Key: ...
# Default region name: us-east-1
# Default output format: json
```

**Files created:**
```
~/.aws/credentials:
[default]
aws_access_key_id = AKIA...
aws_secret_access_key = ...

~/.aws/config:
[default]
region = us-east-1
output = json
```

**How cwlogs handles it:**
- Detects `default` profile from both files
- Automatically uses it (no profile selection needed)
- Shows: "Using AWS profile: default"

## Scenario 2: Multiple Named Profiles

**What happens:**
```bash
# User has multiple profiles
aws configure --profile work
aws configure --profile personal
```

**Files created:**
```
~/.aws/credentials:
[default]
aws_access_key_id = AKIA...
aws_secret_access_key = ...

[work]
aws_access_key_id = AKIA...
aws_secret_access_key = ...

[personal]
aws_access_key_id = AKIA...
aws_secret_access_key = ...
```

**How cwlogs handles it:**
- Detects all profiles: `default`, `work`, `personal`
- Shows profile selection menu
- User chooses which profile to use

## Scenario 3: AWS SSO Setup

**What happens:**
```bash
# User configures SSO
aws configure sso --profile company-admin
aws sso login --profile company-admin
```

**Files created:**
```
~/.aws/config:
[profile company-admin]
sso_start_url = https://company.awsapps.com/start
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = AdministratorAccess
region = us-east-1
```

**How cwlogs handles it:**
- Detects `company-admin` profile from config
- User selects it from menu
- Uses SSO credentials automatically

## Scenario 4: Environment Variables Only

**What happens:**
```bash
# User sets environment variables
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...
export AWS_DEFAULT_REGION=us-east-1
```

**Files created:**
- No AWS files needed

**How cwlogs handles it:**
- No profiles found in files
- Attempts to load default AWS config
- If successful, uses "default" profile
- Shows: "Using AWS profile: default"

## Scenario 5: No AWS Setup

**What happens:**
- User hasn't run `aws configure`
- No AWS files exist
- No environment variables set

**How cwlogs handles it:**
- No profiles found
- Default config load fails
- Shows clear error message:
  ```
  no AWS profiles found and default configuration failed
  
  Please run 'aws configure' or set up AWS credentials
  ```

## Testing Your Setup

You can test your AWS setup before running cwlogs:

```bash
# Check if AWS CLI is configured
aws sts get-caller-identity

# List available profiles
aws configure list-profiles

# Test CloudWatch access
aws logs describe-log-groups --max-items 5
```

## Improved Error Messages

The application now provides helpful guidance:

- **Missing credentials**: Suggests running `aws configure`
- **Permission issues**: Suggests checking IAM permissions
- **Region issues**: Suggests setting default region
- **SSO expiry**: Shows exact command to re-authenticate

This makes the tool much more user-friendly for developers who just want to quickly view their CloudWatch logs without complex AWS setup.

## Command-Line Usage Examples

### Quick Profile Access

**Skip profile selection when you know which one to use:**
```bash
# Use specific profile directly (short syntax)
./cwlogs production

# Development environment (short syntax)
./cwlogs dev

# AWS SSO profile (short syntax)
./cwlogs company-admin

# Alternative flag syntax (explicit)
./cwlogs --profile production
```

### Automation and Scripting

**Shell scripts for monitoring:**
```bash
#!/bin/bash
# production-monitor.sh
echo "🔍 Starting production log monitoring..."
./cwlogs production
```

**CI/CD integration:**
```bash
# In your deployment script
echo "Checking application logs..."
timeout 30s ./cwlogs staging
```

**Multi-environment monitoring:**
```bash
#!/bin/bash
# monitor-all.sh
echo "Select environment:"
select env in "dev" "staging" "production"; do
    ./cwlogs $env
    break
done
```

### Integration with AWS Tools

**With AWS SSO:**
```bash
# Login and monitor in one go (clean syntax)
aws sso login --profile company-admin && ./cwlogs company-admin
```

**With AWS CLI:**
```bash
# Verify profile works, then monitor (clean syntax)
aws sts get-caller-identity --profile myprofile && ./cwlogs myprofile
```

**Profile validation:**
```bash
# Test if profile has CloudWatch access
aws logs describe-log-groups --profile myprofile --max-items 1
./cwlogs myprofile
```

This command-line interface makes cwlogs much more convenient for power users, automation, and integration with existing AWS workflows.
