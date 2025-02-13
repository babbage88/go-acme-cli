package commands

import (
	"fmt"

	"github.com/urfave/cli/v3"
)

const versionNumber = "v1.0.0"

type GoInfraCli struct {
	Name        string       `json:"name"`
	BaseCommand *cli.Command `json:"baseCmd"`
	Version     string       `json:"version"`
}

type UrFaveCliDocumentationSucks struct {
	Name  string `json:"authorName"`
	Email string `json:"email"`
}

func (author *UrFaveCliDocumentationSucks) String() string {
	return fmt.Sprintf("Name: %s Email: %s", author.Name, author.Email)
}

func (c *GoInfraCli) SubCommands() []*cli.Command {
	return c.BaseCommand.Commands
}

func (c *GoInfraCli) Authors(a *UrFaveCliDocumentationSucks) string {
	return a.String()
}

func (c *GoInfraCli) GetDnsSubCommands(sub *cli.Command) error {
	start := len(c.BaseCommand.Commands)
	c.BaseCommand.Commands = append(c.BaseCommand.Commands, sub)
	if len(c.BaseCommand.Commands) == start {
		err := fmt.Errorf("error adding SubCommands, length %d did not change", start)
		return err
	}
	return nil
}
