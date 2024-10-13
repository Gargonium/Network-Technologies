package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	message = "AreYouMyCopy?"
	timeout = 5 * time.Second
)

var (
	peers     = make(map[string]time.Time)
	mutex     = &sync.Mutex{}
	localAddr string
)

func sendDiscoveryMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	for {
		_, err := conn.WriteToUDP([]byte(message), addr)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(2 * time.Second)
	}
}

func listenForResponses(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if string(buffer[:n]) == message && remoteAddr.String() != localAddr {
			isNew := false
			if peers[remoteAddr.IP.String()].IsZero() {
				isNew = true
			}
			mutex.Lock()
			peers[remoteAddr.IP.String()] = time.Now()
			mutex.Unlock()
			if isNew {
				showActivePeers()
			}
		}
	}
}

func monitorPeers() {
	for {
		time.Sleep(1 * time.Second)
		mutex.Lock()
		now := time.Now()
		updated := false
		for addr, lastSeen := range peers {
			if now.Sub(lastSeen) > timeout {
				delete(peers, addr)
				updated = true
			}
		}
		mutex.Unlock()
		if updated {
			showActivePeers()
		}
	}
}

func showActivePeers() {
	mutex.Lock()
	defer mutex.Unlock()
	active := make([]string, 0, len(peers))
	active = append(active, localAddr)
	for addr := range peers {
		active = append(active, addr)
	}
	fmt.Println("Active nodes:", active)
}

func getLocalAddress() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal("Ошибка при получении IP адресов:", err)
	}

	for _, addr := range addresses {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}

		if ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		} else if ipNet.IP.To16() != nil {
			return ipNet.IP.String()
		}
	}
	log.Fatal("Не удалось найти локальный IP адрес")
	return ""
}

func getMulticastAddress(address string) (*net.UDPAddr, error) {
	if strings.Contains(address, ".") {
		return net.ResolveUDPAddr("udp4", address)
	}
	return net.ResolveUDPAddr("udp6", address)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Using: go run main.go <multicast-address>")
		return
	}
	multicastAddress := os.Args[1] // "224.0.0.1"

	localAddr = getLocalAddress()
	fmt.Println("This is my name! -->", localAddr)
	addr, err := getMulticastAddress("[" + multicastAddress + "]" + ":8886")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	go sendDiscoveryMessage(conn, addr)
	go listenForResponses(conn)
	go monitorPeers()
	showActivePeers()

	select {}
}
