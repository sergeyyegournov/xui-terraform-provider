# Releasing `terraform-provider-xui`

This repository uses GoReleaser + GitHub Actions to publish provider binaries and signed checksums from tags.

## One-time setup

1. Generate a dedicated GPG key pair for releases (or use an existing one):
   - `gpg --full-generate-key`
2. Get key fingerprint:
   - `gpg --list-secret-keys --keyid-format LONG`
3. Export private key in ASCII armor:
   - `gpg --armor --export-secret-keys <KEY_ID>`
4. Add repository secrets in GitHub:
   - `GPG_PRIVATE_KEY`: full armored private key block
   - `GPG_PASSPHRASE`: passphrase for the private key
   - `GPG_FINGERPRINT`: full key fingerprint used for signing

## Terraform Registry setup

1. Ensure repository name follows Terraform provider convention:
   - `terraform-provider-xui`
2. Ensure provider source address in `main.go` uses your real namespace and type:
   - `registry.terraform.io/<namespace>/xui`
3. Publish your public key (or key details) where users can access it.
4. In Terraform Registry, create/publish provider `xui` under your namespace and connect this repo.

## Create a release

1. Commit and push to `main`.
2. Tag a version:
   - `git tag v0.1.0`
   - `git push origin v0.1.0`
3. GitHub Action `release` runs automatically and creates a GitHub release with:
   - platform ZIP archives
   - `terraform-provider-xui_<version>_SHA256SUMS`
   - `terraform-provider-xui_<version>_SHA256SUMS.sig`

## Local dry-run (optional)

- Snapshot build:
  - `goreleaser release --snapshot --clean`

