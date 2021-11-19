package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hofstadter-io/cuetils/cmd/cuetils/flags"
	"github.com/hofstadter-io/cuetils/structural"
)

var insertLong = `insert into file(s) with code (only if not present)`

func InsertRun(code string, globs []string) (err error) {

	// you can safely comment this print out
	// fmt.Println("not implemented")

	results, err := structural.InsertGlobs(code, globs, &flags.RootPflags)
	if err != nil {
		return err
	}

	err = structural.ProcessOutputs(results, &flags.RootPflags)

	return err
}

var InsertCmd = &cobra.Command{

	Use: "insert <code> [files...]",

	Aliases: []string{
		"i",
	},

	Short: "insert into file(s) with code (only if not present)",

	Long: insertLong,

	PreRun: func(cmd *cobra.Command, args []string) {

	},

	Run: func(cmd *cobra.Command, args []string) {
		var err error

		// Argument Parsing

		if 0 >= len(args) {
			fmt.Println("missing required argument: 'code'")
			cmd.Usage()
			os.Exit(1)
		}

		var code string

		if 0 < len(args) {

			code = args[0]

		}

		var globs []string

		if 1 < len(args) {

			globs = args[1:]

		}

		err = InsertRun(code, globs)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	extra := func(cmd *cobra.Command) bool {

		return false
	}

	ohelp := InsertCmd.HelpFunc()
	ousage := InsertCmd.UsageFunc()
	help := func(cmd *cobra.Command, args []string) {
		if extra(cmd) {
			return
		}
		ohelp(cmd, args)
	}
	usage := func(cmd *cobra.Command) error {
		if extra(cmd) {
			return nil
		}
		return ousage(cmd)
	}

	InsertCmd.SetHelpFunc(help)
	InsertCmd.SetUsageFunc(usage)

}
