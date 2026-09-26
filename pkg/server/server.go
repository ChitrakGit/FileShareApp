package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"fileshare/pkg/discovery"
	"fileshare/pkg/store"
	"fileshare/pkg/transfer"
	"fileshare/web"
)

// Server handles the HTTP API and Web UI serving
type Server struct {
	Port        int
	DownloadDir string
	Discovery   *discovery.Service
	AutoAccept  bool
	ExpectedPin string

	clientsMu  sync.Mutex
	sseClients map[chan string]bool
}

// NewServer initializes a new FileShare server instance
func NewServer(port int, downloadDir string, disc *discovery.Service, autoAccept bool, pin string) *Server {
	if downloadDir == "" {
		downloadDir = store.DefaultSettings().Get().DownloadDir
	}

	return &Server{
		Port:        port,
		DownloadDir: downloadDir,
		Discovery:   disc,
		AutoAccept:  autoAccept,
		ExpectedPin: pin,
		sseClients:  make(map[chan string]bool),
	}
}

// Start boots up the HTTP server and blocks
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/peers", s.handlePeersList)
	mux.HandleFunc("/api/peers/add", s.handlePeersAdd)
	mux.HandleFunc("/api/peers/remove", s.handlePeersRemove)
	mux.HandleFunc("/api/peers/scan-subnet", s.handleSubnetScan)
	mux.HandleFunc("/api/upload", s.handleMultipartUpload)
	mux.HandleFunc("/api/stream-upload", s.handleStreamUpload)
	mux.HandleFunc("/api/sse", s.handleSSE)
	mux.HandleFunc("/api/firewall", s.handleFirewallConfig)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/history", s.handleHistory)

	// Embedded Web UI
	mux.Handle("/", http.FileServer(http.FS(web.GetFrontendAssets())))

	// Start background routine to push peer updates to Web UI
	go s.peerNotificationLoop()

	addr := fmt.Sprintf("0.0.0.0:%d", s.Port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) Stop() {
	// Dummy method for tests that call Stop()
}

func (s *Server) peerNotificationLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		peers := s.Discovery.Registry.GetAllPeers(10 * time.Second)
		msg, _ := json.Marshal(map[string]interface{}{
			"type":  "peer_update",
			"peers": peers,
		})
		s.broadcastSSE(string(msg))
	}
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.Discovery.LocalPeer)
}

func (s *Server) handlePeersList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	peers := s.Discovery.Registry.GetAllPeers(10 * time.Second)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(peers)
}

func (s *Server) handlePeersAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
		Name    string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Address == "" {
		http.Error(w, "Invalid address specified", http.StatusBadRequest)
		return
	}

	peer, err := discovery.ProbePeer(req.Address, 8990, 2*time.Second)
	if err != nil {
		http.Error(w, "Could not reach device: "+err.Error(), http.StatusBadGateway)
		return
	}

	if req.Name != "" {
		peer.Name = req.Name
	}

	if err := s.Discovery.Registry.AddSavedPeer(*peer); err != nil {
		http.Error(w, "Failed to save peer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(peer)
}

func (s *Server) handlePeersRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		var req struct {
			ID string `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		id = req.ID
	}

	if id == "" {
		http.Error(w, "Missing peer id or ip", http.StatusBadRequest)
		return
	}

	removed, err := s.Discovery.Registry.RemoveSavedPeer(id)
	if err != nil {
		http.Error(w, "Failed to remove peer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": removed,
	})
}

func (s *Server) handleSubnetScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Subnet string `json:"subnet"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	subnet := req.Subnet
	if subnet == "" {
		// Use default local IP subnet
		localIP := s.Discovery.LocalPeer.IP
		lastDot := strings.LastIndex(localIP, ".")
		if lastDot != -1 {
			subnet = localIP[:lastDot]
		} else {
			subnet = "192.168.1"
		}
	}

	discovered := discovery.ScanSubnet(subnet, 8990, 400*time.Millisecond)
	for _, p := range discovered {
		s.Discovery.Registry.AddOrUpdate(p)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(discovered)
}

// handleMultipartUpload handles browser drag & drop file/folder uploads
func (s *Server) handleMultipartUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32 MB in-memory boundary for multipart headers
	totalLen, _ := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	progressReader := transfer.NewProgressReader(r.Body, totalLen, 0, func(speedMBps float64, etaSec int, sf int64, tf int64, cb int64, tb int64) {
		percent := 0.0
		if tb > 0 {
			percent = float64(cb) / float64(tb) * 100.0
		}
		msg, _ := json.Marshal(map[string]interface{}{
			"type":      "transfer_progress",
			"percent":   percent,
			"speedMBps": speedMBps,
			"etaSec":    etaSec,
		})
		s.broadcastSSE(string(msg))
	}, nil)
	r.Body = io.NopCloser(progressReader)

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Failed to read multipart data: "+err.Error(), http.StatusBadRequest)
		return
	}

	providedPin := r.Header.Get("X-Fileshare-Pin")
	if s.ExpectedPin != "" && providedPin != s.ExpectedPin {
		http.Error(w, "Invalid authentication PIN", http.StatusUnauthorized)
		return
	}

	filesWritten := 0
	var totalBytes int64

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Error parsing stream: "+err.Error(), http.StatusBadRequest)
			return
		}

		if part.FormName() == "pin" {
			pinVal, _ := io.ReadAll(part)
			if s.ExpectedPin != "" && string(pinVal) != s.ExpectedPin {
				http.Error(w, "Invalid PIN", http.StatusUnauthorized)
				return
			}
			continue
		}

		filename := part.FileName()
		if filename == "" {
			continue
		}

		// Security: clean path and prevent path traversal
		cleanedPath := filepath.Clean(filepath.FromSlash(filename))
		if strings.HasPrefix(cleanedPath, "..") || filepath.IsAbs(cleanedPath) {
			http.Error(w, "Security violation: invalid file path", http.StatusBadRequest)
			return
		}

		destPath := filepath.Join(s.DownloadDir, cleanedPath)
		_ = os.MkdirAll(filepath.Dir(destPath), 0755)

		outFile, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		n, copyErr := io.Copy(outFile, part)
		outFile.Close()
		if copyErr != nil {
			http.Error(w, "Error writing file: "+copyErr.Error(), http.StatusInternalServerError)
			return
		}

		totalBytes += n
		filesWritten++
	}

	fmt.Printf("\n[Receiver] Received %d items (%s) saved to %s\n",
		filesWritten, transfer.FormatBytes(totalBytes), s.DownloadDir)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"filesCount": filesWritten,
		"totalBytes": totalBytes,
	})
}

