package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/edsrzf/mmap-go"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Hello, World!")

	// Example for github.com/edsrzf/mmap-go
	fmt.Println("\n--- mmap-go example ---")
	mmapExample()

	// Example for github.com/olekukonko/tablewriter
	fmt.Println("\n--- tablewriter example ---")
	tablewriterExample()

	// Example for github.com/urfave/cli/v3
	// Note: cli/v3 apps are typically run by os.Args,
	// so this example shows how to define a simple app.
	// To run a specific command, you would build and run the executable
	// with arguments, e.g., ./your_app_name mycommand
	fmt.Println("\n--- urfave/cli/v3 example ---")
	cliApp := cliExample()
	// We'll just print the help text here for demonstration.
	// In a real app, you'd call cliApp.Run(context.Background(), os.Args)
	fmt.Println("To run the CLI app, build and execute with commands like 'greet --name Roo'")
	_ = cliApp.Run(context.Background(), []string{os.Args[0], "--help"}) // Show help for demonstration

	// Example for go.uber.org/zap
	fmt.Println("\n--- zap logger example ---")
	zapExample()

	fmt.Println("\nAll examples executed.")
}

func mmapExample() {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "example.mmap")
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up

	// Write some data to the file
	data := []byte("Hello from mmap!")
	if _, err := tmpfile.Write(data); err != nil {
		log.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatalf("Failed to close temp file: %v", err)
	}

	// Open the file for mmap
	f, err := os.OpenFile(tmpfile.Name(), os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Failed to open file for mmap: %v", err)
	}
	defer f.Close()

	// Memory-map the file
	mmapData, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		log.Fatalf("Failed to mmap file: %v", err)
	}
	defer mmapData.Unmap()

	// Read from mmap
	fmt.Printf("Read from mmap: %s\n", mmapData[:len(data)])

	// Write to mmap
	copy(mmapData[len(data):], []byte(" Appended!"))
	mmapData.Flush() // Ensure changes are written to disk

	// Re-read to verify
	updatedContent := make([]byte, len(data)+len(" Appended!"))
	fileRead, err := os.Open(tmpfile.Name())
	if err != nil {
		log.Fatalf("Failed to open file for re-read: %v", err)
	}
	defer fileRead.Close()
	_, err = fileRead.Read(updatedContent)
	if err != nil {
		log.Fatalf("Failed to read updated content: %v", err)
	}
	fmt.Printf("Read from file after mmap write: %s\n", updatedContent)
}

func tablewriterExample() {
	data := [][]string{
		[]string{"A", "The Good", "500"},
		[]string{"B", "The Very very Bad Man", "288"},
		[]string{"C", "The Ugly", "120"},
		[]string{"D", "The Gopher", "800"},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Sign", "Rating"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
}

func cliExample() *cli.Command {
	cmd := &cli.Command{
		Name:  "greet",
		Usage: "say hello",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Value: "World",
				Usage: "name to greet",
			},
		},
		Action: func(ctx context.Context, cCtx *cli.Command) error {
			name := cCtx.String("name")
			fmt.Printf("Hello, %s!\n", name)
			return nil
		},
	}

	app := &cli.Command{
		Name:  "myCLI",
		Usage: "A simple CLI application",
		Commands: []*cli.Command{
			cmd,
			{
				Name:  "add",
				Usage: "add two numbers",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "a", Value: 0, Usage: "first number"},
					&cli.IntFlag{Name: "b", Value: 0, Usage: "second number"},
				},
				Action: func(ctx context.Context, cCtx *cli.Command) error {
					a := cCtx.Int("a")
					b := cCtx.Int("b")
					fmt.Printf("%d + %d = %d\n", a, b, a+b)
					return nil
				},
			},
		},
	}
	return app
}

func zapExample() {
	// Basic logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	logger.Info("Zap logger initialized (production)")

	// Example with fields
	logger.Info("failed to fetch URL",
		zap.String("url", "http://example.com"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", 1*time.Second),
	)

	// Sugar logger for more idiomatic use
	sugar := logger.Sugar()
	sugar.Infow("failed to fetch URL (sugared)",
		"url", "http://example.com",
		"attempt", 3,
		"backoff", strconv.Itoa(1)+"s", // Using strconv for simplicity here
	)
	sugar.Infof("Formatted log: %s attempt %d", "http://example.com", 3)
}
