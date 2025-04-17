package tests

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

func TestStdio(t *testing.T) {
	mockServerTrPath, err := compileMockStdioServerTr()
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		if err = os.Remove(name); err != nil {
			fmt.Printf("Failed to remove mock server: %v\n", err)
		}
	}(mockServerTrPath)

	fmt.Println(mockServerTrPath)
	transportClient, err := transport.NewStdioClientTransport(mockServerTrPath, []string{"-transport", "stdio"})
	if err != nil {
		t.Fatalf("Failed to create transport client: %v", err)
	}

	test(t, func() error {
		<-make(chan error)
		return nil
	}, transportClient)
}

func TestStdio2Stderr(t *testing.T) {
	mockServerTrPath, err := compileMockStdioServerTr()
	if err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		if err = os.Remove(name); err != nil {
			fmt.Printf("Failed to remove mock server: %v\n", err)
		}
	}(mockServerTrPath)

	fmt.Println(mockServerTrPath)
	transportClient, err := transport.NewStdioClientTransport(mockServerTrPath, []string{"-transport", "stdio"})
	if err != nil {
		t.Fatalf("Failed to create transport client: %v", err)
	}

	// test stderr
	isOk := false
	go func() {
		bufReader := bufio.NewReader(transportClient.GetStderr())
		buffer := make([]byte, 1024)
		for {
			// 读取数据到缓冲区
			n, err := bufReader.Read(buffer)
			if n > 0 {
				errInfo := string(buffer[:n])
				if strings.Contains(errInfo, "server StdErr") {
					isOk = true
				}
				// 处理读取到的数据，这里我们只是简单地打印到标准输出
				fmt.Printf("client stderr : %s \n", errInfo)
			}

			// 检查错误类型
			if err == io.EOF {
				// 读取结束
				break
			} else if err != nil {
				// 发生其他错误，进行处理
				fmt.Fprintf(os.Stderr, "Error reading from reader: %v\n", err)
				// break
			}
		}
	}()
	// Create MCP client using transport
	mcpClient, err := client.NewClient(transportClient, client.WithClientInfo(protocol.Implementation{
		Name:    "Example MCP Client",
		Version: "1.0.0",
	}))
	// Call tool
	_, err = mcpClient.CallTool(
		context.Background(),
		protocol.NewCallToolRequest("current time", map[string]interface{}{
			"timezone": "America/New_York_err",
		}))
	if !isOk {
		t.Fatalf("Failed to stderr")
	}
}
