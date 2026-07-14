package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/slashpai/jctl/internal/completion"
)

type CompletionCmd struct {
	Shell string `arg:"" optional:"" enum:"bash,zsh,fish," default:"" help:"Shell to generate completion for (default: detect from $SHELL)."`
	Code  bool   `short:"c" help:"Print shell initialization code only."`
}

func (c *CompletionCmd) Run(ctx *kong.Context) error {
	shell := c.Shell
	if shell == "" {
		shell = completion.DetectShell()
	}
	if shell == "" {
		return fmt.Errorf("could not detect shell — pass one of: bash, zsh, fish")
	}

	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable path: %w", err)
	}
	binPath, err = filepath.Abs(binPath)
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	if c.Code {
		script, err := completion.ShellScript(shell, ctx.Model.Name, binPath)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(ctx.Stdout, script)
		return err
	}

	hint, err := completion.ShellSetupHint(shell, ctx.Model.Name, "completion")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(ctx.Stdout, hint)
	return err
}
