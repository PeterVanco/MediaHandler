package cmd

import (
	"MediaHandler/cmd/api"
	"MediaHandler/constants"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var runDuplicatesApiCmd *cobra.Command

func init() {

	runDuplicatesApiCmd = &cobra.Command{
		Use:   "run-duplicates-api",
		Short: "runs duplicates API",
		Run:   runDuplicatesApi,
	}

	RootCmd.AddCommand(runDuplicatesApiCmd)
	runDuplicatesApiCmd.Flags().Uint16("port", 8080, "port to start the HTTP server on")
	runDuplicatesApiCmd.Flags().String("photo-root", constants.PhotoPath, "photo path")
}

func runDuplicatesApi(cmd *cobra.Command, args []string) {
	r := gin.Default()

	r.GET("/list", func(c *gin.Context) {
		root, _ := runDuplicatesApiCmd.Flags().GetString("photo-root")
		api.GetList(c, root)
	})

	r.POST("/single", func(c *gin.Context) {
		root, _ := runDuplicatesApiCmd.Flags().GetString("photo-root")
		api.GetSingle(c, root)
	})

	r.POST("/resolve", api.GetResolve)

	port, _ := runDuplicatesApiCmd.Flags().GetUint16("port")
	r.Run(fmt.Sprintf(":%d", port))
}
