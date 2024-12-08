package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	proxyAddr := "localhost:8886"

	// Client Greeting
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		log.Fatalf("Ошибка подключения к SOCKS5 серверу: %v\n", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		log.Fatalf("Ошибка отправки запроса аутентификации: %v\n", err)
	}

	// Server Choice
	response := make([]byte, 2)
	_, err = conn.Read(response)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа от сервера: %v\n", err)
	}
	if response[0] != 0x05 || response[1] != 0x00 {
		log.Fatalf("Ошибка аутентификации\n")
	}

	// 3. Отправляем запрос на соединение с целевым сервером (CONNECT)
	// Версия SOCKS5 (0x05), команда CONNECT (0x01), RSV (0x00), тип адреса IPv4 (0x01)
	// Адрес целевого сервера (example.com), порт 80
	// Client Connection request
	//targetAddrBytes := net.ParseIP("172.217.10.68").To4()
	//if targetAddrBytes == nil {
	//	log.Fatalf("Не удалось преобразовать адрес целевого сервера\n")
	//}

	targetAddrBytes := []byte{0x5D, 0xB8, 0xD8, 0x22}

	port := uint16(80)
	_, err = conn.Write([]byte{
		0x05, // SOCKS5 версия
		0x01, // Команда CONNECT
		0x00, // RSV
		0x01, // Тип адреса (IPv4)
		targetAddrBytes[0], targetAddrBytes[1], targetAddrBytes[2], targetAddrBytes[3],
		byte(port >> 8), byte(port & 0xFF),
	})
	if err != nil {
		log.Fatalf("Ошибка отправки запроса на подключение: %v\n", err)
	}

	// 4. Получаем ответ от прокси сервера (ожидаем успешное подключение)
	connResponse := make([]byte, 10)
	_, err = conn.Read(connResponse)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа от прокси сервера: %v\n", err)
	}
	if connResponse[1] != 0x00 {
		log.Fatalf("Ошибка при подключении к целевому серверу через прокси\n")
	}

	// 5. Установлено соединение, теперь можно передавать данные
	// Например, отправляем HTTP-запрос на целевой сервер
	httpRequest := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	_, err = conn.Write([]byte(httpRequest))
	if err != nil {
		log.Fatalf("Ошибка отправки HTTP запроса: %v\n", err)
	}

	// 6. Читаем ответ от целевого сервера через прокси
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа: %v\n", err)
	}

	// Выводим полученные данные (HTTP ответ)
	fmt.Printf("Ответ от сервера:\n%s\n", string(buffer[:n]))
}
