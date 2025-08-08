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
				} else {
					// if arguments contains fileKey and fileName, open the Figma design
					if params, ok := rpcReq.Params.(map[string]interface{}); ok {
						if arguments, exists := params["arguments"]; exists {
							if argsMap, ok := arguments.(map[string]interface{}); ok {
								fileKey, fileKeyExists := argsMap["fileKey"].(string)
								fileName, fileNameExists := argsMap["fileName"].(string)
								if fileKeyExists && fileNameExists {
									util.OpenFigmaDesign(fileKey, fileName)
								}
							}
						}
					}
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
			// modify the response so that any tool call that has nodeId in the inputSchema.properties also takes a fileKey and fileName property
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
						for _, tool := range tools {
							if toolMap, ok := tool.(map[string]interface{}); ok {
								if inputSchema, exists := toolMap["inputSchema"]; exists {
									if inputSchemaMap, ok := inputSchema.(map[string]interface{}); ok {
										if properties, exists := inputSchemaMap["properties"]; exists {
											if propertiesMap, ok := properties.(map[string]interface{}); ok {
												if _, exists := propertiesMap["nodeId"]; exists {
													// Update the tool description to mention fileKey and fileName
													if desc, ok := toolMap["description"].(string); ok {
														toolMap["description"] = desc + " Use the fileKey and fileName parameters to specify a file. If a URL is provided, extract the fileKey and fileName from the URL, for example, if given the URL https://figma.com/design/1234/5678?node-id=1-2, the extracted fileKey would be `1234` and the extracted fileName would be `5678`."
													}
													// Update the tool properties to include fileKey and fileName
													propertiesMap["fileKey"] = map[string]interface{}{
														"type":        "string",
														"description": "The key of the file, extracted from the URL. For example, in https://figma.com/design/1234/5678?node-id=1-2, the fileKey is `1234`.",
													}
													propertiesMap["fileName"] = map[string]interface{}{
														"type":        "string",
														"description": "The name of the file, extracted from the URL. For example, in https://figma.com/design/1234/5678?node-id=1-2, the fileName is `5678`.",
													}
												}
											}
										}
									}
								}
							}
						}
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

	// {
	// 	"clientFrameworks": "react,next.js",
	// 	"clientLanguages": "typescript,javascript,html,css",
	// 	"clientName": "GitHub Copilot",
	// 	"nodeId": "1:119",
	// 	"fileKey": "JqWii6wYby2bPqnaaALroQ",
	// 	"fileName": "USER-10"
	// }
	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.ContentLength > 0 && r.ContentLength < 1024*1024 {
			body, err := io.ReadAll(r.Body)
			if err == nil {
				log.Printf("[MCP] %s %s from %s: %s", r.Method, r.URL.Path, r.RemoteAddr, string(body))
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
