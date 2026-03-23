package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/okunix/terraform-provider-stash/provider"
)

var (
	debug bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "turn on debug mode")
}

func main() {
	flag.Parse()
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/okunix/stash",
		Debug:   debug,
	}
	err := providerserver.Serve(context.Background(), provider.New(), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
