package cmd

import (
	"github.com/spf13/cobra"

	"github.com/IguteChung/flakbase/pkg/net"
)

var (
	flagHost  string
	flagRest  bool
	flagPort  string
	flagMongo string
)

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start Flakbase server",
	Args:  cobra.NoArgs,
	Run:   serve,
}

func init() {
	cmdServe.Flags().BoolVarP(&flagRest, "rest", "r", false, "enable Flakbase restful api")
	cmdServe.Flags().StringVarP(&flagPort, "port", "p", ":5000", "port to serve")
	cmdServe.Flags().StringVarP(&flagHost, "host", "", "localhost", "host name to serve")
	cmdServe.Flags().StringVarP(&flagMongo, "mongo", "m", "", "mongodb url to use")
}

func serve(cmd *cobra.Command, args []string) {
	net.Run(&net.Config{
		Rest:  flagRest,
		Port:  flagPort,
		Mongo: flagMongo,
	})
}
