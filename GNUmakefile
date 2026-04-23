.PHONY: docs build test testacc
default: build

build:
	go build -o terraform-provider-xui .

test:
	go test ./...

# Acceptance tests spin up a real 3x-ui container via testcontainers-go and
# drive the provider end-to-end with terraform. Requires Docker to be running.
testacc:
	TF_ACC=1 go test ./internal/provider -run '^TestAcc' -v -timeout 30m

# Regenerate docs/ from provider schema (requires terraform in PATH or downloaded by the tool).
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate --provider-name xui
