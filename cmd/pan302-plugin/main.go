package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jianxcao/pan-302-plugin/pkg/pluginpkg"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 2
	}
	switch args[0] {
	case "build":
		flags := flag.NewFlagSet("build", flag.ContinueOnError)
		output := flags.String("output", "", "output .panplugin path")
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: pan302-plugin build <source-dir> --output <file.panplugin>")
			return 2
		}
		sourceDir := args[1]
		if err := flags.Parse(args[2:]); err != nil {
			return 2
		}
		if flags.NArg() != 0 || *output == "" {
			fmt.Fprintln(os.Stderr, "usage: pan302-plugin build <source-dir> --output <file.panplugin>")
			return 2
		}
		if err := pluginpkg.BuildPackage(sourceDir, *output); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(*output)
		return 0
	case "validate":
		flags := flag.NewFlagSet("validate", flag.ContinueOnError)
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if flags.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "usage: pan302-plugin validate <file.panplugin>")
			return 2
		}
		validated, err := pluginpkg.ValidatePackageFile(context.Background(), flags.Arg(0), pluginpkg.DefaultPackageLimits())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Printf("%s %s\n", validated.Manifest.Name, validated.Manifest.Version)
		return 0
	case "help", "-h", "--help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Println(`pan302-plugin:
  pan302-plugin build <source-dir> --output <file.panplugin>
  pan302-plugin validate <file.panplugin>`)
}
