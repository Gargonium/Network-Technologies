package socks5

import (
	"fmt"
	"log"
	"net"
	"noSOCKS/dns"
)

func closeConnection(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		log.Printf("Error closing connection: %v\n", err)
	}
}

func HandleClient(c net.Conn, portNum int) {

	buf := make([]byte, 1024)

	// Client Greeting
	_, err := c.Read(buf)
	if err != nil {
		log.Println("Ошибка чтения запроса:", err)
		closeConnection(c)
		return
	}

	if buf[0] != 0x05 {
		log.Println("Неверная версия SOCKS")
		closeConnection(c)
		return
	}

	// Server Choice
	_, err = c.Write([]byte{0x05, 0x00})
	if err != nil {
		log.Println("Server Choice:", err)
		closeConnection(c)
		return
	}

	// Client Connection request
	_, err = c.Read(buf[:])
	if err != nil {
		log.Println("Ошибка чтения команды:", err)
		closeConnection(c)
		return
	}

	if buf[0] != 0x05 {
		log.Println("Неверная версия SOCKS")
		closeConnection(c)
		return
	}

	if buf[1] != 0x01 {
		log.Println("Поддерживается только CONNECT запрос")
		closeConnection(c)
		return
	}

	if buf[2] != 0x00 {
		log.Println("RSV must be 0x00")
		closeConnection(c)
		return
	}

	var targetAddr string
	var targetPort uint16

	if buf[3] == 0x01 { // IPv4
		targetAddr = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
		targetPort = uint16(buf[8])<<8 | uint16(buf[9])
	} else if buf[3] == 0x03 { // Доменное имя
		nameLen := buf[4]
		domainName := string(buf[5 : 5+nameLen])
		targetPort = uint16(buf[5+nameLen])<<8 | uint16(buf[6+nameLen])

		targetAddr, err = dns.ResolveDNSName(domainName)
		if err != nil {
			log.Println("Ошибка разрешения DNS:", err)
			closeConnection(c)
			return
		}
	} else {
		log.Println("Address must be IPv4 or DNS name")
		closeConnection(c)
		return
	}

	ip := []byte{0x7F, 0x00, 0x00, 0x01}
	port := []byte{byte(portNum >> 8), byte(portNum & 0xFF)}

	// Response packet from server
	_, err = c.Write([]byte{
		0x05,                       // Версия SOCKS5
		0x00,                       // Статус (спешная аутентификация)
		0x00,                       // Резерв
		0x01,                       // Тип адреса (IPv4)
		ip[0], ip[1], ip[2], ip[3], // Адрес (127.0.0.1)
		port[0], port[1], // Порт
	})
	if err != nil {
		log.Println("Response packet from server write error:", err)
		closeConnection(c)
		return
	}

	target := fmt.Sprintf("%s:%d", targetAddr, targetPort)

	serverConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Println("Error connecting to server:", err)
		closeConnection(c)
		return
	}

	EstablishConnection(c, serverConn)
}
