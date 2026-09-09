package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"
)

const (
	DefaultBroadcastPort = 53535
	DefaultBroadcastAddr = "255.255.255.255"
	BeaconInterval       = 2 * time.Second
	PeerTimeout          = 6 * time.Second
)

// BeaconPacket represents the broadcasted announcement message.
type BeaconPacket struct {
	Magic    string `json:"magic"` // "FILESHARE_BEACON"
	ID       string `json:"id"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
	OS       string `json:"os"`
}

// Service manages background discovery via UDP broadcast.
type Service struct {
	LocalPeer Peer
	Registry  *PeerRegistry
	stopChan  chan struct{}
}

// NewService creates a discovery service with local device details.
func NewService(port int, registry *PeerRegistry) (*Service, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Device-" + fmt.Sprintf("%d", time.Now().Unix()%10000)
	}

	localIP := GetLocalIP()
	id := fmt.Sprintf("%s-%d", hostname, port)

	localPeer := Peer{
		ID:   id,
		Name: hostname,
		IP:   localIP,
		Port: port,
		OS:   runtime.GOOS,
	}

	return &Service{
		LocalPeer: localPeer,
		Registry:  registry,
		stopChan:  make(chan struct{}),
	}, nil
}

// Start begins broadcasting and listening for peers.
func (s *Service) Start() {
	go s.listen()
	go s.broadcastLoop()
}

// Stop terminates the discovery service.
func (s *Service) Stop() {
	close(s.stopChan)
}

func (s *Service) broadcastLoop() {
	ticker := time.NewTicker(BeaconInterval)
	defer ticker.Stop()

	// Initial broadcast immediately
	s.broadcast()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.broadcast()
		}
	}
}

func (s *Service) broadcast() {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", DefaultBroadcastAddr, DefaultBroadcastPort))
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()

	packet := BeaconPacket{
		Magic: "FILESHARE_BEACON",
		ID:    s.LocalPeer.ID,
		Name:  s.LocalPeer.Name,
		Port:  s.LocalPeer.Port,
		OS:    s.LocalPeer.OS,
	}

	data, err := json.Marshal(packet)
	if err != nil {
		return
	}

	_, _ = conn.Write(data)
}

func (s *Service) listen() {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%d", DefaultBroadcastPort))
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		// Port may be taken or firewalled; non-fatal, manual IP will still work
		return
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		select {
		case <-s.stopChan:
			return
		default:
			_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			var packet BeaconPacket
			if err := json.Unmarshal(buf[:n], &packet); err != nil {
				continue
			}

			if packet.Magic != "FILESHARE_BEACON" || packet.ID == s.LocalPeer.ID {
				// Ignore non-fileshare packets or self broadcasts
				continue
			}

			s.Registry.AddOrUpdate(Peer{
				ID:   packet.ID,
				Name: packet.Name,
				IP:   remoteAddr.IP.String(),
				Port: packet.Port,
				OS:   packet.OS,
			})
		}
	}
}

// GetLocalIP attempts to find the primary non-loopback IPv4 address.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
