package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/slashpai/jctl/cmd"
	"github.com/slashpai/jctl/internal/completion"
)

func main() {
	var cli cmd.CLI
	parser, err := kong.New(&cli,
		kong.Name("jctl"),
		kong.Description("A CLI for managing Jira issues.\n\nGet started:\n  jctl configure            Set your Jira URL, email, and API token\n  jctl issue create          Create a new issue\n  jctl issue view KEY-123    View issue details\n  jctl issue list -p PROJ    List issues in a project"),
		kong.UsageOnError(),
	)
	if err != nil {
		panic(err)
	}

	if completion.Handle(parser) {
		parser.Exit(0)
	}

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
