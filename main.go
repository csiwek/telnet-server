package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// ##### 1

type client struct {
	Conn net.Conn
	Name string
}

var clients = make(map[string]*client)
var messages = make(chan string)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:1234")
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()
	go broadcastMsg()
	for {
		con, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(con)
	}

}

func handleConn(con net.Conn) {
	defer con.Close()
	clientReader := bufio.NewReader(con)
	addr := con.RemoteAddr().String()
	client := &client{
		Conn: con,
		Name: "User" + fmt.Sprintf("%d", len(clients)+1),
	}
	clients[addr] = client
	// ##### 1
	defer handleDisconn(addr)

	for {
		clientMsg, err := clientReader.ReadString('\n')
		if err == nil {
			clientMsg := strings.TrimSpace(clientMsg)

			if clientMsg == "/quit" {
				if _, err = con.Write([]byte("Than you for joining our chat!\n")); err != nil {
					log.Printf("failed to respond to client: %v\n", err)
				}
				log.Println("client requested server to close the connection so closing")
				return
			} else if strings.HasPrefix(clientMsg, "/name ") {
				newName := strings.SplitAfterN(clientMsg, "/name ", 2)[1]
				messages <- client.Name + " has changed name to: " + newName
				client.Name = newName
				clients[addr] = client

			} else {
				log.Printf("Received: %s, bytes: %d \n", string(clientMsg), len(clientMsg))
				messages <- fmt.Sprintf("(%s) %s", client.Name, string(clientMsg))
			}
		}
	}

}

func broadcastMsg() {

	for {
		msg := <-messages
		log.Println("received message from channel: ", msg)
		// ##### value is not con - its a client
		for ipaddr, client := range clients {
			log.Printf("broadcasting msg to ip: %v\n", ipaddr)
			if _, err := client.Conn.Write([]byte(ipaddr + " : " + msg + "\n")); err != nil {
				log.Printf("failed to send to client: %v\n", err)
			}

		}
	}

}

// ###### 2 handling user disconnection

func handleDisconn(addr string) {
	client := clients[addr]
	messages <- fmt.Sprintf("user '%s' has left the chat", client.Name)
	delete(clients, addr)

}
