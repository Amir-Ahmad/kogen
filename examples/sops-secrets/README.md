# SOPS Secrets Example

Demonstrates using SOPS-encrypted files to inject secrets into Kubernetes manifests.

## Usage

```bash
export SOPS_AGE_KEY='AGE-SECRET-KEY-1V9ES5TS93UMLDT9QF2MGPG5K6AVVY7CERD9RMMX4L4Y2WMADVXAQV8JLX3'
kogen build ./...
```
