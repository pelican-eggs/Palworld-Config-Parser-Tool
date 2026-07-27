package main

// This file holds everything needed to talk to the Palworld server's REST
// admin API and to dispatch a single CLI command against it: the low-level
// API wrapper (PalworldAPI), the thin client wrapper (PalconClient), and
// the command table / dispatcher (runCLI) that main.go calls when invoked
// with `-cli "<command line>"`.

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------
// Low-level API client
// ---------------------------------------------------------------------

// PalworldAPI is a client for the Palworld server's REST admin API.
type PalworldAPI struct {
	ServerURL string
	Headers   map[string]string
	Client    *http.Client
}

// NewPalworldAPI creates a new API client. username defaults to "admin" when empty.
func NewPalworldAPI(serverURL, password, username string) *PalworldAPI {
	if username == "" {
		username = "admin"
	}
	token := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return &PalworldAPI{
		ServerURL: strings.TrimRight(serverURL, "/"),
		Headers: map[string]string{
			"Accept":        "application/json",
			"Content-Type":  "application/json",
			"Authorization": "Basic " + token,
		},
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

// kv is a single ordered key/value pair, used to preserve the original
// JSON key order (Python's json module preserves insertion order, and the
// CLI relies on that when printing "key: value" lines).
type kv struct {
	Key   string
	Value interface{}
}

// result is the generic response returned by every API call: it behaves
// like a map for lookups (e.g. checking for an "error" key) but also keeps
// the original top-level key order for display purposes.
type result struct {
	m       map[string]interface{}
	ordered []kv
}

func (r result) get(key string) (interface{}, bool) {
	v, ok := r.m[key]
	return v, ok
}

func errResult(format string, args ...interface{}) result {
	msg := fmt.Sprintf(format, args...)
	return result{
		m:       map[string]interface{}{"error": msg},
		ordered: []kv{{"error", msg}},
	}
}

// decodeOrderedObject decodes a JSON object while preserving key order.
// Returns ok=false if the payload isn't a JSON object.
func decodeOrderedObject(data []byte) (result, bool) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return result{}, false
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return result{}, false
	}

	r := result{m: map[string]interface{}{}}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return result{}, false
		}
		key, ok := keyTok.(string)
		if !ok {
			return result{}, false
		}
		var val interface{}
		if err := dec.Decode(&val); err != nil {
			return result{}, false
		}
		r.m[key] = val
		r.ordered = append(r.ordered, kv{key, val})
	}
	return r, true
}

func (a *PalworldAPI) doRequest(method, url string, payload interface{}) result {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return errResult("%s", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errResult("%s", err)
	}
	for k, v := range a.Headers {
		req.Header.Set(k, v)
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		if isTimeout(err) {
			return errResult("Request timeout")
		}
		return errResult("Connection error")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return errResult("%s", err)
	}

	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(data))
		return errResult("Client error %d: %s", resp.StatusCode, msg)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") && len(data) > 0 {
		if r, ok := decodeOrderedObject(data); ok {
			return r
		}
		// JSON but not an object (e.g. an array or scalar) - wrap it so
		// callers always get a result.
		var parsed interface{}
		if err := json.Unmarshal(data, &parsed); err == nil {
			return result{
				m:       map[string]interface{}{"result": parsed},
				ordered: []kv{{"result", parsed}},
			}
		}
	}

	text := string(data)
	return result{
		m:       map[string]interface{}{"result": text},
		ordered: []kv{{"result", text}},
	}
}

func isTimeout(err error) bool {
	type timeout interface{ Timeout() bool }
	if t, ok := err.(timeout); ok {
		return t.Timeout()
	}
	return strings.Contains(err.Error(), "timeout")
}

// Fetch performs a GET request against the given endpoint path.
func (a *PalworldAPI) Fetch(endpoint string) result {
	return a.doRequest(http.MethodGet, a.ServerURL+endpoint, nil)
}

