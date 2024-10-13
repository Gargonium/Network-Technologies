package main

// CLIENT

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

func main() {

	if len(os.Args) != 4 {
		fmt.Println("Using: go run main.go <path to file> <IP-serverIP or DNS-name> <serverPort>")
		return
	}

	filePath := os.Args[1]
	serverIP := os.Args[2]
	serverPort := os.Args[3]

	fmt.Println("File:", filePath)

	tcpServer, err := net.ResolveTCPAddr("tcp", serverIP+":"+serverPort)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer func(conn *net.TCPConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// Открыть файл
	file, err := os.Open(filepath.Join("client", filePath))
	if err != nil {
		fmt.Println("Error opening file:", err)
		log.Fatal(err)
	}

	// Получить инфу о файле
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error opening file:", err)
		log.Fatal(err)
	}

	// Отправить имя файла
	_, err = conn.Write([]byte(filepath.Base(filePath) + "\n"))
	if err != nil {
		fmt.Println("Error sending fileName:", err)
		log.Fatal(err)
	}

	// Отправить размер файла
	_, err = conn.Write([]byte(fmt.Sprintf("%d\n", fileInfo.Size())))
	if err != nil {
		fmt.Println("Error sending fileSize:", err)
		log.Fatal(err)
	}

	// Отправить файл
	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file:", err)
		log.Fatal(err)
	}

	// Получить ответ от сервера
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Server response:", string(buf[:n]))
}
