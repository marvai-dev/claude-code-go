# Dangerous Package - Security-Sensitive Claude Code Features

🚨 **WARNING: This package contains security-sensitive operations that can bypass Claude's safety controls** 🚨

This package provides access to Claude Code CLI features that were intentionally omitted from the main SDK due to security concerns. These features should only be used in controlled environments with proper security review.

## Features

### 🔓 Permission Bypass

```go
// COMPLETELY DISABLES Claude's safety mechanisms
result, err := client.BYPASS_ALL_PERMISSIONS("prompt", opts)
```

### 🌍 Environment Variable Injection

```go
// Can expose sensitive data to Claude process
err := client.SET_ENVIRONMENT_VARIABLES(map[string]string{
    "API_KEY": "secret",  // ⚠️ Security risk!
})
```

### 🐛 MCP Debug Mode

```go
// May expose sensitive information in logs
err := client.ENABLE_MCP_DEBUG()
```

## Security Requirements

### 🛡️ Multi-Layer Protection

1. **Environment Variable Gate**:

   ```bash
   export CLAUDE_ENABLE_DANGEROUS="i-accept-all-risks"
   ```

2. **Production Environment Blocking**:

   - Automatically fails if `NODE_ENV=production`
   - Automatically fails if `GO_ENV=production`
   - Automatically fails if `ENVIRONMENT=production` or `prod`

3. **Explicit Security Documentation**:

   ```go
   // SECURITY REVIEW REQUIRED: Using dangerous Claude client
   // JUSTIFICATION: [Why dangerous operations are needed]
   // RISK ASSESSMENT: [Analysis of potential security risks]
   // MITIGATION: [Steps taken to minimize risks]
   ```

## Usage

### Basic Setup

```go
package main

import (
    "github.com/marvai-dev/claude-code-go/pkg/claude/dangerous"
)

func main() {
    // SECURITY REVIEW REQUIRED: Using dangerous operations
    // JUSTIFICATION: Automated testing requires permission bypass
    // RISK ASSESSMENT: Isolated test environment, controlled input
    // MITIGATION: Output logged, environment validated

    client, err := dangerous.NewDangerousClient("claude")
    if err != nil {
        // Will fail unless security requirements are met
        log.Fatal(err)
    }

    // Use dangerous operations...
}
```

### Environment Variable Requirements

```bash
# Required for any dangerous operations
export CLAUDE_ENABLE_DANGEROUS="i-accept-all-risks"

# Must NOT be in production
export NODE_ENV="development"  # or unset

# Then run your program
go run main.go
```

## Security Warnings

### 🚨 Permission Bypass Risks

`BYPASS_ALL_PERMISSIONS()` completely disables Claude's safety mechanisms:

- ✅ **Safe for**: Controlled automation, validated inputs, isolated environments
- ❌ **Never use for**: User-facing apps, untrusted input, shared environments

### ⚠️ Environment Variable Injection Risks

`SET_ENVIRONMENT_VARIABLES()` can expose sensitive data:

- **API keys and tokens** become accessible to Claude
- **System paths** can be hijacked for malicious purposes
- **Configuration** may be altered unexpectedly

### 🐛 MCP Debug Risks

`ENABLE_MCP_DEBUG()` may expose sensitive information:

- **Debug logs** can contain secrets or internal data
- **Performance impact** from verbose logging
- **Log files** may persist sensitive information

## Examples

### Automated Deployment

```go
func deployApplication() error {
    // SECURITY REVIEW REQUIRED: Deployment automation
    // JUSTIFICATION: CI/CD pipeline requires unattended operation
    // RISK ASSESSMENT: Container isolation, network restrictions
    // MITIGATION: Input validation, audit logging, limited scope

    client, err := dangerous.NewDangerousClient("claude")
    if err != nil {
        return err
    }

    // Configure deployment environment
    deployEnv := map[string]string{
        "DEPLOYMENT_MODE": "automated",
        "LOG_LEVEL":       "info",
    }
    if err := client.SET_ENVIRONMENT_VARIABLES(deployEnv); err != nil {
        return err
    }

    // Execute deployment with bypassed permissions
    return client.BYPASS_ALL_PERMISSIONS("Deploy the application", &claude.RunOptions{
        Format:   claude.JSONOutput,
        MaxTurns: 5,
    })
}
```

