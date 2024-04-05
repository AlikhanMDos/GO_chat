package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	connPort      = ":8888"
	connType      = "tcp"
	msgDisconnect = "Disconnected from the server.\n"
	HelpCommand   = "/help"
	JoinCommand   = "/join"
	ExitCommand   = "/exit"
)

var (
	wg          sync.WaitGroup
	username    string
	currentRoom string
)

func main() {
	wg.Add(1)

	conn, err := net.Dial(connType, connPort)
	if err != nil {
		fmt.Println(err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your username: ")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)

	go read(conn)
	go write(conn, reader)

	wg.Wait()
}

func read(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf(msgDisconnect)
			wg.Done()
			return
		}
		fmt.Print(str) // Выводим сообщение от сервера
	}
}

func write(conn net.Conn, reader *bufio.Reader) {
	writer := bufio.NewWriter(conn)
	defer conn.Close()

	printHelpMessage(writer) // Display help message

	for {
		if currentRoom == "" {
			joinRoom(conn, writer)
		}

		fmt.Print("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		str = strings.TrimSpace(str)

		if str == HelpCommand {
			printHelpMessage(writer)
		} else if strings.HasPrefix(str, JoinCommand) {
			joinRoom(conn, writer)
		} else if str == ExitCommand {
			_, _ = fmt.Fprintf(writer, "%s has left the chat.\n", username)
			writer.Flush()
			wg.Done()
			return
		} else {
			sendMessage(conn, writer, str)
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func printHelpMessage(writer *bufio.Writer) {
	fmt.Fprintln(writer, "List of commands:")
	fmt.Fprintln(writer, "/help: Display this help message")
	fmt.Fprintln(writer, "/join [room]: Join the specified chat room")
	fmt.Fprintln(writer, "/exit: Exit the chat")
}

func joinRoom(conn net.Conn, writer *bufio.Writer) {
	fmt.Println("Available chat rooms:")
	fmt.Println("1. AITU")
	fmt.Println("2. NU")
	fmt.Println("3. ENU")
	fmt.Print("Choose chat room (1, 2, or 3): ")

	reader := bufio.NewReader(os.Stdin)
	str, _ := reader.ReadString('\n')
	str = strings.TrimSpace(str)

	var roomName string
	switch str {
	case "1":
		roomName = "AITU"
	case "2":
		roomName = "NU"
	case "3":
		roomName = "ENU"
	default:
		fmt.Fprintf(writer, "Invalid choice\n")
		return
	}

	currentRoom = roomName
	_, _ = fmt.Fprintf(writer, "Joined chat room: %s\n", currentRoom)

	// Отправляем команду серверу о входе в чат
	_, _ = fmt.Fprintf(writer, "%s\n", JoinCommand+" "+currentRoom)
	writer.Flush()
}

func sendMessage(conn net.Conn, writer *bufio.Writer, message string) {
	_, err := fmt.Fprintf(writer, "%s: %s\n", username, message)
	if err != nil {
		fmt.Println("Error sending message:", err)
		os.Exit(1)
	}
}
