.PHONY: docs build
default: build

build:
	go build -o terraform-provider-xui .

# Regenerate docs/ from provider schema (requires terraform in PATH or downloaded by the tool).
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate --provider-name xui