// handleStreamUpload handles raw high-speed tar stream from CLI or scripts
func (s *Server) handleStreamUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providedPin := r.Header.Get("X-Fileshare-Pin")
	if s.ExpectedPin != "" && providedPin != s.ExpectedPin {
		http.Error(w, "Unauthorized: Invalid PIN", http.StatusUnauthorized)
		return
	}

	senderName := r.Header.Get("X-Fileshare-Sender")
	if senderName == "" {
		senderName = r.RemoteAddr
	}

	fmt.Printf("\n[Receiver] Incoming stream transfer from %s...\n", senderName)

	totalLen, _ := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	progressReader := transfer.NewProgressReader(r.Body, totalLen, 0, func(speedMBps float64, etaSec int, sf int64, tf int64, cb int64, tb int64) {
		percent := 0.0
		if tb > 0 {
			percent = float64(cb) / float64(tb) * 100.0
		}
		msg, _ := json.Marshal(map[string]interface{}{
			"type":      "transfer_progress",
			"percent":   percent,
			"speedMBps": speedMBps,
			"etaSec":    etaSec,
		})
		s.broadcastSSE(string(msg))
	}, nil)

	fileCount, totalBytes, err := transfer.ExtractTar(progressReader, s.DownloadDir, nil)
	if err != nil {
		fmt.Printf("[Receiver] Transfer error: %v\n", err)
		http.Error(w, "Extraction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("[Receiver] Successfully unpacked %d files (%s) into %s\n",
		fileCount, transfer.FormatBytes(totalBytes), s.DownloadDir)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"filesCount": fileCount,
		"totalBytes": totalBytes,
	})
}

// handleSSE provides real-time server-sent events for peer list updates
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	messageChan := make(chan string, 10)
	s.clientsMu.Lock()
	s.sseClients[messageChan] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.sseClients, messageChan)
		s.clientsMu.Unlock()
		close(messageChan)
	}()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (s *Server) broadcastSSE(msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for ch := range s.sseClients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (s *Server) handleFirewallConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	cmdStr := "New-NetFirewallRule -DisplayName 'FileShare TCP' -Direction Inbound -LocalPort 8990 -Protocol TCP -Action Allow; New-NetFirewallRule -DisplayName 'FileShare UDP' -Direction Inbound -LocalPort 53535 -Protocol UDP -Action Allow"
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Start-Process powershell -Verb RunAs -ArgumentList \"-Command `\"%s`\"\"", cmdStr))
	err := cmd.Run()
	if err != nil {
		http.Error(w, "Failed to run firewall command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		settings := store.DefaultSettings().Get()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(settings)
		return
	}

	if r.Method == http.MethodPost {
		var req store.AppSettings
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err := store.DefaultSettings().Update(req)
		if err != nil {
			http.Error(w, "Failed to save settings: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	// Simple implementation to list files in the DownloadDir
	entries, err := os.ReadDir(s.DownloadDir)
	if err != nil {
		http.Error(w, "Failed to read directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type FileItem struct {
		Name  string `json:"name"`
		IsDir bool   `json:"isDir"`
		Size  int64  `json:"size"`
		Time  string `json:"time"`
	}

	var history []FileItem
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		history = append(history, FileItem{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
			Time:  info.ModTime().Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(history)
}
