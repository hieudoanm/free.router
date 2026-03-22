package cmd

import (
	"context"
	"fmt"

	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/hieudoanm/free.router/src/libs/config"
	"github.com/hieudoanm/free.router/src/libs/openrouter"
	"github.com/hieudoanm/free.router/src/libs/proxy"
	"github.com/spf13/cobra"
)

var (
	runPort string
	runKey  string
)

var runCmd = &cobra.Command{
	Use:   "run <model>",
	Short: "Run the run operation for the free.router app",
	Long:  `The run command is a specific utility to execute operations related to run within the free.router application.

As a component of the ai tools, this command empowers you to interact directly with free.router's run features via the CLI.`,
	Args: cobra.ExactArgs(1),
	RunE: runRun,
}

func init() {
	runCmd.Flags().StringVarP(&runPort, "port", "p", "11434", "Port to listen on")
	runCmd.Flags().StringVarP(&runKey, "key", "k", "", "OpenRouter API key (overrides OPENROUTER_API_KEY / ~/.freerouter)")
}

func runRun(cmd *cobra.Command, args []string) error {
	modelArg := args[0]

	// ── 1. Resolve API key ───────────────────────────────────────────────
	apiKey := runKey
	if apiKey == "" {
		apiKey = config.LoadAPIKey()
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, color.RedString("✖ No OpenRouter API key found."))
		fmt.Fprintln(os.Stderr, color.New(color.Faint).Sprint(
			"  Set it via:\n"+
				"    • env var:  export OPENROUTER_API_KEY=sk-or-...\n"+
				"    • file:     echo 'OPENROUTER_API_KEY=sk-or-...' > ~/.freerouter\n"+
				"    • flag:     freerouter run <model> --key sk-or-..."))
		os.Exit(1)
	}

	// ── 2. Fetch free models & resolve ───────────────────────────────────
	fmt.Fprint(os.Stderr, color.CyanString("⠿ Resolving model \"%s\"...\n", modelArg))

	freeModels, err := openrouter.FetchFreeModels()
	if err != nil {
		return fmt.Errorf("could not fetch model list: %w", err)
	}

	model := openrouter.ResolveModel(modelArg, freeModels)
	if model == nil {
		fmt.Fprint(os.Stderr, color.RedString("✖ Model %q not found in free models.\n", modelArg))
		fmt.Fprintln(os.Stderr, color.New(color.Faint).Sprint("  Run `freerouter models` to see available models."))
		os.Exit(1)
	}

	color.Green("✔ Model resolved: %s", color.New(color.FgWhite, color.Bold).Sprint(model.ID))

	// ── 3. Start proxy server ────────────────────────────────────────────
	addr := "127.0.0.1:" + runPort
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot bind to %s: %w", addr, err)
	}

	alias := proxy.AliasID(model.ID)
	handler := proxy.NewHandler(model, apiKey)
	srv := &http.Server{Handler: handler}

	bold := color.New(color.Bold)
	white := color.New(color.FgWhite)
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	bold.Println("  🟢 freerouter is running!")
	fmt.Println()
	white.Print("  Model  ")
	cyan.Printf("%s", model.ID)
	dim.Printf("  (alias: %s)\n", alias)
	white.Print("  URL    ")
	cyan.Printf("http://localhost:%s\n", runPort)
	if model.ContextLength > 0 {
		white.Print("  Ctx    ")
		dim.Printf("%s tokens\n", formatCtxFull(model.ContextLength))
	}

	fmt.Println()
	bold.Println("  ── Add to Cursor ──────────────────────────────────────")
	dim.Println("  Cursor → Settings → Models → Add Custom Model:")
	fmt.Println()
	white.Print("    Base URL : ")
	yellow.Printf("http://localhost:%s/v1\n", runPort)
	white.Print("    Model    : ")
	yellow.Println(alias)
	dim.Printf("    (full ID : %s)\n", model.ID)
	white.Print("    API Key  : ")
	yellow.Println("freerouter")

	fmt.Println()
	bold.Println("  ── Quick test ─────────────────────────────────────────")
	dim.Printf("  curl http://localhost:%s/v1/chat/completions \\\n", runPort)
	dim.Printf("    -H \"Content-Type: application/json\" \\\n")
	dim.Printf("    -d '{\"model\":\"%s\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello!\"}]}'\n", alias)
	fmt.Println()
	dim.Println("  Press Ctrl+C to stop.")

	// ── 4. Graceful shutdown on SIGINT / SIGTERM ─────────────────────────
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println(color.YellowString("\n  Shutting down freerouter..."))
		_ = srv.Shutdown(context.Background())
	}()

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}

	color.New(color.Faint).Println("  Server closed. Bye!")
	return nil
}

func formatCtxFull(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%dM", n/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%dk", n/1_000)
	}
	return fmt.Sprintf("%d", n)
}
