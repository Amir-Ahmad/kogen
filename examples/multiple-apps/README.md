# Multiple Apps Example

Manages multiple applications in a Cue module. This uses an "App" abstraction that I've opensourced https://github.com/Amir-Ahmad/cue-k8s-modules/tree/main/app.

This monorepo style is how I personally manage my kubernetes manifests.

## Usage

```bash
kogen build ./...
```

## Structure

```
multiple-apps/
├── app.cue           # Shared definitions and generator logic
├── cue.mod/          # Cue module information
└── apps/
    ├── foo/app.cue   # Deployment
    └── bar/app.cue   # StatefulSet
```
