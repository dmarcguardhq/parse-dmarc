---

## Install {{ .Tag }}

**Docker**

```bash
docker run -d --name parse-dmarc -p 8080:8080 \
  -e IMAP_HOST=imap.example.com \
  -e IMAP_USERNAME=dmarc@example.com \
  -e IMAP_PASSWORD='your-app-password' \
  -v parse-dmarc:/data \
  ghcr.io/dmarcguardhq/parse-dmarc:{{ .Tag }}
```

Also on Docker Hub as `dmarcguard/parse-dmarc:{{ .Tag }}`.

**Homebrew**

```bash
brew install dmarcguardhq/tap/parse-dmarc
```

**Binary** (swap `linux_amd64` for your platform; see the assets below)

```bash
curl -fsSL https://github.com/dmarcguardhq/parse-dmarc/releases/download/{{ .Tag }}/parse-dmarc_linux_amd64.tar.gz | tar xz
```

## Verify

Archives carry [build provenance attestations](https://github.com/dmarcguardhq/parse-dmarc/attestations) and images are signed with cosign (keyless, GitHub Actions OIDC).

```bash
gh attestation verify parse-dmarc_linux_amd64.tar.gz -R dmarcguardhq/parse-dmarc

cosign verify ghcr.io/dmarcguardhq/parse-dmarc:{{ .Tag }} \
  --certificate-identity-regexp '^https://github.com/dmarcguardhq/parse-dmarc/' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

**Full changelog**: [{{ .PreviousTag }}...{{ .Tag }}](https://github.com/dmarcguardhq/parse-dmarc/compare/{{ .PreviousTag }}...{{ .Tag }})