// Post performs a POST request against the given endpoint path with an
// optional JSON payload.
func (a *PalworldAPI) Post(endpoint string, payload interface{}) result {
	return a.doRequest(http.MethodPost, a.ServerURL+endpoint, payload)
}

func (a *PalworldAPI) GetServerInfo() result     { return a.Fetch("/v1/api/info") }
func (a *PalworldAPI) GetPlayerList() result     { return a.Fetch("/v1/api/players") }
func (a *PalworldAPI) GetServerMetrics() result  { return a.Fetch("/v1/api/metrics") }
func (a *PalworldAPI) GetServerSettings() result { return a.Fetch("/v1/api/settings") }

func (a *PalworldAPI) KickPlayer(userid, message string) result {
	return a.Post("/v1/api/kick", map[string]string{"userid": userid, "message": message})
}

func (a *PalworldAPI) BanPlayer(userid, message string) result {
	return a.Post("/v1/api/ban", map[string]string{"userid": userid, "message": message})
}

func (a *PalworldAPI) UnbanPlayer(userid string) result {
	return a.Post("/v1/api/unban", map[string]string{"userid": userid})
}

func (a *PalworldAPI) SaveServerState() result { return a.Post("/v1/api/save", nil) }

func (a *PalworldAPI) MakeAnnouncement(message string) result {
	return a.Post("/v1/api/announce", map[string]string{"message": message})
}

func (a *PalworldAPI) ShutdownServer(waittime int, message string) result {
	return a.Post("/v1/api/shutdown", map[string]interface{}{"waittime": waittime, "message": message})
}

func (a *PalworldAPI) StopServer() result { return a.Post("/v1/api/stop", nil) }

// ---------------------------------------------------------------------
// Thin client wrapper
// ---------------------------------------------------------------------

// PalconClient is a thin wrapper around PalworldAPI.
type PalconClient struct {
	api *PalworldAPI
}

// NewPalconClient creates a client for the given server URL and admin password.
func NewPalconClient(serverURL, password string) *PalconClient {
	return &PalconClient{api: NewPalworldAPI(serverURL, password, "admin")}
}

func (c *PalconClient) GetServerInfo() result       { return c.api.GetServerInfo() }
func (c *PalconClient) GetPlayers() result          { return c.api.GetPlayerList() }
func (c *PalconClient) GetServerSettings() result   { return c.api.GetServerSettings() }
func (c *PalconClient) GetServerMetrics() result    { return c.api.GetServerMetrics() }
func (c *PalconClient) UnbanPlayer(id string) result { return c.api.UnbanPlayer(id) }
func (c *PalconClient) StopServer() result          { return c.api.StopServer() }
func (c *PalconClient) SaveServerState() result     { return c.api.SaveServerState() }

func (c *PalconClient) KickPlayer(playerID, reason string) result {
	return c.api.KickPlayer(playerID, reason)
}

func (c *PalconClient) BanPlayer(playerID, reason string) result {
	return c.api.BanPlayer(playerID, reason)
}

func (c *PalconClient) MakeAnnouncement(message string) result {
	return c.api.MakeAnnouncement(message)
}

func (c *PalconClient) ShutdownServer(seconds int, message string) result {
	return c.api.ShutdownServer(seconds, message)
}

// ---------------------------------------------------------------------
// CLI command layer
// ---------------------------------------------------------------------

const helpText = `
Available commands:
  info
  players
  settings
  metrics
  kick <player_id> <reason>
  ban <player_id> <reason>
  unban <player_id>
  announce <message>
  shutdown <seconds> <message>
  stop
  save
`

func showHelp() {
	fmt.Println(helpText)
}

// formatOutput renders a result as one "key: value" line per top-level
// field, in the original JSON key order.
func formatOutput(r result) string {
	if len(r.ordered) == 0 {
		return ""
	}
	lines := make([]string, 0, len(r.ordered))
	for _, pair := range r.ordered {
		lines = append(lines, fmt.Sprintf("%s: %s", pair.Key, pyStr(pair.Value)))
	}
	return strings.Join(lines, "\n")
}

