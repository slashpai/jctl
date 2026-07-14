package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ShellScript returns shell initialization code for enabling tab completion.
func ShellScript(shell, binName, binPath string) (string, error) {
	switch shell {
	case "bash":
		return bashScript(binName, binPath), nil
	case "zsh":
		return zshScript(binName, binPath), nil
	case "fish":
		return fishScript(binName, binPath), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shell)
	}
}

// ShellSetupHint returns instructions for enabling completion in the user's shell.
func ShellSetupHint(shell, binName, subcommand string) (string, error) {
	switch shell {
	case "bash":
		return fmt.Sprintf(
			"Execute the following command to activate tab completion for %s in bash:\n\n  source <(%s %s -c bash)\n\nFor permanent activation, add that line to ~/.bashrc.",
			binName, binName, subcommand,
		), nil
	case "zsh":
		return fmt.Sprintf(
			"Execute the following command to activate tab completion for %s in zsh:\n\n  source <(%s %s -c zsh)\n\nFor permanent activation, add that line to ~/.zshrc.",
			binName, binName, subcommand,
		), nil
	case "fish":
		return fmt.Sprintf(
			"Execute the following command to activate tab completion for %s in fish:\n\n  %s %s -c fish | source\n\nFor permanent activation, add that line to ~/.config/fish/config.fish.",
			binName, binName, subcommand,
		), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shell)
	}
}

// DetectShell returns bash, zsh, or fish based on the SHELL environment variable.
func DetectShell() string {
	switch filepath.Base(os.Getenv("SHELL")) {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	default:
		return ""
	}
}

func bashScript(binName, binPath string) string {
	return fmt.Sprintf("complete -o default -o bashdefault -C %s %s\n", shellQuote(binPath), shellQuote(binName))
}

func zshScript(binName, binPath string) string {
	return fmt.Sprintf(`#compdef %s
autoload -U +X bashcompinit && bashcompinit
complete -o default -o bashdefault -C %s %s
`, binName, shellQuote(binPath), shellQuote(binName))
}

func fishScript(binName, binPath string) string {
	return fmt.Sprintf(`function __complete_%s
    set -lx COMP_LINE (commandline -cp)
    test -z (commandline -ct)
    and set COMP_LINE "$COMP_LINE "
    %s
end
complete -f -c %s -a "(__complete_%s)"
`, binName, shellQuote(binPath), binName, binName)
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if !strings.ContainsAny(value, " \t\n'\"\\$`") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", `'\'"'"''`) + "'"
}
