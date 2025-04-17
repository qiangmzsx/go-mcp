package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

func main() {
	t, err := transport.NewStdioClientTransport("../everything", []string{"-transport", "stdio"})
	if err != nil {
		log.Fatal(err)
		return
	}
	sseClient, err := client.NewClient(t, client.WithClientInfo(protocol.Implementation{
		Name:    "everything_client",
		Version: "1.0.0",
	}))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func() {
		if err = sseClient.Close(); err != nil {
			log.Fatalf("Failed to close MCP client: %v", err)
			return
		}
	}()

	go func() {
		bufReader := bufio.NewReader(t.GetStderr())
		buffer := make([]byte, 1024)
		for {
			// Read data into buffer
			n, err := bufReader.Read(buffer)
			if n > 0 {
				// Process the read data, here we simply print it to standard output
				fmt.Printf("client StdErr: %s \n", string(buffer[:n]))
			}

			// Check for error type
			if err == io.EOF {
				// End of reading
				break
			} else if err != nil {
				// Handle other errors
				fmt.Fprintf(os.Stderr, "Error reading from reader: %v\n", err)
			}
		}

	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("Listing available tools...")
	tools, err := sseClient.ListTools(ctx)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}
	for _, tool := range tools.Tools {
		log.Printf("- %s: %s\n", tool.Name, tool.Description)
	}
	result, err := sseClient.CallTool(ctx, &protocol.CallToolRequest{
		Name: "current time",
		Arguments: map[string]interface{}{
			"timezone": "1America/New_York",
		},
	})
	if err != nil {
		log.Printf("Failed to call tool: %v \n", err)
	}
	log.Println(result)
	time.Sleep(3 * time.Second)
}
