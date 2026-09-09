package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"fileshare/pkg/client"
	"fileshare/pkg/discovery"
	"fileshare/pkg/server"
)

func main() {
	if len(os.Args) < 2 {
		runUI(8990, "", "")
		return
	}

	command := os.Args[1]

	switch command {
	case "scan":
		runScan(os.Args[2:])

	case "add":
		runAdd(os.Args[2:])

	case "remove", "rm":
		runRemove(os.Args[2:])

	case "list", "peers":
		runList()

	case "send":
		runSend(os.Args[2:])

	case "receive":
		runReceive(os.Args[2:])

	case "ui":
		fs := flag.NewFlagSet("ui", flag.ExitOnError)
		port := fs.Int("port", 8990, "Port for Web UI and transfer server")
		dir := fs.String("dir", "", "Download directory")
		pin := fs.String("pin", "", "Optional 6-digit security PIN")
		_ = fs.Parse(os.Args[2:])
		runUI(*port, *dir, *pin)

	case "help", "--help", "-h":
		printUsage()

	default:
		// If first arg is a file or folder, treat as send directly
		if _, err := os.Stat(command); err == nil {
			runSend(os.Args[1:])
			return
		}
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("FileShare - Fast LAN & Wi-Fi File Transfer Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fileshare ui      [--port 8990] [--dir <path>]      Start local Web UI & background server")
	fmt.Println("  fileshare scan    [--subnet 192.168.1]             Scan local network or specific subnet for peers")
	fmt.Println("  fileshare add     <address> [custom-name]          Save a remote device IP permanently")
	fmt.Println("  fileshare remove  <address-or-name>                Remove a saved device")
	fmt.Println("  fileshare list                                     List all known and saved devices")
	fmt.Println("  fileshare send    <file|dir...> --to <target>       Send files/folders to a peer")
	fmt.Println("  fileshare receive [--dir <path>] [--pin <code>]     Start receiver daemon")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fileshare scan")
	fmt.Println("  fileshare scan --subnet 192.168.1")
	fmt.Println("  fileshare add 192.168.1.45:8990 Office-PC")
	fmt.Println("  fileshare send ./project/ image.png --to Office-PC")
	fmt.Println("  fileshare send ./data/ --to 192.168.1.45:8990")
	fmt.Println("  fileshare ui")
}

func runAdd(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: fileshare add <address:port> [optional-name]")
		fmt.Println("Example: fileshare add 192.168.1.50:8990 Laptop-A")
		os.Exit(1)
	}

	addr := args[0]
	customName := ""
	if len(args) > 1 {
		customName = args[1]
	}

	fmt.Printf("[Device Manager] Connecting to %s to verify...\n", addr)
	peer, err := discovery.ProbePeer(addr, 8990, 3*time.Second)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Make sure FileShare is running on that device and port 8990 is accessible.")
		os.Exit(1)
	}

	if customName != "" {
		peer.Name = customName
	}

	reg := discovery.NewPeerRegistry()
	if err := reg.AddSavedPeer(*peer); err != nil {
		fmt.Printf("Failed to save device: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSuccessfully added and saved device!\n")
	fmt.Printf("  • Name:    %s\n", peer.Name)
	fmt.Printf("  • Address: %s:%d\n", peer.IP, peer.Port)
	fmt.Printf("  • OS:      %s\n", peer.OS)
	fmt.Println("\nYou can now send files directly using its name or IP:")
	fmt.Printf("  fileshare send <files...> --to \"%s\"\n", peer.Name)
}

func runRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: fileshare remove <address-or-name>")
		fmt.Println("Example: fileshare remove 192.168.1.50")
		os.Exit(1)
	}

	target := args[0]
	reg := discovery.NewPeerRegistry()
	removed, err := reg.RemoveSavedPeer(target)
	if err != nil {
		fmt.Printf("Error removing device: %v\n", err)
		os.Exit(1)
	}

	if removed {
		fmt.Printf("Device '%s' removed from saved list.\n", target)
	} else {
		fmt.Printf("No saved device found matching '%s'. Run 'fileshare list' to see saved devices.\n", target)
	}
}

