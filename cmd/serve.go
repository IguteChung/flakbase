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
	flagRule  string
)

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start Flakbase server",
	Args:  cobra.NoArgs,
	Run:   serve,
}

func init() {
	cmdServe.Flags().BoolVarP(&flagRest, "rest", "r", false, "enable Flakbase restful api")
	cmdServe.Flags().StringVarP(&flagHost, "host", "", "localhost:9527", "host name to serve")
	cmdServe.Flags().StringVarP(&flagMongo, "mongo", "m", "", "mongodb url to use")
	cmdServe.Flags().StringVarP(&flagRule, "rule", "", "", "security rule json file")
}

func serve(cmd *cobra.Command, args []string) {
	net.Run(&net.Config{
		Rest:  flagRest,
		Host:  flagHost,
		Rule:  flagRule,
		Mongo: flagMongo,
	})
}
