package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fileshare/pkg/discovery"
	"fileshare/pkg/transfer"
	"fileshare/web"
)

// Server handles HTTP API, Web UI, and incoming file transfers.
type Server struct {
	Port         int
	DownloadDir  string
	AutoAccept   bool
	ExpectedPin  string
	Discovery    *discovery.Service
	httpServer   *http.Server
	clientsMu    sync.Mutex
	sseClients   map[chan string]bool
}

// NewServer initializes a file share server instance.
func NewServer(port int, downloadDir string, disc *discovery.Service, autoAccept bool, pin string) *Server {
	if downloadDir == "" {
		home, _ := os.UserHomeDir()
		downloadDir = filepath.Join(home, "Downloads", "FileShare")
	}
	_ = os.MkdirAll(downloadDir, 0755)

	return &Server{
		Port:        port,
		DownloadDir: downloadDir,
		AutoAccept:  autoAccept,
		ExpectedPin: pin,
		Discovery:   disc,
		sseClients:  make(map[chan string]bool),
	}
}

// Start runs the HTTP server on the configured port.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 1. Web UI Static Files
	fileServer := http.FileServer(web.GetFileSystem())
	mux.Handle("/", fileServer)

	// 2. API Endpoints
	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/peers", s.handlePeers)
	mux.HandleFunc("/api/peers/add", s.handlePeersAdd)
	mux.HandleFunc("/api/peers/remove", s.handlePeersRemove)
	mux.HandleFunc("/api/peers/scan", s.handleSubnetScan)
	mux.HandleFunc("/api/upload", s.handleMultipartUpload)
	mux.HandleFunc("/api/stream-upload", s.handleStreamUpload)
	mux.HandleFunc("/events", s.handleSSE)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	fmt.Printf("[Server] FileShare listening on http://%s:%d\n", s.Discovery.LocalPeer.IP, s.Port)
	fmt.Printf("[Server] Destination folder: %s\n", s.DownloadDir)

	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the server.
func (s *Server) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"deviceName":  s.Discovery.LocalPeer.Name,
		"ip":          s.Discovery.LocalPeer.IP,
		"port":        s.Port,
		"os":          s.Discovery.LocalPeer.OS,
		"downloadDir": s.DownloadDir,
	})
}

func (s *Server) handlePeers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	peers := s.Discovery.Registry.GetAllPeers(discovery.PeerTimeout)
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

	fileCount, totalBytes, err := transfer.ExtractTar(r.Body, s.DownloadDir, nil)
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