func runList() {
	reg := discovery.NewPeerRegistry()
	peers := reg.GetAllPeers(5 * time.Second)

	if len(peers) == 0 {
		fmt.Println("No saved or active peers found.")
		fmt.Println("Tip: Use 'fileshare scan' to find devices or 'fileshare add <IP:8990>' to save one.")
		return
	}

	fmt.Printf("\nKnown & Discovered Devices (%d total):\n", len(peers))
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("%-24s %-22s %-8s %-8s\n", "NAME", "ADDRESS", "TYPE", "STATUS")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, p := range peers {
		addr := fmt.Sprintf("%s:%d", p.IP, p.Port)
		peerType := "Broadcast"
		if p.IsSaved {
			peerType = "Saved"
		}
		status := "Active"
		if !p.Online {
			status = "Offline"
		}
		fmt.Printf("%-24s %-22s %-8s %-8s\n", p.Name, addr, peerType, status)
	}
	fmt.Println("--------------------------------------------------------------------------------")
}

func runScan(args []string) {
	var subnet string
	for i := 0; i < len(args); i++ {
		if (args[i] == "--subnet" || args[i] == "-s") && i+1 < len(args) {
			subnet = args[i+1]
			i++
		}
	}

	reg := discovery.NewPeerRegistry()

	if subnet != "" {
		fmt.Printf("[Discovery] Actively scanning subnet %s.1-254 for FileShare devices...\n", strings.TrimSuffix(subnet, "."))
		found := discovery.ScanSubnet(subnet, 8990, 350*time.Millisecond)
		for _, p := range found {
			reg.AddOrUpdate(p)
		}
	} else {
		fmt.Println("[Discovery] Scanning local Wi-Fi / LAN for devices running FileShare (2s)...")
		svc, err := discovery.NewService(8990, reg)
		if err == nil {
			svc.Start()
			time.Sleep(2 * time.Second)
			svc.Stop()
		}

		// Also do a fast active scan on local subnet
		localIP := discovery.GetLocalIP()
		lastDot := strings.LastIndex(localIP, ".")
		if lastDot != -1 {
			localSubnet := localIP[:lastDot]
			fmt.Printf("[Discovery] Also probing local subnet %s.x...\n", localSubnet)
			found := discovery.ScanSubnet(localSubnet, 8990, 200*time.Millisecond)
			for _, p := range found {
				reg.AddOrUpdate(p)
			}
		}
	}

	peers := reg.GetAllPeers(10 * time.Second)
	if len(peers) == 0 {
		fmt.Println("[Discovery] No active FileShare peers found.")
		fmt.Println("Tip 1: Make sure the other device is running `fileshare ui` or `fileshare receive`.")
		fmt.Println("Tip 2: If connected to a different router/subnet, use: fileshare scan --subnet <ip-prefix>")
		fmt.Println("       or add directly: fileshare add <device-ip:8990>")
		return
	}

	fmt.Printf("\nFound %d peer(s) on the network:\n", len(peers))
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("%-24s %-22s %-10s %-8s\n", "NAME", "ADDRESS", "OS", "SAVED")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, p := range peers {
		addr := fmt.Sprintf("%s:%d", p.IP, p.Port)
		savedTag := "No"
		if p.IsSaved {
			savedTag = "Yes"
		}
		fmt.Printf("%-24s %-22s %-10s %-8s\n", p.Name, addr, p.OS, savedTag)
	}
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("To send to a peer, run:")
	fmt.Println("  fileshare send <path...> --to <ADDRESS-or-NAME>")
}