### Testing Framework Integration

```go
func runTestSuite() error {
    // SECURITY REVIEW REQUIRED: Test automation
    // JUSTIFICATION: Test suite requires bypassing interactive prompts
    // RISK ASSESSMENT: Test environment only, controlled test data
    // MITIGATION: Test isolation, no sensitive data, output validation

    client, err := dangerous.NewDangerousClient("claude")
    if err != nil {
        return err
    }

    // Enable debug logging for test troubleshooting
    if err := client.ENABLE_MCP_DEBUG(); err != nil {
        return err
    }

    // Run tests without permission prompts
    result, err := client.BYPASS_ALL_PERMISSIONS("Run the test suite", nil)
    if err != nil {
        return err
    }

    // Validate test results
    return validateTestOutput(result.Result)
}
```

## Runtime Warnings

The package provides automatic warnings:

```
🚨 WARNING: Executing Claude with ALL PERMISSIONS BYPASSED 🚨
This removes all safety controls and allows unrestricted access.

⚠️  WARNING: Setting potentially sensitive environment variable: API_KEY
⚠️  WARNING: Modifying PATH environment variable can affect executable resolution

🐛 DEBUG: MCP debugging enabled - sensitive information may be logged

🔧 SET: 3 environment variables configured for Claude process
```

## Security Best Practices

### ✅ Safe Usage Patterns

1. **Isolated Environments**: Use only in containers, VMs, or test environments
2. **Validated Input**: Never pass untrusted user input to dangerous operations
3. **Audit Logging**: Log all dangerous operations for security review
4. **Limited Scope**: Use minimal permissions and timeboxed operations
5. **Environment Checks**: Validate environment before dangerous operations

### ❌ Unsafe Usage Patterns

1. **Production Deployment**: Never deploy dangerous operations to production
2. **User-Facing Apps**: Never expose dangerous operations to end users
3. **Shared Systems**: Avoid on multi-tenant or shared development systems
4. **Unvalidated Input**: Never pass untrusted data to dangerous operations
5. **Persistent Changes**: Avoid operations that make lasting system changes

## Development Guidelines

### Code Review Requirements

All code using dangerous operations MUST include:

```go
// SECURITY REVIEW REQUIRED: Brief description
// JUSTIFICATION: Specific reason why dangerous operations are needed
// RISK ASSESSMENT: Analysis of potential security impacts
// MITIGATION: Specific steps taken to reduce risk
```

### Testing

```bash
# Run dangerous package tests
go test ./pkg/claude/dangerous -v

# Test with required environment
CLAUDE_ENABLE_DANGEROUS="i-accept-all-risks" NODE_ENV="development" go test ./pkg/claude/dangerous -v

# Verify production blocking works
NODE_ENV="production" go test ./pkg/claude/dangerous -v  # Should fail
```

## Related Documentation

- [SECURITY_SENSITIVE_FEATURES.md](../../../ai_docs/SECURITY_SENSITIVE_FEATURES.md) - Implementation details
- [Main README](../../../README.md) - Overall SDK documentation
- [Official Claude Code Docs](https://docs.anthropic.com/en/docs/claude-code) - Claude Code CLI reference

## Support

For security questions or concerns about this package:

1. Review the implementation guide in `ai_docs/SECURITY_SENSITIVE_FEATURES.md`
2. Check existing test cases for usage patterns
3. Ensure your use case includes proper security review documentation

**Remember**: These features exist for legitimate automation needs, but require careful consideration of security implications.
