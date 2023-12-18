package main

import (
	"github.com/pon-network/open-relay/cmd"
	_ "github.com/pon-network/open-relay/docs"
)

var RelayVersion = "dev"

func main() {
	cmd.RelayVersion = RelayVersion
	cmd.Execute()
}
