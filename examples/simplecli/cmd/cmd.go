package cmd

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/component/ginc"
)

func newServiceCtx() sctx.ServiceContext {
	return sctx.NewServiceContext(
		sctx.WithName("simple-gin-http"),
		sctx.WithComponent(ginc.NewGin("gin")),
	)
}

type GINComponent interface {
	GetPort() int
	GetRouter() *gin.Engine
}

var outEnvCmd = &cobra.Command{
	Use:   "outenv",
	Short: "Output all environment variables to std",
	Run: func(cmd *cobra.Command, args []string) {
		newServiceCtx().OutEnv()
	},
}

func Execute() {
	rootCmd := &cobra.Command{}

	// Add sub-command
	rootCmd.AddCommand(outEnvCmd)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
