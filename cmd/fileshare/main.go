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
		runScan()

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
	fmt.Println("  fileshare ui      [--port 8990] [--dir <path>]     Start local Web UI & background server")
	fmt.Println("  fileshare scan                                     Scan local network for active peers")
	fmt.Println("  fileshare send    <file|dir...> --to <target>      Send files/folders to a peer")
	fmt.Println("  fileshare receive [--dir <path>] [--pin <code>]    Start receiver daemon")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fileshare scan")
	fmt.Println("  fileshare send ./project/ image.png --to 192.168.1.45:8990")
	fmt.Println("  fileshare receive --dir ~/Downloads")
	fmt.Println("  fileshare ui")
}

func runScan() {
	fmt.Println("[Discovery] Scanning local Wi-Fi / LAN for devices running FileShare (3s)...")
	reg := discovery.NewPeerRegistry()
	svc, err := discovery.NewService(8990, reg)
	if err != nil {
		fmt.Printf("Error starting discovery: %v\n", err)
		return
	}
	svc.Start()
	defer svc.Stop()

	time.Sleep(3 * time.Second)

	peers := reg.GetActivePeers(10 * time.Second)
	if len(peers) == 0 {
		fmt.Println("[Discovery] No active FileShare peers found on this subnet.")
		fmt.Printf("Tip: Make sure the other device is on the same Wi-Fi and running `fileshare ui` or `fileshare receive`.\n")
		return
	}

	fmt.Printf("\nFound %d peer(s) on the network:\n", len(peers))
	fmt.Println("----------------------------------------------------------------------")
	fmt.Printf("%-24s %-22s %-10s\n", "NAME", "ADDRESS", "OS")
	fmt.Println("----------------------------------------------------------------------")
	for _, p := range peers {
		addr := fmt.Sprintf("%s:%d", p.IP, p.Port)
		fmt.Printf("%-24s %-22s %-10s\n", p.Name, addr, p.OS)
	}
	fmt.Println("----------------------------------------------------------------------")
	fmt.Println("To send to a peer, run:")
	fmt.Println("  fileshare send <path...> --to <ADDRESS>")
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

	if target == "" {
		// Scan to see if there's only one peer available
		fmt.Println("No target specified with --to. Scanning network for available peers...")
		reg := discovery.NewPeerRegistry()
		svc, _ := discovery.NewService(8990, reg)
		svc.Start()
		time.Sleep(2 * time.Second)
		svc.Stop()

		peers := reg.GetActivePeers(10 * time.Second)
		if len(peers) == 1 {
			target = fmt.Sprintf("%s:%d", peers[0].IP, peers[0].Port)
			fmt.Printf("Auto-selected sole available peer: %s (%s)\n", peers[0].Name, target)
		} else if len(peers) > 1 {
			fmt.Println("Multiple peers found. Please specify one with --to <IP:PORT>:")
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
