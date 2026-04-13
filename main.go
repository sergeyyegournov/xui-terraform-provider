//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate --provider-name xui

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/syegournov/xkeen-gen/terraform-provider-xui/internal/provider"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with debug logging")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// Must match the source address Terraform uses for this provider (see examples/debug/main.tf).
		Address: "example.com/xui/xui",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}

var version = "dev"