// pyStr renders a decoded JSON value the way Python's str() would render
// the equivalent object (dicts/lists print with Python-style literals).
func pyStr(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "None"
	case bool:
		if val {
			return "True"
		}
		return "False"
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'g', -1, 64)
	case map[string]interface{}:
		parts := make([]string, 0, len(val))
		for k, vv := range val {
			parts = append(parts, fmt.Sprintf("'%s': %s", k, pyRepr(vv)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case []interface{}:
		parts := make([]string, 0, len(val))
		for _, vv := range val {
			parts = append(parts, pyRepr(vv))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// pyRepr is like pyStr but quotes strings, matching Python's repr() used
// when printing nested values inside a dict/list.
func pyRepr(v interface{}) string {
	if s, ok := v.(string); ok {
		return "'" + s + "'"
	}
	return pyStr(v)
}

// formatPlayers renders the player list in human-readable form.
func formatPlayers(r result) string {
	playersVal, ok := r.get("players")
	if !ok {
		return "No players online"
	}
	players, ok := playersVal.([]interface{})
	if !ok || len(players) == 0 {
		return "No players online"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "=== Player List (%d online) ===\n\n", len(players))

	for i, p := range players {
		player, _ := p.(map[string]interface{})
		accountName := stringOr(player["accountName"], "N/A")
		userID := stringOr(player["userId"], "N/A")
		playerID := stringOr(player["playerId"], "N/A")

		fmt.Fprintf(&b, "%d. %s\n", i+1, accountName)
		fmt.Fprintf(&b, "   User ID: %s\n", userID)
		fmt.Fprintf(&b, "   Player ID: %s\n", playerID)
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func stringOr(v interface{}, fallback string) string {
	if v == nil {
		return fallback
	}
	return pyStr(v)
}

func printResult(r result, formatter func(result) string) {
	fmt.Println(formatter(r))
}

// command represents a single CLI subcommand.
type command struct {
	name string
	help string
	run  func(client *PalconClient, args []string)
}

func usage(cmdLine string) {
	fmt.Printf("Usage: %s\n", cmdLine)
	fmt.Println("Type 'help' for more commands")
}

var commands []command

func init() {
	commands = []command{
		{"info", "Get server information", func(c *PalconClient, args []string) {
			printResult(c.GetServerInfo(), formatOutput)
		}},
		{"players", "List online players", func(c *PalconClient, args []string) {
			printResult(c.GetPlayers(), formatPlayers)
		}},
		{"settings", "Get server settings", func(c *PalconClient, args []string) {
			printResult(c.GetServerSettings(), formatOutput)
		}},
		{"metrics", "Get server metrics", func(c *PalconClient, args []string) {
			printResult(c.GetServerMetrics(), formatOutput)
		}},
		{"kick", "Kick a player from the server", func(c *PalconClient, args []string) {
			if len(args) < 2 {
				usage("kick <player_id> <reason>")
				return
			}
			playerID := args[0]
			reason := strings.Join(args[1:], " ")
			r := c.KickPlayer(playerID, reason)
			if errMsg, ok := r.get("error"); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", pyStr(errMsg))
				os.Exit(1)
			}
			fmt.Printf("Player %s kicked\n", playerID)
		}},
		{"ban", "Ban a player from the server", func(c *PalconClient, args []string) {
			if len(args) < 2 {
				usage("ban <player_id> <reason>")
				return
			}
			playerID := args[0]
			reason := strings.Join(args[1:], " ")
			r := c.BanPlayer(playerID, reason)
			if errMsg, ok := r.get("error"); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", pyStr(errMsg))
				os.Exit(1)
			}
			fmt.Printf("Player %s banned\n", playerID)
		}},
		{"unban", "Unban a player", func(c *PalconClient, args []string) {
			if len(args) < 1 {
				usage("unban <player_id>")
				return
			}
			playerID := args[0]
			r := c.UnbanPlayer(playerID)
			if errMsg, ok := r.get("error"); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", pyStr(errMsg))
				os.Exit(1)
			}
			fmt.Printf("Player %s unbanned\n", playerID)
		}},
		{"announce", "Make a server announcement", func(c *PalconClient, args []string) {
			if len(args) < 1 {
				usage("announce <message>")
				return
			}
			message := strings.Join(args, " ")
			r := c.MakeAnnouncement(message)
			if errMsg, ok := r.get("error"); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", pyStr(errMsg))
				os.Exit(1)
			}
			fmt.Println("Announcement sent")
		}},
		{"shutdown", "Shutdown server with delay", func(c *PalconClient, args []string) {
			if len(args) < 2 {
				usage("shutdown <seconds> <message>")
				return
			}
			seconds, err := strconv.Atoi(args[0])
			if err != nil {
				usage("shutdown <seconds> <message>")
				return
			}
			message := strings.Join(args[1:], " ")
			r := c.ShutdownServer(seconds, message)
			if errMsg, ok := r.get("error"); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", pyStr(errMsg))
				os.Exit(1)
			}
			fmt.Printf("Server shutting down in %d seconds\n", seconds)
		}},
		{"stop", "Stop server immediately", func(c *PalconClient, args []string) {
			r := c.StopServer()
			if errMsg, ok := r.get("error"); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", pyStr(errMsg))
				return
			}
			fmt.Println("Server stopped")
		}},
		{"save", "Save server state", func(c *PalconClient, args []string) {
			r := c.SaveServerState()
			if errMsg, ok := r.get("error"); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", pyStr(errMsg))
				return
			}
			fmt.Println("Server saved")
		}},
		{"help", "Show available commands", func(c *PalconClient, args []string) {
			showHelp()
		}},
	}
}

