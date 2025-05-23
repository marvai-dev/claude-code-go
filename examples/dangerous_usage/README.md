# Dangerous Operations Example

🚨 **WARNING: This example demonstrates security-sensitive Claude Code features** 🚨

This example shows how to properly use the `dangerous` package for legitimate use cases that require bypassing Claude's safety controls.

## Security Requirements

Before running this example, you must:

1. **Set required environment variable**:

   ```bash
   export CLAUDE_ENABLE_DANGEROUS="i-accept-all-risks"
   ```

2. **Ensure development environment**:

   ```bash
   export NODE_ENV="development"  # or unset
   ```

3. **Have Claude CLI installed and available in PATH**

## Running the Example

```bash
# Set security bypass (REQUIRED)
export CLAUDE_ENABLE_DANGEROUS="i-accept-all-risks"
export NODE_ENV="development"

# Build and run
go run main.go
```

## What This Example Demonstrates

### ✅ Safe Operations Shown

- Creating a dangerous client with proper environment checks
- Setting environment variables with security warnings
- Enabling MCP debug mode
- Tracking active security warnings
- Resetting dangerous settings

### ⚠️ Dangerous Operations (Explained but NOT Executed)

- Permission bypass with `BYPASS_ALL_PERMISSIONS()`
- Environment variable injection risks
- Production environment blocking

## Expected Output

```
🚨 Dangerous Claude Operations Example 🚨
This example demonstrates security-sensitive features.

1. Creating dangerous client...
✅ Dangerous client created successfully

2. Setting environment variables...
Setting safe environment variables...
🔧 SET: 3 environment variables configured for Claude process
Setting potentially sensitive variables (will show warnings)...
⚠️  WARNING: Setting potentially sensitive environment variable: DEMO_SECRET
🔧 SET: 1 environment variables configured for Claude process
✅ Environment variables configured

3. Enabling MCP debug mode...
🐛 DEBUG: MCP debugging enabled - sensitive information may be logged
✅ MCP debugging enabled

4. Current security warnings:
⚠️  Environment injection active (4 variables)
⚠️  MCP debug logging enabled

5. Permission bypass example (NOT EXECUTED):
   // This would bypass ALL Claude safety controls:
   // result, err := client.BYPASS_ALL_PERMISSIONS("dangerous prompt", nil)
   // WARNING: Only use with trusted, validated input!

6. Resetting dangerous settings...
🔄 RESET: All dangerous settings cleared
✅ All dangerous settings cleared

🎓 Example completed safely!
Remember: These features should only be used when absolutely necessary
and with proper security review and justification.
```

## Security Documentation Requirements

Any real usage of dangerous operations MUST include:

```go
// SECURITY REVIEW REQUIRED: Using dangerous Claude client
// JUSTIFICATION: [Specific reason why dangerous operations are needed]
// RISK ASSESSMENT: [Analysis of potential security risks]
// MITIGATION: [Specific steps taken to minimize risks]
```

## Production Safety

This example will **automatically fail** if run in production:

```bash
export NODE_ENV="production"
go run main.go
# Output: ❌ This example can only run in development environment
```

## Real-World Use Cases

The dangerous package is intended for:

- **Automated CI/CD pipelines** that need unattended operation
- **Testing frameworks** that require bypassing interactive prompts
- **Development tooling** where user supervision isn't practical
- **Container environments** with controlled, isolated execution

## Never Use For

❌ **User-facing applications**  
❌ **Processing untrusted input**  
❌ **Production web services**  
❌ **Shared development environments**