func runSend(args []string) {
	var paths []string
	var target string
	var pin string

	for i := 0; i < len(args); i++ {
		if args[i] == "--to" && i+1 < len(args) {
			target = args[i+1]
			i++
		} else if args[i] == "--pin" && i+1 < len(args) {
			pin = args[i+1]
			i++
		} else if !strings.HasPrefix(args[i], "--") {
			paths = append(paths, args[i])
		}
	}

	if len(paths) == 0 {
		fmt.Println("Error: Please specify at least one file or directory to send.")
		fmt.Println("Example: fileshare send ./Documents/ photo.jpg --to 192.168.1.50:8990")
		os.Exit(1)
	}

	reg := discovery.NewPeerRegistry()

	// Check if target is a known device name or alias
	if target != "" {
		peers := reg.GetAllPeers(10 * time.Minute)
		for _, p := range peers {
			if strings.EqualFold(p.Name, target) || strings.EqualFold(p.ID, target) {
				resolved := fmt.Sprintf("%s:%d", p.IP, p.Port)
				fmt.Printf("Resolved device '%s' -> %s\n", target, resolved)
				target = resolved
				break
			}
		}
	}

	if target == "" {
		// Scan to see if there's only one peer available
		fmt.Println("No target specified with --to. Checking known and nearby peers...")
		peers := reg.GetAllPeers(10 * time.Second)
		if len(peers) == 1 {
			target = fmt.Sprintf("%s:%d", peers[0].IP, peers[0].Port)
			fmt.Printf("Auto-selected available peer: %s (%s)\n", peers[0].Name, target)
		} else if len(peers) > 1 {
			fmt.Println("Multiple devices found. Please specify one with --to <IP:PORT or NAME>:")
			for _, p := range peers {
				fmt.Printf("  - %s (%s:%d)\n", p.Name, p.IP, p.Port)
			}
			os.Exit(1)
		} else {
			fmt.Println("No peers found. Please specify target address with --to <IP:PORT>.")
			os.Exit(1)
		}
	}

	// Ensure port if only IP is provided
	if !strings.Contains(target, ":") {
		target = target + ":8990"
	}

	hostname, _ := os.Hostname()
	if err := client.SendPaths(target, paths, pin, hostname); err != nil {
		fmt.Printf("\nTransfer error: %v\n", err)
		os.Exit(1)
	}
}

func runReceive(args []string) {
	fs := flag.NewFlagSet("receive", flag.ExitOnError)
	port := fs.Int("port", 8990, "Port to listen on")
	dir := fs.String("dir", "", "Download directory")
	pin := fs.String("pin", "", "Optional 6-digit security PIN")
	autoAccept := fs.Bool("auto-accept", true, "Automatically accept incoming transfers")
	_ = fs.Parse(args)

	reg := discovery.NewPeerRegistry()
	disc, err := discovery.NewService(*port, reg)
	if err != nil {
		fmt.Printf("Error creating discovery: %v\n", err)
		os.Exit(1)
	}
	disc.Start()
	defer disc.Stop()

	srv := server.NewServer(*port, *dir, disc, *autoAccept, *pin)
	fmt.Printf("[Receiver] Daemon started on port %d. Ready to receive files.\n", *port)
	if err := srv.Start(); err != nil {
		fmt.Printf("Server exited: %v\n", err)
	}
}

func runUI(port int, dir string, pin string) {
	reg := discovery.NewPeerRegistry()
	disc, err := discovery.NewService(port, reg)
	if err != nil {
		fmt.Printf("Error creating discovery: %v\n", err)
		os.Exit(1)
	}
	disc.Start()
	defer disc.Stop()

	srv := server.NewServer(port, dir, disc, true, pin)

	localURL := fmt.Sprintf("http://localhost:%d", port)
	networkURL := fmt.Sprintf("http://%s:%d", disc.LocalPeer.IP, port)

	fmt.Println("==================================================================")
	fmt.Println("         FileShare - High-Speed LAN & Wi-Fi Sharing               ")
	fmt.Println("==================================================================")
	fmt.Printf("  • Local Dashboard:    %s\n", localURL)
	fmt.Printf("  • Wi-Fi Network URL:  %s\n", networkURL)
	fmt.Println("------------------------------------------------------------------")
	fmt.Println("  Opening Web UI in your default browser...")
	fmt.Println("  Press Ctrl+C to stop.")
	fmt.Println("==================================================================")

	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(localURL)
	}()

	if err := srv.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
