FROM gcr.io/distroless/base:nonroot

ENV HELM_CACHE_HOME=/tmp/helm-cache
ENV KOGEN_CACHE_DIR=/tmp/kogen-cache

# argocd-repo-server expects the UID to be 999
USER 999

COPY argo-plugin/plugin.yaml /home/argocd/cmp-server/config/plugin.yaml

COPY dist/kogen /kogen

ENTRYPOINT ["/kogen"]
