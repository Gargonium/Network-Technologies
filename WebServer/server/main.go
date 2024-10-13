package main

// SERVER

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Using: go run main.go <port>")
		return
	}

	port := os.Args[1]

	address := getLocalAddress() + ":" + port

	err := os.MkdirAll("server/uploads", 0750)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Starting server on", address)

	listen, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		log.Fatal(err)
	}
	defer listen.Close()

	var wg sync.WaitGroup

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			log.Fatal(err)
		}
		wg.Add(1)
		go handleConnection(conn, &wg)
	}

	wg.Wait()
}

func handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	fmt.Println("\n\nClient connected:", conn.RemoteAddr())
	defer conn.Close()
	defer wg.Done()

	reader := bufio.NewReader(conn)

	// Принять имя файла
	fileName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading fileName:", err.Error())
		return
	}
	fileName = fileName[:len(fileName)-1]

	fileSizeStr, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading fileSizeStr:", err.Error())
		return
	}
	fileSizeStr = fileSizeStr[:len(fileSizeStr)-1]

	fileSize, err := strconv.Atoi(fileSizeStr)
	if err != nil {
		fmt.Println("Error parsing fileSizeStr:", err.Error())
		return
	}

	filePath := getAvailableFileName("server/uploads", fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err.Error())
		return
	}
	defer outFile.Close()

	var totalBytesReceived int64
	start := time.Now()
	quit := make(chan struct{})
	go func() {
		countTicker := time.NewTicker(100 * time.Millisecond)
		printTicker := time.NewTicker(3 * time.Second)
		defer countTicker.Stop()
		defer printTicker.Stop()

		instantSpeed := 0.0
		lastBytesReceived := int64(0)

		for {
			select {
			case <-countTicker.C:
				bytesReceived := totalBytesReceived - lastBytesReceived
				instantSpeed = float64(bytesReceived * 10)
				lastBytesReceived = totalBytesReceived
			case <-printTicker.C:
				elapsed := time.Since(start).Seconds()
				averageSpeed := float64(totalBytesReceived) / elapsed
				fmt.Printf("Client: %s, Average Speed: %.2f bytes/sec, Instant Speed: %.2f bytes/sec, Total received: %d bytes\n", conn.RemoteAddr(), averageSpeed, instantSpeed, totalBytesReceived)
			case <-quit:
				elapsed := time.Since(start).Seconds()
				var averageSpeed float64
				if elapsed > 0 {
					averageSpeed = float64(totalBytesReceived) / elapsed
					if instantSpeed == 0 {
						instantSpeed = averageSpeed
					}
				} else {
					averageSpeed = instantSpeed
				}
				fmt.Printf("Client: %s, Average Speed: %.2f bytes/sec, Instant Speed: %.2f bytes/sec, Total received: %d bytes\n", conn.RemoteAddr(), averageSpeed, instantSpeed, totalBytesReceived)
				fmt.Printf("Time elapsed: %.2f seconds\n", elapsed)
				return
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			fmt.Println("Error reading file:", err.Error())
			break
		}
		totalBytesReceived += int64(n)
		outFile.Write(buf[:n])
		if totalBytesReceived == int64(fileSize) {
			break
		}
	}

	close(quit)

	time.Sleep(10 * time.Millisecond)

	if totalBytesReceived == int64(fileSize) {
		fmt.Println("File received successfully:", filePath)
		conn.Write([]byte("Success\n"))
	} else {
		fmt.Println("File size mismatch:", fileName)
		conn.Write([]byte("Failure\n"))
		outFile.Close()
		os.Remove(filePath)
	}

	fmt.Println()
}

func getAvailableFileName(directory, fileName string) string {
	// Начальное имя файла
	fullPath := filepath.Join(directory, fileName)

	// Если файла не существует, возвращаем это имя
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fullPath
	}

	// Разделим имя файла и расширение
	ext := filepath.Ext(fileName)
	name := fileName[:len(fileName)-len(ext)]

	// Добавляем суффикс к имени файла
	for i := 1; ; i++ {
		newFileName := fmt.Sprintf("%s%d%s", name, i, ext)
		fullPath = filepath.Join(directory, newFileName)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fullPath
		}
	}
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
		}
	}
	log.Fatal("Не удалось найти локальный IP адрес")
	return ""
}
