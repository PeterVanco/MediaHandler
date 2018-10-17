package api

import (
	flag "github.com/spf13/pflag"
)

var Flags *flag.FlagSet

func Init(globalFlags *flag.FlagSet) {
	Flags = globalFlags
}
