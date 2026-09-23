package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/ladev707/3dbuilder-backend/internal/api"
	"github.com/ladev707/3dbuilder-backend/internal/config"
)

const usage = `usage: console <command>

commands:
  routes:show  list registered HTTP routes
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch command := os.Args[1]; command {
	case "routes:show":
		gin.SetMode(gin.ReleaseMode)

		if err := api.WriteRoutes(os.Stdout, api.NewEngine(config.Config{})); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", command, err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", command, usage)
		os.Exit(2)
	}
}
