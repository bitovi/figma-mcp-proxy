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
	"sync"
	"time"

	"context"

	"github.com/bitovi/figma-mcp-proxy/util"
	"github.com/google/uuid"
)

type MCPRequestBody struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

var designFileMutex sync.Mutex

type ctxKeyRequestID struct{}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()
		log.Printf("[MIDDLEWARE] [%s] New request: %s %s from %s", reqID, r.Method, r.URL.String(), r.RemoteAddr)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, reqID)
		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		duration := time.Since(start)
		log.Printf("[MIDDLEWARE] [%s] Request completed in %v", reqID, duration)
	})
}

func getRequestID(r *http.Request) string {
	if v := r.Context().Value(ctxKeyRequestID{}); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func main() {
	log.Printf("[MAIN] Starting Figma MCP Proxy application")

	targetURL := os.Getenv("TARGET_URL")
	log.Printf("[MAIN] Environment variable TARGET_URL: %q", targetURL)
	if targetURL == "" {
		targetURL = "http://localhost:3845"
		log.Printf("[MAIN] No TARGET_URL specified, using default: %s", targetURL)
	} else {
		log.Printf("[MAIN] Using TARGET_URL from environment: %s", targetURL)
	}

	log.Printf("[MAIN] Parsing target URL: %s", targetURL)
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("[MAIN] Failed to parse target URL: %v", err)
	}
	log.Printf("[MAIN] Successfully parsed target URL - Scheme: %s, Host: %s", target.Scheme, target.Host)

	log.Printf("[MAIN] Creating reverse proxy to target: %s", target.String())
	proxy := httputil.NewSingleHostReverseProxy(target)
	log.Printf("[MAIN] Reverse proxy created successfully")

	proxyRequestToTarget := proxy.Director

	proxy.Director = func(req *http.Request) {
		reqID := getRequestID(req)
		log.Printf("[DIRECTOR] [%s] Processing request: %s %s", reqID, req.Method, req.URL.String())
		log.Printf("[DIRECTOR] [%s] MCP Session ID: %s", reqID, req.Header.Get("Mcp-Session-Id"))

		externalDNSName := os.Getenv("EXTERNAL_DNS_NAME")
		log.Printf("[DIRECTOR] [%s] External DNS name: %q", reqID, externalDNSName)
		if externalDNSName != "" {
			log.Printf("[DIRECTOR] [%s] Setting up external DNS routing", reqID)
			remote, err := url.Parse(externalDNSName)
			if err != nil {
				log.Fatalf("[DIRECTOR] [%s] Failed to parse remote URL: %v", reqID, err)
			}
			log.Printf("[DIRECTOR] [%s] Remote URL parsed - Scheme: %s, Host: %s", reqID, remote.Scheme, remote.Host)

			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			forwardedHost := req.Header.Get("X-Forwarded-Host")
			req.Host = forwardedHost
			log.Printf("[DIRECTOR] [%s] Updated request URL to %s, Host header set to %s", reqID, req.URL.String(), forwardedHost)
		} else {
			log.Printf("[DIRECTOR] [%s] No external DNS name configured, using default routing", reqID)
		}

		var rpcReq MCPRequestBody
		log.Printf("[DIRECTOR] [%s] Checking for request body", reqID)
		if req.Body != nil {
			log.Printf("[DIRECTOR] [%s] Request body found, reading...", reqID)
			requestBody, err := readBody(req.Body)
			if err != nil {
				log.Printf("[DIRECTOR] [%s] ERROR: Failed to read request body: %v", reqID, err)
			} else {
				log.Printf("[DIRECTOR] [%s] Request body read successfully, length: %d", reqID, len(requestBody))
				if err := json.Unmarshal([]byte(requestBody), &rpcReq); err != nil {
					log.Printf("[DIRECTOR] [%s] ERROR: Failed to unmarshal request body: %v", reqID, err)
				} else {
					log.Printf("[DIRECTOR] [%s] Request body parsed - Method: %s, ID: %d", reqID, rpcReq.Method, rpcReq.ID)
					// Check if arguments contains fileKey and fileName to open Figma design
					log.Printf("[DIRECTOR] [%s] Checking for Figma design parameters in request params", reqID)
					if params, ok := rpcReq.Params.(map[string]interface{}); ok {
						log.Printf("[DIRECTOR] [%s] Request params found, checking for arguments", reqID)
						if arguments, exists := params["arguments"]; exists {
							log.Printf("[DIRECTOR] [%s] Arguments found in params, checking for Figma parameters", reqID)
							if argsMap, ok := arguments.(map[string]interface{}); ok {
								fileKey, fileKeyExists := argsMap["fileKey"].(string)
								fileName, fileNameExists := argsMap["fileName"].(string)
								nodeId, nodeIdExists := argsMap["nodeId"].(string)
								log.Printf("[DIRECTOR] [%s] Figma params check - fileKey: %v, fileName: %v, nodeId: %v", reqID, fileKeyExists, fileNameExists, nodeIdExists)

								if fileKeyExists && fileNameExists && nodeIdExists {
									log.Printf("[DIRECTOR] [%s] All Figma parameters present, attempting to open design: %s/%s?node-id=%s", reqID, fileKey, fileName, nodeId)
									designFileMutex.Lock()
									log.Printf("[DIRECTOR] [%s] Mutex LOCKED for figma://design/%s/%s?node-id=%s", reqID, fileKey, fileName, nodeId)
									defer func() {
										designFileMutex.Unlock()
										log.Printf("[DIRECTOR] [%s] Mutex UNLOCKED for figma://design/%s/%s?node-id=%s", reqID, fileKey, fileName, nodeId)
									}()
									if err := util.OpenFigmaDesign(fileKey, fileName, nodeId); err != nil {
										log.Printf("[DIRECTOR] [%s] ERROR: Failed to open Figma design: %v", reqID, err)
									} else {
										log.Printf("[DIRECTOR] [%s] Successfully opened Figma design: figma://design/%s/%s?node-id=%s", reqID, fileKey, fileName, nodeId)
									}
								} else {
									log.Printf("[DIRECTOR] [%s] Missing Figma parameters, skipping design open", reqID)
								}
							} else {
								log.Printf("[DIRECTOR] [%s] Arguments not in expected format", reqID)
							}
						} else {
							log.Printf("[DIRECTOR] [%s] No arguments found in params", reqID)
						}
					} else {
						log.Printf("[DIRECTOR] [%s] Params not in expected format", reqID)
					}

					// Store the original request body before it gets consumed so it can be used to modify the response later
					log.Printf("[DIRECTOR] [%s] Storing original request body for response modification", reqID)
					req.Body = io.NopCloser(strings.NewReader(requestBody))
					req.Header.Set("X-Original-Request-Body", requestBody)
					log.Printf("[DIRECTOR] [%s] Request body stored in header", reqID)
				}
			}
		} else {
			log.Printf("[DIRECTOR] [%s] No request body present", reqID)
		}

		log.Printf("[DIRECTOR] [%s] Calling target proxy for %s %s", reqID, req.Method, req.URL.String())
		proxyRequestToTarget(req)
		log.Printf("[DIRECTOR] [%s] Director processing completed for %s %s", reqID, req.Method, req.URL.String())
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		reqID := getRequestID(resp.Request)
		log.Printf("[MODIFY_RESPONSE] [%s] Processing response for %s %s (Status: %d)", reqID, resp.Request.Method, resp.Request.URL.String(), resp.StatusCode)

		// Get the original request body that was stored in the Director function
		requestBody := resp.Request.Header.Get("X-Original-Request-Body")
		log.Printf("[MODIFY_RESPONSE] [%s] Retrieved original request body, length: %d", reqID, len(requestBody))

		var rpcReq MCPRequestBody
		if requestBody != "" {
			log.Printf("[MODIFY_RESPONSE] [%s] Unmarshaling request body for method detection", reqID)
			if err := json.Unmarshal([]byte(requestBody), &rpcReq); err != nil {
				log.Printf("[MODIFY_RESPONSE] [%s] ERROR: Failed to unmarshal request body: %v", reqID, err)
			} else {
				log.Printf("[MODIFY_RESPONSE] [%s] Request body unmarshaled - Method: %s", reqID, rpcReq.Method)
			}
		} else {
			log.Printf("[MODIFY_RESPONSE] [%s] No request body to process", reqID)
		}

		if rpcReq.Method == "tools/list" {
			log.Printf("[MODIFY_RESPONSE] [%s] Processing tools/list response for modification", reqID)
			// modify the response so that any tool call that has nodeId in the inputSchema.properties also takes a fileKey and fileName property
			if resp.StatusCode == http.StatusOK {
				log.Printf("[MODIFY_RESPONSE] [%s] Response status OK, reading response body", reqID)
				// Read the entire response body as text
				rawBody, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("[MODIFY_RESPONSE] [%s] ERROR: Failed to read response body: %v", reqID, err)
					return err
				}
				log.Printf("[MODIFY_RESPONSE] [%s] Response body read, length: %d", reqID, len(rawBody))

				// Try to extract JSON from SSE format (lines starting with "data: ")
				log.Printf("[MODIFY_RESPONSE] [%s] Extracting JSON payload from SSE format", reqID)
				var jsonPayload string
				for _, line := range strings.Split(string(rawBody), "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "data: ") {
						jsonPayload = strings.TrimPrefix(line, "data: ")
						log.Printf("[MODIFY_RESPONSE] [%s] Found JSON payload in SSE format, length: %d", reqID, len(jsonPayload))
						break
					}
				}
				if jsonPayload == "" {
					log.Printf("[MODIFY_RESPONSE] [%s] No JSON payload found in response, skipping modification", reqID)
					return nil
				}

				var responseBody map[string]interface{}
				if err := json.Unmarshal([]byte(jsonPayload), &responseBody); err != nil {
					log.Printf("[MODIFY_RESPONSE] [%s] ERROR: Failed to decode JSON payload: %v", reqID, err)
					return err
				}
				log.Printf("[MODIFY_RESPONSE] [%s] JSON payload decoded successfully", reqID)

				if result, ok := responseBody["result"].(map[string]interface{}); ok {
					log.Printf("[MODIFY_RESPONSE] [%s] Found result object in response", reqID)
					if tools, ok := result["tools"].([]interface{}); ok {
						log.Printf("[MODIFY_RESPONSE] [%s] Found %d tools in response", reqID, len(tools))
						toolsModified := 0
						for _, tool := range tools {
							if toolMap, ok := tool.(map[string]interface{}); ok {
								if inputSchema, exists := toolMap["inputSchema"]; exists {
									if inputSchemaMap, ok := inputSchema.(map[string]interface{}); ok {
										if properties, exists := inputSchemaMap["properties"]; exists {
											if propertiesMap, ok := properties.(map[string]interface{}); ok {
												if _, exists := propertiesMap["nodeId"]; exists {
													log.Printf("[MODIFY_RESPONSE] [%s] Found tool with nodeId property, modifying: %v", reqID, toolMap["name"])
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

													// Add fileKey and fileName to required array
													if required, exists := inputSchemaMap["required"]; exists {
														if requiredArray, ok := required.([]interface{}); ok {
															requiredArray = append(requiredArray, "fileKey", "fileName")
															inputSchemaMap["required"] = requiredArray
														}
													} else {
														inputSchemaMap["required"] = []string{"fileKey", "fileName"}
													}
													toolsModified++
													log.Printf("[MODIFY_RESPONSE] [%s] Tool modified successfully: %v", reqID, toolMap["name"])
												}
											}
										}
									}
								}
							}
						}
						log.Printf("[MODIFY_RESPONSE] [%s] Tool modification completed, %d tools modified", reqID, toolsModified)
					} else {
						log.Printf("[MODIFY_RESPONSE] [%s] No tools array found in result", reqID)
					}
				} else {
					log.Printf("[MODIFY_RESPONSE] [%s] No result object found in response", reqID)
				}

				log.Printf("[MODIFY_RESPONSE] [%s] Marshaling modified response body", reqID)
				modifiedBody, err := json.Marshal(responseBody)
				if err != nil {
					log.Printf("[MODIFY_RESPONSE] [%s] ERROR: Failed to marshal modified response body: %v", reqID, err)
					return err
				}
				log.Printf("[MODIFY_RESPONSE] [%s] Response body marshaled, creating new response", reqID)
				resp.Body = io.NopCloser(strings.NewReader(fmt.Sprintf("event: message\ndata: %s\n\n", modifiedBody)))
				log.Printf("[MODIFY_RESPONSE] [%s] Modified response body set", reqID)
			} else {
				log.Printf("[MODIFY_RESPONSE] [%s] Response status not OK (%d), skipping modification", reqID, resp.StatusCode)
			}
		} else {
			log.Printf("[MODIFY_RESPONSE] [%s] Not a tools/list request, skipping modification", reqID)
		}

		log.Printf("[MODIFY_RESPONSE] [%s] ModifyResponse completed", reqID)
		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		reqID := getRequestID(r)
		log.Printf("[ERROR_HANDLER] [%s] Proxy error for %s %s: %v", reqID, r.Method, r.URL.String(), err)
		log.Printf("[ERROR_HANDLER] [%s] MCP Session ID: %s", reqID, r.Header.Get("Mcp-Session-Id"))
		http.Error(w, "Proxy error: "+err.Error(), http.StatusBadGateway)
	}

	var apiKey = os.Getenv("API_KEY")
	log.Printf("[MAIN] API key configured: %v", apiKey != "")
	http.Handle("/mcp", withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := getRequestID(r)
		log.Printf("[MCP_HANDLER] [%s] Processing /mcp request", reqID)

		if apiKey != "" {
			log.Printf("[MCP_HANDLER] [%s] API key authentication required", reqID)
			authHeader := r.Header.Get("Authorization")
			expectedAuth := fmt.Sprintf("Bearer %s", apiKey)
			if authHeader != expectedAuth {
				log.Printf("[MCP_HANDLER] [%s] Authentication failed - received: %q", reqID, authHeader)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			log.Printf("[MCP_HANDLER] [%s] Authentication successful", reqID)
		} else {
			log.Printf("[MCP_HANDLER] [%s] No API key configured, skipping authentication", reqID)
		}

		if r.Method != "GET" && r.ContentLength > 0 && r.ContentLength < 1024*1024 {
			log.Printf("[MCP_HANDLER] [%s] Reading request body (Content-Length: %d)", reqID, r.ContentLength)
			body, err := io.ReadAll(r.Body)
			if err == nil {
				log.Printf("[MCP_HANDLER] [%s] Received Request %s %s from %s: %s", reqID, r.Method, r.URL.Path, r.RemoteAddr, string(body))
				r.Body = io.NopCloser(strings.NewReader(string(body)))
				log.Printf("[MCP_HANDLER] [%s] Request body restored for proxy", reqID)
			} else {
				log.Printf("[MCP_HANDLER] [%s] ERROR: Failed to read request body: %v", reqID, err)
			}
		} else {
			log.Printf("[MCP_HANDLER] [%s] Skipping body read - Method: %s, ContentLength: %d", reqID, r.Method, r.ContentLength)
		}

		log.Printf("[MCP_HANDLER] [%s] Proxying request to target", reqID)
		proxy.ServeHTTP(w, r)
		log.Printf("[MCP_HANDLER] [%s] Request processing completed", reqID)
	})))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HEALTH] Health check requested from %s", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := struct {
			Status    string `json:"status"`
			TargetURL string `json:"targetURL"`
		}{
			Status:    "OK",
			TargetURL: targetURL,
		}
		log.Printf("[HEALTH] Responding with status OK, target URL: %s", targetURL)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("[HEALTH] ERROR: Failed to encode JSON response: %v", err)
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		} else {
			log.Printf("[HEALTH] Health check completed successfully")
		}
	})

	port := os.Getenv("PORT")
	log.Printf("[MAIN] Environment variable PORT: %q", port)
	if port == "" {
		port = "3846"
		log.Printf("[MAIN] No PORT specified, using default: %s", port)
	} else {
		log.Printf("[MAIN] Using PORT from environment: %s", port)
	}

	log.Printf("[MAIN] Starting server on port %s", port)
	log.Printf("[MAIN] Proxying /mcp requests to: %s", targetURL)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Printf("[MAIN] Server configured with timeouts - Read: %v, Write: %v, Idle: %v",
		server.ReadTimeout, server.WriteTimeout, server.IdleTimeout)

	log.Printf("[MAIN] Server starting to listen and serve on address: %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func readBody(rc io.ReadCloser) (string, error) {
	log.Printf("[READ_BODY] Starting to read request body")
	var body string
	if rc != nil {
		log.Printf("[READ_BODY] ReadCloser provided, reading all content")
		b, err := io.ReadAll(rc)
		if err == nil {
			body = string(b)
			log.Printf("[READ_BODY] Successfully read %d bytes from body", len(body))
			rc.Close()
			log.Printf("[READ_BODY] ReadCloser closed")
		} else {
			log.Printf("[READ_BODY] ERROR: Failed to read from ReadCloser: %v", err)
			return "", err
		}
	} else {
		log.Printf("[READ_BODY] No ReadCloser provided, returning empty body")
	}
	return body, nil
}
