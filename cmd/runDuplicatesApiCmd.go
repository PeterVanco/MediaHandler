package cmd

import (
	"MediaHandler/cmd/api"
	"MediaHandler/constants"
	. "MediaHandler/util"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"os"
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
	api.Init(runDuplicatesApiCmd.Flags())
}

func runDuplicatesApi(cmd *cobra.Command, args []string) {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/list", func(c *gin.Context) {
		root, _ := runDuplicatesApiCmd.Flags().GetString("photo-root")
		api.GetList(c, root)
	})

	router.GET("/single", func(c *gin.Context) {
		root, _ := runDuplicatesApiCmd.Flags().GetString("photo-root")
		api.GetSingle(c, root)
	})

	router.POST("/single", func(c *gin.Context) {
		var rootDefinition api.RootDefinition
		err := c.BindJSON(&rootDefinition)
		if err != nil {
			Logger.Panic(err)
		}
		if _, err := os.Stat(rootDefinition.Root); os.IsNotExist(err) {
			Logger.Printf("root %s does not exist", rootDefinition.Root)
			c.Error(err)
		}
		api.GetSingle(c, rootDefinition.Root)
	})

	router.POST("/resolve", api.GetResolve)

	port, _ := runDuplicatesApiCmd.Flags().GetUint16("port")
	router.Run(fmt.Sprintf(":%d", port))
}