func findCommand(name string) *command {
	for i := range commands {
		if commands[i].name == name {
			return &commands[i]
		}
	}
	return nil
}

// runCLI parses global flags (--url/-u, --password/-p) out of args,
// resolves the remaining subcommand and its arguments, and dispatches.
// Called once per process by main.go for every line read from stdin
// (see the bash one-liner in the package doc comment in main.go).
func runCLI(args []string) {
	if len(args) == 0 {
		showHelp()
		return
	}

	url := ""
	password := ""
	var rest []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--url", "-u":
			if i+1 < len(args) {
				url = args[i+1]
				i++
			}
		case "--password", "-p":
			if i+1 < len(args) {
				password = args[i+1]
				i++
			}
		default:
			if strings.HasPrefix(args[i], "--url=") {
				url = strings.TrimPrefix(args[i], "--url=")
			} else if strings.HasPrefix(args[i], "--password=") {
				password = strings.TrimPrefix(args[i], "--password=")
			} else {
				rest = append(rest, args[i])
			}
		}
	}

	if len(rest) == 0 {
		showHelp()
		return
	}

	cmdName := rest[0]
	cmdArgs := rest[1:]

	cmd := findCommand(cmdName)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
		showHelp()
		return
	}

	if cmdName == "help" {
		showHelp()
		return
	}

	if url == "" {
		url = os.Getenv("PALWORLD_URL")
	}
	if url == "" {
		if port := os.Getenv("REST_API_PORT"); port != "" {
			url = "http://localhost:" + port
		}
	}
	if password == "" {
		password = os.Getenv("ADMIN_PASSWORD")
	}

	if url == "" {
		fmt.Fprintln(os.Stderr, "Error: Server URL required")
		return
	}
	if password == "" {
		fmt.Fprintln(os.Stderr, "Error: Admin password required")
		return
	}

	client := NewPalconClient(url, password)
	cmd.run(client, cmdArgs)
}