# The one Sandbox Image every Go Track boots (ADR 0004). Every Workshop under
# workshops/ with a sandbox/ folder is baked in at /root/workshop/<slug>; a
# Track's setup-workshop roots the editor at its own folder.
# Build context is workshops/.
#
#   docker buildx build --platform linux/amd64 -f _sandbox/Dockerfile \
#     -t ghcr.io/temporalio/temporal-devdays-go-sandbox:latest --push .
#
# linux/amd64 is required to push: an arm64 image fails to start on Instruqt with
# an empty log and no error. Build linux/arm64 with --load to test on Apple
# Silicon (scripts/test-track-local.sh).

FROM golang:1.26-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
      curl wget git vim nano less unzip ca-certificates bash-completion jq procps \
    && rm -rf /var/lib/apt/lists/*

ARG TARGETARCH

# --- Temporal CLI -------------------------------------------------------------
# Also what the Checks' tests start their own dev server from, so a Check never
# downloads a server binary.
ARG TEMPORAL_CLI_VERSION=1.10.0
RUN curl -fsSL \
      "https://github.com/temporalio/cli/releases/download/v${TEMPORAL_CLI_VERSION}/temporal_cli_${TEMPORAL_CLI_VERSION}_linux_${TARGETARCH:-amd64}.tar.gz" \
    | tar -xz -C /usr/local/bin temporal \
    && chmod +x /usr/local/bin/temporal

# --- Go tools the editor needs, installed now so it never prompts --------------
RUN go install golang.org/x/tools/gopls@v0.24.0 \
    && go install github.com/go-delve/delve/cmd/dlv@v1.27.2

# --- code-server (VS Code in the browser, the Exercise tab) -------------------
ARG CODE_SERVER_VERSION=4.104.3
RUN curl -fsSL https://code-server.dev/install.sh | sh -s -- --version ${CODE_SERVER_VERSION}

# security.workspace.trust.enabled MUST be false: with it on, code-server shows a
# blank page until the folder is trusted, which reads as a broken tab.
# files.autoSave is load-bearing: the assignments say the editor saves as you type.
RUN mkdir -p /root/.cs-workshop/User && printf '%s\n' \
      '{' \
      '    "security.workspace.trust.enabled": false,' \
      '    "workbench.colorTheme": "Default Dark Modern",' \
      '    "workbench.startupEditor": "none",' \
      '    "workbench.editor.enablePreview": false,' \
      '    "workbench.secondarySideBar.defaultVisibility": "hidden",' \
      '    "chat.disableAIFeatures": true,' \
      '    "editor.fontSize": 14,' \
      '    "editor.minimap.enabled": false,' \
      '    "editor.stickyScroll.enabled": true,' \
      '    "files.autoSave": "afterDelay",' \
      '    "files.autoSaveDelay": 500,' \
      '    "telemetry.telemetryLevel": "off",' \
      '    "explorer.compactFolders": false,' \
      '    "files.exclude": { "**/*.bak": false },' \
      '    "go.toolsManagement.checkForUpdates": "off",' \
      '    "go.toolsManagement.autoUpdate": false,' \
      '    "go.survey.prompt": false,' \
      '    "gopls": { "ui.semanticTokens": true },' \
      '    "errorLens.enabledDiagnosticLevels": ["error", "warning"],' \
      '    "errorLens.messageMaxChars": 0' \
      '}' > /root/.cs-workshop/User/settings.json

RUN code-server --user-data-dir /root/.cs-workshop --install-extension golang.go \
    && (code-server --user-data-dir /root/.cs-workshop --install-extension usernamehw.errorlens \
        || echo "==> WARNING: Error Lens did not install; diagnostics will still underline.")

# --- Check and Skip helpers (ADR 0003) ----------------------------------------
COPY _sandbox/bin/ /usr/local/bin/
COPY _sandbox/workshop-solve/ /tmp/workshop-solve/
RUN cd /tmp/workshop-solve && go build -o /usr/local/bin/workshop-solve . \
    && rm -rf /tmp/workshop-solve && chmod +x /usr/local/bin/workshop-check

# --- Bake every Workshop ------------------------------------------------------
COPY . /tmp/workshops/
RUN mkdir -p /root/workshop && cd /tmp/workshops && for d in */; do \
      slug="${d%/}"; [ "$slug" = "_sandbox" ] && continue; \
      [ -d "$slug/sandbox" ] || continue; \
      mkdir -p "/root/workshop/$slug"; \
      cp -r "$slug/exercise" "$slug/solution" "/root/workshop/$slug/"; \
      if [ -d "$slug/sandbox/bin" ]; then install -m 755 "$slug/sandbox/bin/"* /usr/local/bin/; fi; \
    done && rm -rf /tmp/workshops

# --- Pre-warm and prove -------------------------------------------------------
# Every Solution's tests must pass, against the dev server they start themselves,
# or the build fails. Every Exercise must compile as shipped. Building both now
# fills the module and build caches, so the learner's first `go run` and first
# Check take seconds.
#
# PROVE=0 skips the Solutions' tests (still compiles). Use it only for an amd64
# build under emulation on Apple Silicon; prove the same commit natively first.
ARG PROVE=1
RUN set -e; for w in /root/workshop/*/; do \
      echo "==> $w"; \
      (cd "$w/solution" && go mod download && go vet ./... && go test -count=1 -run XXX_NONE ./... >/dev/null \
        && if [ "$PROVE" = "1" ]; then go test -count=1 ./...; fi); \
      (cd "$w/exercise" && go mod download && go vet ./... && go test -count=1 -run XXX_NONE ./... >/dev/null); \
    done

RUN echo 'export TEMPORAL_ADDRESS="${TEMPORAL_ADDRESS:-127.0.0.1:7233}"' >> /root/.bashrc

WORKDIR /root
CMD ["/bin/bash"]
