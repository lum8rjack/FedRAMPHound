package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func runUpload(args []string) int {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	endpoint := fs.String("endpoint", "", "BloodHound base URL (e.g. https://bloodhound.example.com)")
	schemaPath := fs.String("schema", "fedramp-marketplace-schema.json", "extension schema JSON file")
	tokenID := fs.String("token-id", "", "BloodHound API token ID")
	tokenKey := fs.String("token-key", "", "BloodHound API token key")
	bearer := fs.String("bearer", "", "optional JWT bearer token (alternative to token-id/token-key)")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	_ = fs.Parse(args)

	if *endpoint == "" || *schemaPath == "" {
		fmt.Fprintln(os.Stderr, "usage: fedramphound upload -endpoint https://bloodhound.example.com -schema fedramp-marketplace-schema.json (-token-id ID -token-key KEY | -bearer JWT)")
		return 1
	}
	if *bearer == "" && (*tokenID == "" || *tokenKey == "") {
		fmt.Fprintln(os.Stderr, "provide either -bearer JWT or both -token-id and -token-key")
		return 1
	}

	body, err := os.ReadFile(*schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read schema: %v\n", err)
		return 1
	}

	base, err := url.Parse(strings.TrimRight(*endpoint, "/"))
	if err != nil || base.Scheme == "" || base.Host == "" {
		fmt.Fprintf(os.Stderr, "invalid endpoint URL: %q\n", *endpoint)
		return 1
	}

	reqURL := base.ResolveReference(&url.URL{Path: "/api/v2/extensions"})
	req, err := http.NewRequest(http.MethodPut, reqURL.String(), bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "build request: %v\n", err)
		return 1
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "fedramphound")
	req.Header.Set("Accept", "application/json")

	if *bearer != "" {
		req.Header.Set("Authorization", "Bearer "+*bearer)
	} else if err := signBloodHoundRequest(*tokenID, *tokenKey, req, body); err != nil {
		fmt.Fprintf(os.Stderr, "sign request: %v\n", err)
		return 1
	}

	client := &http.Client{Timeout: 60 * time.Second}
	if *insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // intentional for local CE
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "upload failed: %s\n%s\n", resp.Status, strings.TrimSpace(string(respBody)))
		return 1
	}

	fmt.Printf("Uploaded schema %s to %s (%s)\n", *schemaPath, reqURL.String(), resp.Status)
	if len(respBody) > 0 {
		fmt.Println(strings.TrimSpace(string(respBody)))
	}
	return 0
}

// signBloodHoundRequest applies BloodHound HMAC signed-request auth.
// See https://bloodhound.specterops.io/integrations/bloodhound-api/working-with-api
func signBloodHoundRequest(tokenID, tokenKey string, req *http.Request, body []byte) error {
	datetime := time.Now().Format(time.RFC3339)

	digester := hmac.New(sha256.New, []byte(tokenKey))
	if _, err := digester.Write([]byte(req.Method + req.URL.Path)); err != nil {
		return err
	}

	digester = hmac.New(sha256.New, digester.Sum(nil))
	if len(datetime) < 13 {
		return fmt.Errorf("unexpected RequestDate format")
	}
	if _, err := digester.Write([]byte(datetime[:13])); err != nil {
		return err
	}

	digester = hmac.New(sha256.New, digester.Sum(nil))
	if len(body) > 0 {
		if _, err := digester.Write(body); err != nil {
			return err
		}
	}

	req.Header.Set("Authorization", "bhesignature "+tokenID)
	req.Header.Set("RequestDate", datetime)
	req.Header.Set("Signature", base64.StdEncoding.EncodeToString(digester.Sum(nil)))

	// Body was already consumed into the request; reset for send.
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	return nil
}
