# rockchip-dra-driver devcontainer

Fedora 44 development container for this repository. Docker-shaped and
host-agnostic: macOS (Docker Desktop / Colima / Podman Desktop) and Universal
Blue Aurora/Bluefin DX (Docker or Podman).

No engine-specific `runArgs` (no Podman `keep-id`, no SELinux flags). UID
alignment uses the portable `updateRemoteUserUID` setting.

## What you get

| Tool | Purpose |
|------|---------|
| `go` 1.26.2 | Toolchain. Pin matches `common.mk`. `GOTOOLCHAIN=local`. |
| `gcc` | `go test -race` (cgo) |
| `golangci-lint` | `make lint` |
| `kubectl`, `helm` | Inspect a running cluster and the chart |
| `kind` | Kind binary. Creating a cluster still needs a host engine (see below). |
| `ssh` | Debug nodes. Host `~/.ssh` and the SSH agent are mounted. |
| `gh`, `jq`, `yq`, `git`, `make` | QoL and the Makefile |
| `vim` | Editor inside the container |

Host state is bind-mounted (not copied): kubeconfig, SSH, gitconfig, and
`gh` config. Do not copy kubeconfigs or private keys into the image or the
repo.

## Prerequisites (host)

1. Container runtime with Dev Containers support:
   - **Aurora/Bluefin DX:** Docker (VS Code default) or Podman
   - **macOS:** Docker Desktop, Colima, or Podman Desktop
2. VS Code / Cursor Dev Containers extension, **or** [`devcontainer` CLI](https://github.com/devcontainers/cli)
3. SSH agent running (`ssh-add -l` works) before creating the container, if
   you will SSH to nodes

Ensure host directories exist (the `initializeCommand` creates them if missing):

```sh
mkdir -p ~/.kube ~/.ssh ~/.config/gh
touch ~/.gitconfig
```

## Open the environment

### VS Code / Cursor

1. Open this repo folder.
2. Command Palette → **Dev Containers: Reopen in Container**.

For Podman instead of Docker, user settings only (not in this repo):

```json
{
  "dev.containers.dockerPath": "podman",
  "dev.containers.dockerComposePath": "podman-compose",
  "dev.containers.dockerSocketPath": "/run/user/1000/podman/podman.sock"
}
```

### CLI

```sh
# from repo root
devcontainer up --workspace-folder .
devcontainer exec --workspace-folder . bash
```

(`brew install devcontainer` if the CLI is missing.)

If your default CLI is Podman:

```sh
# either
devcontainer up --workspace-folder . --docker-path podman
# or shell alias / PATH wrapper:
# alias devcontainer='devcontainer --docker-path podman'
```

Do not hard-code `--docker-path podman` when Docker is the engine. Detect
from the environment (`docker info` vs `podman info`, or `DOCKER_HOST`
pointing at a podman sock) or from an existing `devcontainer` alias.

`DOCKER_HOST=unix://…/podman.sock` with the **docker** binary is fine for this
config (no Podman-only flags). Rootful Docker Desktop needs no extra flags.

## Cluster and node access

`~/.kube` is mounted at `/home/vscode/.kube`, so `kubectl` uses the host
kubeconfig (including the current context). `~/.ssh` is mounted for
`config` / `known_hosts` / keys, and `${SSH_AUTH_SOCK}` is bind-mounted to
`/ssh-agent.sock`.

- Start the agent and add keys **before** creating or rebuilding the container.
- If `SSH_AUTH_SOCK` is unset, container create may fail on the agent mount.
  Export it or start the desktop keyring session first.
- Prefer the agent over copying private keys into the image.

## Kind and image builds

The host container engine socket is **not** mounted. Docker and Podman socket
paths differ, and a mounted socket is root-equivalent on the host.

`kind create cluster` and `make setup-e2e` / image builds therefore do not
work inside this container. Run those on the host engine or in CI (kind e2e
with mock devices). Do not install Go on the host to work around that.

## Day-to-day commands

```sh
make cmds
make test
make lint
make vet

kubectl config current-context
kubectl get resourceclaims -A
ssh <node>
```

## Updating tool versions

CLI pins live as `ARG *_VERSION` in [`Dockerfile`](./Dockerfile).
`GO_VERSION` must stay aligned with `common.mk`. `KUBECTL_VERSION` should
stay aligned with the `k8s.io/*` modules in `go.mod`.

Rebuild the container after a pin bump.

## Layout

```
.devcontainer/
  devcontainer.json
  Dockerfile
  README.md
  scripts/
    post-create.sh   # tool smoke checks
    post-start.sh    # SSH and kubeconfig warnings
```
