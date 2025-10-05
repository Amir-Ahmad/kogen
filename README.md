## Kogen

Kogen is a kubernetes manifest generator powered by [CUE](https://github.com/cue-lang/cue).

## Why another manifest generator

Kogen combines several software that I believe work best together:

- **Helm** - The vast majority of tools and operators such as cert-manager and external-dns are provided via helm charts. Kogen supports declaratively processing helm charts similar to `helm template`.
- **Kustomize** - Kustomize's patching and transformers are very flexible, but over time you end up with too much yaml and duplication. Kogen supports patching helm charts (or yaml files) with Kustomize, to solve the issue of helm charts not supporting a certain property that you need to configure.
- **Cue** - Cue is a great DSL for config management, but out of the box there is no support for Helm/kustomize. However, it's great for writing manifests for your applications or for home, which Kogen supports via the "Objects" generator.

Kogen also has integration with sops to allow for storing encrypted secrets within your repo, but having them automatically decrypted when generating resources.

## Comparisons to other tools

**Timoni**
- Timoni solves the problem of distributing and consuming cue artifacts, and is an alternative to Helm.
- Timoni optionally does deployment to clusters as well, whereas this is something Kogen delegates to other tools such as ArgoCD.
- Timoni doesn't have support for helm charts, kustomize patching, or sops integration.
