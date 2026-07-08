package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/slashpai/jctl/internal/config"
	"golang.org/x/term"
)

type ConfigureCmd struct{}

func (c *ConfigureCmd) Run() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Jira Base URL (e.g. https://yourorg.atlassian.net): ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.TrimRight(baseURL, "/")

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("API Token: ")
	tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))

	if baseURL == "" || email == "" || token == "" {
		return fmt.Errorf("all fields are required")
	}

	cfg := &config.Config{
		BaseURL: baseURL,
		Email:   email,
		Token:   token,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	path, _ := config.ConfigFilePath()
	color.Green("✓ Configuration saved to %s", path)
	return nil
}
