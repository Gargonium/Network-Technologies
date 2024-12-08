package socks5

import (
	"fmt"
	"github.com/lesismal/nbio"
	"log"
	"net"
)

// Client Greeting
// Server Choice
// Client Connection request
// Response packet from server

// SOCKS5 address

func StartServer(portNum int) error {

	engine := nbio.NewEngine(nbio.Config{
		Network: "tcp",
		Addrs:   []string{fmt.Sprintf(":%d", portNum)},
	})

	engine.OnOpen(func(c *nbio.Conn) {
		log.Printf("Новый клиент подключился: %s\n", c.RemoteAddr().String())
		handleClient(c, portNum)
	})

	err := engine.Start()

	if err != nil {
		return fmt.Errorf("Ошибка при запуске сервера: %v\n", err)
	}
	defer engine.Stop()

	<-make(chan int)

	return nil
}

func handleClient(c *nbio.Conn, portNum int) {
	defer c.Close()

	buf := make([]byte, 262)

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

	if buf[3] != 0x01 {
		log.Println("SOCKS5 address must be IPv4")
		fmt.Printf("%x\n", buf[3])
		return
	}

	targetAddr := fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
	targetPort := uint16(buf[8])<<8 | uint16(buf[9])

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", targetAddr, targetPort))
	if err != nil {
		log.Println("Ошибка подключения к целевому серверу:", err)
		return
	}
	defer serverConn.Close()

	ip := []byte{0x7F, 0x00, 0x00, 0x01}
	port := []byte{byte(portNum >> 8), byte(portNum & 0xFF)}

	// Response packet from server
	c.Write([]byte{
		0x05,                       // Версия SOCKS5
		0x00,                       // Статус (успешная аутентификация)
		0x00,                       // Резерв
		0x01,                       // Тип адреса (IPv4)
		ip[0], ip[1], ip[2], ip[3], // Адрес (127.0.0.1)
		port[0], port[1], // Порт
	})

	for {
		n, err := serverConn.Read(buf)
		if err != nil {
			log.Println("Ошибка при чтении с сервера:", err)
			return
		}
		_, err = c.Write(buf[:n])
		if err != nil {
			log.Println("Ошибка при отправке данных клиенту:", err)
			return
		}

		n, err = c.Read(buf)
		if err != nil {
			log.Println("Ошибка при чтении от клиента:", err)
			return
		}
		_, err = serverConn.Write(buf[:n])
		if err != nil {
			log.Println("Ошибка при отправке данных на сервер:", err)
			return
		}
	}
}
