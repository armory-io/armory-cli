package config

import (
	"fmt"
	"github.com/juju/ansiterm"
	"os"
)

func View() {
	c, err := loadConfig(true)
	if err != nil {
		fmt.Println("Unable to read configuration file:", err.Error())
		return
	}

	if len(c.Contexts) == 0 {
		fmt.Println("Non existent or empty configuration file.")
		return
	}

	wt := ansiterm.NewTabWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprint(wt, "Name\tEndpoint\tSecure?\tAuthentication\n")
	bold := ansiterm.Styles(ansiterm.Bold)
	red := ansiterm.Foreground(ansiterm.Red)
	for _, ctx := range c.Contexts {
		if ctx.Name == c.CurrentContext {
			bold.Fprint(wt, ctx.Name)
		} else {
			fmt.Fprint(wt, ctx.Name)
		}
		_, _ = fmt.Fprintf(wt, "\t%s", ctx.Connection.Grpc)
		if ctx.Connection.Insecure {
			red.Fprint(wt, "\tplaintext")
		} else if ctx.Connection.Tls != nil && ctx.Connection.Tls.InsecureSkipVerify {
			red.Fprint(wt, "\tno verify")
		} else {
			fmt.Fprint(wt, "\tyes")
		}
		_, _ = fmt.Fprintf(wt, "\t%s\n", getAuthMethodString(ctx))
	}
	_ = wt.Flush()
}

func getAuthMethodString(ctx Context) string {
	if ctx.Identity.Token != "" {
		return "Token"
	}
	if ctx.Identity.TokenCommand != nil {
		return "Executable"
	}
	if ctx.Identity.Armory.ClientId != "" {
		return "Armory Cloud"
	}
	return "Unknown"
}
