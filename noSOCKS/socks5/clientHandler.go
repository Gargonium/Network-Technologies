package socks5

import (
	"fmt"
	"github.com/lesismal/nbio"
	"io"
	"log"
	"noSOCKS/dns"
	"time"
)

func HandleClient(c *nbio.Conn, portNum int) {

	buf := make([]byte, 1024)

	// Client Greeting
	_, err := c.Read(buf)
	if err != nil {
		log.Println("Ошибка чтения запроса:", err)
		return
	}

	if buf[0] != 0x05 {
		log.Println("Неверная версия SOCKS")
		return
	}

	// Server Choice
	c.Write([]byte{0x05, 0x00})

	// Client Connection request
	_, err = c.Read(buf[:])
	if err != nil {
		log.Println("Ошибка чтения команды:", err)
		return
	}

	if buf[0] != 0x05 {
		log.Println("Неверная версия SOCKS")
		return
	}

	if buf[1] != 0x01 {
		log.Println("Поддерживается только CONNECT запрос")
		return
	}

	if buf[2] != 0x00 {
		log.Println("RSV must be 0x00")
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
			c.Write([]byte{
				0x05, 0x04, // Код ошибки: Host unreachable
				0x00, 0x01, // Тип адреса
				0x00, 0x00, 0x00, 0x00, // Адрес
				0x00, 0x00, // Порт
			})
			return
		}
	} else {
		log.Println("Address must be IPv4 or DNS name")
		return
	}

	ip := []byte{0x7F, 0x00, 0x00, 0x01}
	port := []byte{byte(portNum >> 8), byte(portNum & 0xFF)}

	// Response packet from server
	c.Write([]byte{
		0x05,                       // Версия SOCKS5
		0x00,                       // Статус (спешная аутентификация)
		0x00,                       // Резерв
		0x01,                       // Тип адреса (IPv4)
		ip[0], ip[1], ip[2], ip[3], // Адрес (127.0.0.1)
		port[0], port[1], // Порт
	})

	target := fmt.Sprintf("%s:%d", targetAddr, targetPort)

	serverConn, err := nbio.Dial("tcp", target)
	if err != nil {
		log.Println("Error connecting to server:", err)
		return
	}

	serverEngine := nbio.NewEngine(nbio.Config{})
	serverEngine.Start()

	_, err = serverEngine.AddConn(serverConn)
	if err != nil {
		log.Println("Error adding connection:", err)
		return
	}

	defer c.Close()
	defer serverConn.Close()

	for {
		forwardData(c, serverConn)
		forwardData(serverConn, c)
		time.Sleep(100 * time.Millisecond)
	}
}

func forwardData(src *nbio.Conn, dst *nbio.Conn) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Println("Error copying data:", err)
		return
	}
}
