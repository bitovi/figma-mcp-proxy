package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bitovi/figma-mcp-proxy/util"
)

type MCPRequestBody struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// handleOpenFigmaDesignTool handles the open_figma_design_file tool call directly
func handleOpenFigmaDesignTool(w http.ResponseWriter, fileKey, fileName string, requestID int) {
	// Try to open the Figma design
	err := util.OpenFigmaDesign(fileKey, fileName)

	var response MCPResponse
	response.JSONRPC = "2.0"
	response.ID = requestID

	if err != nil {
		response.Error = map[string]interface{}{
			"code":    -32603,
			"message": fmt.Sprintf("Failed to open Figma design: %v", err),
		}
	} else {
		response.Result = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Successfully opened Figma design file '%s/%s'", fileKey, fileName),
				},
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[PROXY] Error encoding response: %v", err)
	}
}

func main() {
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		targetURL = "http://localhost:3845"
		log.Printf("No TARGET_URL specified, using default: %s", targetURL)
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxyRequestToTarget := proxy.Director

	proxy.Director = func(req *http.Request) {
		var rpcReq MCPRequestBody
		if req.Body != nil {
			requestBody, err := readBody(req.Body)
			if err != nil {
				log.Printf("[PROXY] Error reading request body in Director: %v", err)
			} else {
				if err := json.Unmarshal([]byte(requestBody), &rpcReq); err != nil {
					log.Printf("[PROXY] Error unmarshalling request body: %v", err)
				}

				// Store the original request body before it gets consumed so it can be used to modify the response later
				req.Body = io.NopCloser(strings.NewReader(requestBody))
				req.Header.Set("X-Original-Request-Body", requestBody)
			}
		}

		proxyRequestToTarget(req)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		// Get the original request body that was stored in the Director function
		requestBody := resp.Request.Header.Get("X-Original-Request-Body")

		var rpcReq MCPRequestBody
		if requestBody != "" {
			if err := json.Unmarshal([]byte(requestBody), &rpcReq); err != nil {
				log.Printf("[PROXY] Error unmarshalling request body: %v", err)
			}
		}

		if rpcReq.Method == "tools/list" {
			// Add the open_figma_design_file tool to the response
			if resp.StatusCode == http.StatusOK {
				// Read the entire response body as text
				rawBody, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("[PROXY] Error reading response body: %v", err)
					return err
				}

				// Try to extract JSON from SSE format (lines starting with "data: ")
				var jsonPayload string
				for _, line := range strings.Split(string(rawBody), "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "data: ") {
						jsonPayload = strings.TrimPrefix(line, "data: ")
						break
					}
				}
				if jsonPayload == "" {
					return nil
				}

				var responseBody map[string]interface{}
				if err := json.Unmarshal([]byte(jsonPayload), &responseBody); err != nil {
					log.Printf("[PROXY] Error decoding JSON payload: %v", err)
					return err
				}

				if result, ok := responseBody["result"].(map[string]interface{}); ok {
					if tools, ok := result["tools"].([]interface{}); ok {
						// Add the new open_figma_design_file tool
						openFigmaDesignTool := map[string]interface{}{
							"name":        "open_figma_design_file",
							"description": `Opens a Figma design file using the fileKey and fileName parameters. Use this tool to open a specific Figma design file. If a URL is provided, extract the fileKey and fileName from the URL. For example, if given the URL https://www.figma.com/design/JqWii6wYby2bPqnaaALroQ/USER-10?node-id=1-82, the extracted fileKey would be "JqWii6wYby2bPqnaaALroQ" and the extracted fileName would be "USER-10".`,
							"inputSchema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"fileKey": map[string]interface{}{
										"type":        "string",
										"description": `The key of the file, extracted from the URL. For example, in https://www.figma.com/design/JqWii6wYby2bPqnaaALroQ/USER-10?node-id=1-82, the fileKey is "JqWii6wYby2bPqnaaALroQ".`,
									},
									"fileName": map[string]interface{}{
										"type":        "string",
										"description": `The name of the file, extracted from the URL. For example, in https://www.figma.com/design/JqWii6wYby2bPqnaaALroQ/USER-10?node-id=1-82, the fileName is "USER-10".`,
									},
								},
							},
						}

						// Add the new tool to the tools array
						tools = append(tools, openFigmaDesignTool)
						result["tools"] = tools
					}
				}

				modifiedBody, err := json.Marshal(responseBody)
				if err != nil {
					log.Printf("[PROXY] Error marshalling modified response body: %v", err)
					return err
				}
				resp.Body = io.NopCloser(strings.NewReader(fmt.Sprintf("event: message\ndata: %s\n\n", modifiedBody)))
			}
		}

		return nil
	}

	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.ContentLength > 0 && r.ContentLength < 1024*1024 {
			body, err := io.ReadAll(r.Body)
			if err == nil {
				log.Printf("[MCP] %s %s from %s: %s", r.Method, r.URL.Path, r.RemoteAddr, string(body))

				// Check if this is a tools/call request for open_figma_design_file
				var rpcReq MCPRequestBody
				if err := json.Unmarshal(body, &rpcReq); err == nil {
					if rpcReq.Method == "tools/call" {
						if params, ok := rpcReq.Params.(map[string]interface{}); ok {
							if name, exists := params["name"]; exists && name == "open_figma_design_file" {
								if arguments, exists := params["arguments"]; exists {
									if argsMap, ok := arguments.(map[string]interface{}); ok {
										fileKey, fileKeyExists := argsMap["fileKey"].(string)
										fileName, fileNameExists := argsMap["fileName"].(string)
										if fileKeyExists && fileNameExists {
											// Handle the tool call directly and return response
											handleOpenFigmaDesignTool(w, fileKey, fileName, rpcReq.ID)
											return
										}
									}
								}
							}
						}
					}
				}

				r.Body = io.NopCloser(strings.NewReader(string(body)))
			}
		}

		proxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := struct {
			Status    string `json:"status"`
			TargetURL string `json:"targetURL"`
		}{
			Status:    "OK",
			TargetURL: targetURL,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3846"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Proxying /mcp requests to: %s", targetURL)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func readBody(rc io.ReadCloser) (string, error) {
	var body string
	if rc != nil {
		b, err := io.ReadAll(rc)
		if err == nil {
			body = string(b)
			rc.Close()
		} else {
			return "", err
		}
	}
	return body, nil
}
