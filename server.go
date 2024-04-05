package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	connPort = ":8888"
	connType = "tcp"
)

var (
	clients     = make(map[net.Conn]*Client)
	clientsLock sync.Mutex
	chatRooms   = map[string][]net.Conn{
		"AITU": {},
		"NU":   {},
		"ENU":  {},
	}
	chatRoomsLock sync.Mutex
)

type Client struct {
	Connection net.Conn
	Username   string
}

func main() {
	listener, err := net.Listen(connType, connPort)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("Listening on " + connPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err)
			continue
		}

		client := &Client{Connection: conn}
		clientsLock.Lock()
		clients[conn] = client
		clientsLock.Unlock()

		go handleClient(client)
	}
}

func handleClient(client *Client) {
	defer client.Connection.Close()

	log.Println("New client connected:", client.Connection.RemoteAddr())

	scanner := bufio.NewScanner(client.Connection)
	for scanner.Scan() {
		message := scanner.Text()
		log.Println("Received from", client.Connection.RemoteAddr(), ":", message)

		if client.Username == "" {
			// First message from client is assumed to be the username
			client.Username = message
			log.Println("Username set for", client.Connection.RemoteAddr(), ":", message)
		} else if strings.HasPrefix(message, "/join") {
			joinChatRoom(client, message)
		} else {
			// Если сообщение не начинается с /join, считаем, что это текст сообщения, и передаем его всем клиентам в комнате
			broadcastMessage(client, message)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading from", client.Connection.RemoteAddr(), ":", err)
	}

	clientsLock.Lock()
	delete(clients, client.Connection)
	clientsLock.Unlock()

	log.Println("Client", client.Connection.RemoteAddr(), "disconnected.")
}

func joinChatRoom(client *Client, message string) {
	parts := strings.SplitN(message, " ", 2)
	if len(parts) < 2 {
		log.Println("Invalid /join command from", client.Connection.RemoteAddr())
		return
	}

	room := parts[1]

	chatRoomsLock.Lock()
	defer chatRoomsLock.Unlock()

	_, ok := chatRooms[room]
	if !ok {
		log.Println("Chat room does not exist:", room)
		return
	}

	chatRooms[room] = append(chatRooms[room], client.Connection)

	log.Printf("Client %s joined chat room %s\n", client.Connection.RemoteAddr(), room)
	sendMessageToClient(client, "Joined chat room: "+room+"\n") // Отправляем клиенту подтверждение о входе в чат
}

func broadcastMessage(sender *Client, message string) {
	chatRoomsLock.Lock()
	defer chatRoomsLock.Unlock()

	currentRoom := ""

	// Находим текущую комнату отправителя
	for room, clients := range chatRooms {
		for _, client := range clients {
			if client == sender.Connection {
				currentRoom = room
				break
			}
		}
		if currentRoom != "" {
			break
		}
	}

	if currentRoom == "" {
		log.Println("User not in any room")
		return
	}

	// Отправляем сообщение только пользователям в текущей комнате
	for _, client := range chatRooms[currentRoom] {
		if client != sender.Connection {
			sendMessageToClient(&Client{Connection: client}, message+"\n")
		}
	}
}

func sendMessageToClient(client *Client, message string) {
	_, err := client.Connection.Write([]byte(message))
	if err != nil {
		log.Println("Error sending message to", client.Connection.RemoteAddr(), ":", err)
	}
}
