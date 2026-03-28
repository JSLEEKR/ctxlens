package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/JSLEEKR/ctxlens/internal/analyzer"
	"github.com/JSLEEKR/ctxlens/internal/display"
	"github.com/JSLEEKR/ctxlens/internal/parser"
	"github.com/JSLEEKR/ctxlens/internal/pricing"
	"github.com/JSLEEKR/ctxlens/internal/tokenizer"
)

// Version is set at build time.
var Version = "0.1.0"

// MaxInputSize is the maximum input size in bytes (50MB).
const MaxInputSize = 50 * 1024 * 1024

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "ctxlens",
		Short: "LLM context window profiler",
		Long:  "ctxlens decomposes LLM context windows into categories, counts tokens per segment, and estimates costs.",
	}

	root.AddCommand(analyzeCmd())
	root.AddCommand(providersCmd())
	root.AddCommand(versionCmd())
	root.AddCommand(topCmd())

	return root
}

func analyzeCmd() *cobra.Command {
	var formatFlag string
	var providerFlag string
	var modelFlag string
	var limitFlag int

	cmd := &cobra.Command{
		Use:   "analyze [file]",
		Short: "Decompose a conversation payload",
		Long:  "Analyze an LLM conversation payload, showing token breakdown by category and estimated costs.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var data []byte
			var source string
			var err error

			if len(args) == 0 || args[0] == "-" {
				// Read from stdin
				data, err = readStdin()
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				source = "stdin"
			} else {
				data, err = readFile(args[0])
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
				source = args[0]
			}

			if len(data) == 0 {
				return fmt.Errorf("empty input")
			}

			if len(data) > MaxInputSize {
				return fmt.Errorf("input exceeds 50MB limit (%d bytes)", len(data))
			}

			// Parse
			conv, err := parser.Parse(data)
			if err != nil {
				return fmt.Errorf("parsing: %w", err)
			}

			// Override provider/model if specified
			if providerFlag != "" {
				conv.Provider = providerFlag
			}
			if modelFlag != "" {
				conv.Model = modelFlag
			}

			// Load config
			cfg, _ := pricing.LoadConfig("")
			pricer := pricing.NewWithConfig(cfg)

			// Analyze
			tok := tokenizer.New()
			a := analyzer.New(tok, pricer)
			result := a.Analyze(conv, source)

			if limitFlag > 0 {
				result.ModelLimit = limitFlag
			}

			// Display
			switch formatFlag {
			case "json":
				display.RenderJSON(os.Stdout, result)
			case "flame":
				display.RenderFlame(os.Stdout, result)
			default:
				display.RenderTable(os.Stdout, result)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&formatFlag, "format", "f", "table", "Output format: table, json, flame")
	cmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "Override provider (anthropic, openai)")
	cmd.Flags().StringVarP(&modelFlag, "model", "m", "", "Override model name")
	cmd.Flags().IntVarP(&limitFlag, "limit", "l", 0, "Context window limit (tokens)")

	return cmd
}

func topCmd() *cobra.Command {
	var byFlag string
	var nFlag int

	cmd := &cobra.Command{
		Use:   "top [dir]",
		Short: "Show biggest token consumers",
		Long:  "Scan a directory of conversation files and show the biggest token consumers.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			entries, err := os.ReadDir(dir)
			if err != nil {
				return fmt.Errorf("reading directory: %w", err)
			}

			tok := tokenizer.New()
			cfg, _ := pricing.LoadConfig("")
			pricer := pricing.NewWithConfig(cfg)
			a := analyzer.New(tok, pricer)

			type fileResult struct {
				source string
				result *analyzer.AnalysisResult
			}
			var results []fileResult

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				path := dir + "/" + entry.Name()
				data, err := readFile(path)
				if err != nil || len(data) == 0 {
					continue
				}
				conv, err := parser.Parse(data)
				if err != nil {
					continue
				}
				r := a.Analyze(conv, entry.Name())
				results = append(results, fileResult{source: entry.Name(), result: r})
			}

			if len(results) == 0 {
				fmt.Println("No valid conversation files found.")
				return nil
			}

			// Sort
			sortByTokens := byFlag != "cost"
			if sortByTokens {
				sortResults(results, func(a, b fileResult) bool {
					return a.result.TotalTokens > b.result.TotalTokens
				})
			} else {
				sortResults(results, func(a, b fileResult) bool {
					return a.result.EstimatedCost > b.result.EstimatedCost
				})
			}

			// Display top N
			if nFlag > len(results) {
				nFlag = len(results)
			}

			fmt.Printf("\nTop %d Conversations by %s\n", nFlag, byFlag)
			fmt.Println("=" + fmt.Sprintf("%*s", 40, "")[1:])
			fmt.Println()
			fmt.Printf("%-5s %-30s %10s %10s\n", "Rank", "File", "Tokens", "Cost")
			fmt.Printf("%-5s %-30s %10s %10s\n", "----", "-----", "------", "----")
			for i := 0; i < nFlag; i++ {
				r := results[i]
				fmt.Printf("%-5d %-30s %10s $%9.3f\n",
					i+1, truncateName(r.source, 30),
					display.FormatNumber(r.result.TotalTokens),
					r.result.EstimatedCost,
				)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringVar(&byFlag, "by", "tokens", "Sort by: tokens or cost")
	cmd.Flags().IntVarP(&nFlag, "n", "n", 10, "Number of results to show")

	return cmd
}

func sortResults[T any](s []T, less func(a, b T) bool) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if less(s[j], s[i]) {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func truncateName(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func providersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "providers",
		Short: "List supported providers and pricing",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := pricing.LoadConfig("")
			pricer := pricing.NewWithConfig(cfg)

			fmt.Println("\nSupported Providers and Pricing")
			fmt.Println("================================")
			fmt.Println()

			for _, provider := range pricer.GetProviders() {
				fmt.Printf("  %s:\n", provider)
				models := pricer.GetModels(provider)
				for model, p := range models {
					fmt.Printf("    %-25s Input: $%.2f/1M  Output: $%.2f/1M\n",
						model, p.Input, p.Output)
				}
				fmt.Println()
			}
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ctxlens v%s\n", Version)
		},
	}
}

func readStdin() ([]byte, error) {
	return io.ReadAll(io.LimitReader(os.Stdin, MaxInputSize+1))
}

func readFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxInputSize {
		return nil, fmt.Errorf("file exceeds 50MB limit (%d bytes)", info.Size())
	}
	return os.ReadFile(path)
}
