# Claude Code Go SDK Demos

This directory contains two demo implementations showcasing different aspects of the Claude Code Go SDK.

## 🚀 Streaming Demo (Default)

**Location**: `streaming/`  
**Command**: `make demo` or `make demo-streaming`

The **streaming demo** provides real-time visibility into Claude's actions using the SDK's streaming JSON output capabilities. Perfect for:

- **Production applications** that need progress tracking
- **Dashboard interfaces** showing AI agent activity  
- **Learning** how Claude Code tool execution works
- **Best user experience** with live step-by-step updates

### Features
- ✅ Real-time tool execution display
- ✅ Progress indicators and step descriptions
- ✅ Educational commentary about streaming JSON
- ✅ Professional-grade monitoring capabilities

## 📝 Basic Demo (Alternative)

**Location**: `basic/`  
**Command**: `make demo-basic`

The **basic demo** shows simple SDK usage with standard JSON output. Ideal for:

- **Learning SDK fundamentals** 
- **Simple integration patterns**
- **Quick prototyping** without complexity
- **Understanding core concepts** before advanced features

### Features
- ✅ Simple request/response pattern
- ✅ Standard JSON output parsing
- ✅ Minimal complexity
- ✅ Easy to understand and modify

## Quick Start

```bash
# Best experience (streaming demo)
make demo

# Alternative approaches
make demo-basic      # Simple JSON output
make demo-streaming  # Same as 'make demo'
```

## Demo Content

Both demos implement the same core functionality:
- Create a Go program that computes Keccac hashes using `sha3.New256()`
- Use Go's built-in `crypto/sha3` package (this is Keccac, not SHA-256)
- Test with provided files to demonstrate working code
- Showcase Claude Code's file creation and execution capabilities

The difference is in **how you see Claude work** - streaming shows real-time progress, basic shows final results.

---

**💡 Tip**: Start with the streaming demo to see the full power of the Claude Code Go SDK, then explore the basic demo to understand the simpler patterns.