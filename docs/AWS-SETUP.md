# AWS Setup Guide

## Prerequisites

- AWS CLI installed: `aws --version`
- AWS account with CloudWatch Logs access
- Terminal access

## Quick Setup

### Option 1: Basic Configuration (Recommended)
```bash
aws configure
```
Enter when prompted:
- **Access Key ID**: Your AWS access key
- **Secret Access Key**: Your AWS secret key  
- **Default region**: e.g., `us-east-1`
- **Output format**: `json` (recommended)

### Option 2: Named Profiles
```bash
# Create multiple profiles for different environments
aws configure --profile dev
aws configure --profile staging  
aws configure --profile production
```

### Option 3: AWS SSO
```bash
# Configure SSO profile
aws configure sso --profile company-admin
aws sso login --profile company-admin
```

### Option 4: Environment Variables
```bash
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export AWS_DEFAULT_REGION=us-east-1
```

## Required Permissions

Your AWS user/role needs these CloudWatch permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:DescribeLogGroups",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
```

## Verification

Test your setup:
```bash
# Check credentials
aws sts get-caller-identity

# List available profiles  
aws configure list-profiles

# Test CloudWatch access
aws logs describe-log-groups --max-items 5

# Test specific profile
aws logs describe-log-groups --profile production --max-items 5
```

## Usage with cwlogs

### Automatic Profile Detection
```bash
# cwlogs automatically detects and uses available profiles
./cwlogs
```

### Specify Profile
```bash
# Use specific profile
./cwlogs production

# Use profile with region
./cwlogs production us-west-2

# Use flags
./cwlogs --profile dev --region eu-west-1
```

## Troubleshooting

### Common Issues

**"No AWS profiles found"**
```bash
# Solution: Configure AWS CLI
aws configure
```

**"Access denied"**
```bash
# Check credentials
aws sts get-caller-identity

# Check CloudWatch permissions
aws logs describe-log-groups
```

**"SSO session expired"**
```bash
# Re-authenticate
aws sso login --profile your-profile
```

**"No log groups found"**
```bash
# Check region
aws configure get region

# List log groups in specific region
aws logs describe-log-groups --region us-west-2
```

### Profile Management

**List profiles:**
```bash
aws configure list-profiles
```

**Switch default profile:**
```bash
export AWS_PROFILE=production
```

**Check current configuration:**
```bash
aws configure list
```

## Best Practices

### Security
- Use IAM roles instead of access keys when possible
- Rotate access keys regularly
- Use least-privilege permissions
- Never commit credentials to code

### Organization
- Use descriptive profile names: `company-dev`, `company-prod`
- Group profiles by environment or account
- Document profile purposes for team members

### Regions
- Set appropriate default regions for each profile
- Use region override in cwlogs when needed: `./cwlogs profile region`
- Consider data residency and compliance requirements
