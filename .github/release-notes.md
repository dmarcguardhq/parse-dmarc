{{- /* owner/repo from the checkout remote: https://github.com/o/r in CI, git@github.com:o/r.git locally */ -}}
{{- $slug := trimsuffix (trimprefix (trimprefix .GitURL "https://github.com/") "git@github.com:") ".git" -}}
{{- $repo := printf "https://github.com/%s" $slug -}}
{{- $owner := dir $slug -}}
---

## Install {{ .Tag }}

**Docker**

```bash
docker run -d --name {{ .ProjectName }} -p 8080:8080 \
  -e IMAP_HOST=imap.example.com \
  -e IMAP_USERNAME=dmarc@example.com \
  -e IMAP_PASSWORD='your-app-password' \
  -v {{ .ProjectName }}:/data \
  ghcr.io/{{ $slug }}:{{ .Tag }}
```

Also on Docker Hub as `dmarcguard/{{ .ProjectName }}:{{ .Tag }}`.

**Homebrew**

```bash
brew install {{ $owner }}/tap/{{ .ProjectName }}
```

**Nix**

```bash
nix profile install github:{{ $owner }}/nur#{{ .ProjectName }}   # or: nix run github:{{ $owner }}/nur#{{ .ProjectName }}
```

**Binary** (Linux/macOS; Windows: grab the `.zip` from the assets below)

```bash
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/' -e 's/i[3-6]86/i386/')
curl -fsSL {{ $repo }}/releases/download/{{ .Tag }}/{{ .ProjectName }}_${os}_${arch}.tar.gz | tar xz
```

## Verify

Archives carry [build provenance attestations]({{ $repo }}/attestations) and images are signed with cosign (keyless, GitHub Actions OIDC).

```bash
gh attestation verify {{ .ProjectName }}_${os}_${arch}.tar.gz -R {{ $slug }}

cosign verify ghcr.io/{{ $slug }}:{{ .Tag }} \
  --certificate-identity-regexp '^{{ $repo }}/' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

**Full changelog**: [{{ .PreviousTag }}...{{ .Tag }}]({{ $repo }}/compare/{{ .PreviousTag }}...{{ .Tag }})
