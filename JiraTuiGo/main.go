package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"jiratui/config"
	"jiratui/jira"
	"jiratui/themes"
	"jiratui/ui"
)

//go:embed CHANGELOG.md
var changelog string

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	repoOwner = "Luunoronti"
	repoName  = "JiraTUI"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("jiratui %s (commit %s, built %s)\n", version, commit, date)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
	}

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save config: %v\n", err)
	}

	themes.Detect()
	themes.Apply(themes.Current())

	var client jira.Client
	if cfg.Conn.BaseURL == "" || cfg.Conn.Email == "" {
		client = jira.NewMockClient()
	} else {
		client = jira.NewRealClient(cfg.Conn.BaseURL, cfg.Conn.Email, config.Unprotect(cfg.Conn.TokenProtected))
	}

	if err := ui.Run(cfg, client, version, repoOwner, repoName, changelog); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
