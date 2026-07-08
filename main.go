package main

import (
	"github.com/alecthomas/kong"
	"github.com/slashpai/jctl/cmd"
)

func main() {
	var cli cmd.CLI
	ctx := kong.Parse(&cli,
		kong.Name("jctl"),
		kong.Description("A CLI for managing Jira issues.\n\nGet started:\n  jctl configure            Set your Jira URL, email, and API token\n  jctl issue create          Create a new issue\n  jctl issue view KEY-123    View issue details\n  jctl issue list -p PROJ    List issues in a project"),
		kong.UsageOnError(),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
