package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func funcMap(cmd *cobra.Command) template.FuncMap {
	return template.FuncMap{
		"localFlags": func() chan *flag.Flag {
			return visitFlags(cmd.LocalFlags())
		},
		"inheritedFlags": func() chan *flag.Flag {
			return visitFlags(cmd.InheritedFlags())
		},
		"defaultValue": defaultValue,
		"flagName":     flagName,
		"argName":      argName,
	}
}

func argName(cmd *cobra.Command) string {
	switch cmd {
	case createVolumeCmd:
		return "VOLUME_NAME [VOLUME_NAME...]"
	case deleteVolumeCmd,
		controllerPublishVolumeCmd,
		controllerUnpublishVolumeCmd,
		valVolCapsCmd,
		nodePublishVolumeCmd,
		nodeUnpublishVolumeCmd:
		return "VOLUME_ID [VOLUME_ID...]"
	case RootCmd, controllerCmd, identityCmd, nodeCmd:
		return "CMD"
		//case docCmd:
		//	return "DIR"
	}

	return ""
}

func helpFunc(cmd *cobra.Command, args []string) {
	format := helpFormat
	if !cmd.Runnable() && cmd.Flags().Lookup("help").Value.String() == "false" {
		format = usageFormat
	}
	tpl, err := template.New("t").Funcs(funcMap(cmd)).Parse(format)
	if err != nil {
		panic(err)
	}
	if err := tpl.Execute(os.Stdout, cmd); err != nil {
		panic(err)
	}
}

func usageFunc(cmd *cobra.Command) error {
	format := usageFormat
	if cmd.Runnable() {
		format = helpFormat
	}
	tpl, err := template.New("t").Funcs(funcMap(cmd)).Parse(format)
	if err != nil {
		return err
	}
	return tpl.Execute(os.Stdout, cmd)
}

func visitFlags(fs *flag.FlagSet) chan *flag.Flag {
	c := make(chan *flag.Flag)
	go func() {
		fs.VisitAll(func(f *flag.Flag) {
			c <- f
		})
		close(c)
	}()
	return c
}

func defaultValue(f *flag.Flag) string {
	switch f.DefValue {
	case "", "false", "0":
		return ""
	}
	switch f.Value.Type() {
	case "string":
		return fmt.Sprintf("\n\n        (default value %q)", f.DefValue)
	default:
		return fmt.Sprintf("\n\n        (default value %v)", f.DefValue)
	}
}

func flagName(f *flag.Flag) string {
	if v := f.Shorthand; v != "" {
		return fmt.Sprintf("-%s, --%s", v, f.Name)
	}
	return fmt.Sprintf("    --%s", f.Name)
}

func setHelpAndUsage(cmd *cobra.Command) {
	cmd.SilenceErrors = true
	if cmd.Runnable() {
		cmd.SilenceUsage = true
	}
	cmd.SetHelpFunc(helpFunc)
	cmd.SetUsageFunc(usageFunc)
	for _, cmd := range cmd.Commands() {
		setHelpAndUsage(cmd)
	}
}

const usageFormat = `NAME
    {{.Use}} -- {{.Short}}

SYNOPSIS
    {{.CommandPath}} [flags] {{argName .}}{{if .HasAvailableSubCommands}}

AVAILABLE COMMANDS{{range .Commands}}{{if (and .IsAvailableCommand (ne .Name "help"))}}{{printf "\n    %s" .Name}}{{end}}{{end}}{{end}}

Use "{{.CommandPath}} -h,--help" for more information
`

const helpFormat = `NAME
    {{.Use}} -- {{.Short}}

SYNOPSIS
    {{.CommandPath}} [flags] {{argName .}}{{if gt (len .Aliases) 0}}

ALIASES
    {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

AVAILABLE COMMANDS{{range .Commands}}{{if (and .IsAvailableCommand (ne .Name "help"))}}{{printf "\n    %s" .Name}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

OPTIONS{{range localFlags}}{{printf "\n    %s\n        %s %s\n" (flagName .) .Usage (defaultValue .)}}{{end}}{{end}}{{if .HasAvailableInheritedFlags}}
GLOBAL OPTIONS{{range inheritedFlags}}{{printf "\n    %s\n        %s %s\n" (flagName .) .Usage (defaultValue .)}}{{end}}{{end}}
ENVIRONMENT OPTIONS
    X_CSI_DEBUG
        Setting X_CSI_DEBUG=true is the same as:
            --log-level=debug --with-request-logging --with-response-logging

    X_CSI_USER_CREDENTIALS
        This environment variable may be used by RPCs to send user credentials.

        csc does not allow user credentials via command line arguments in order
        to prevent sensitive information from appearing as part of a process
        listing.

        One or more credential pairs may be specified, and either the user name
        or passphrase may be quoted to preserve leading or trailing whitespace:

            user1=pass user2="trailing whitespace " "user 3"=' pass'
`
