package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Server struct {
	listener 	     net.Listener
	quit 			 chan struct{}
	exited 			 chan struct{}
	db				 IMDB
	connections 	 map[int]net.Conn
	connCloseTimeout time.Duration
}

func NewServer() *Server{
	l, err := net.Listen("tcp", ":8080")
	if err  != nil {
		log.Fatal("Failed to create listener", err.Error())
	}

	srv := &Server {
		listener: 		  l,
		quit: 			  make(chan struct{}),
		exited: 		  make(chan struct{}),
		connections: 	  map[int]net.Conn{},
		db: 			  newDB(),
		connCloseTimeout: 10*time.Second,
	}

	go srv.serve()
	return srv
}

func (srv *Server) serve() {
	var id int;
	fmt.Println("Listening for clients")
	for {
		select {
		case <- srv.quit:
			fmt.Println("Shutting down the server")
			err := srv.listener.Close()
			if err != nil {
				fmt.Println("Could not close listener", err.Error())
			}
			if len(srv.connections) > 0 {
				srv.warnConnections(srv.connCloseTimeout)
				<- time.After(srv.connCloseTimeout)
				srv.closeConnections()
			}
			close(srv.exited)
			return
		default: 
			tcpListener := srv.listener.(*net.TCPListener)
			err := tcpListener.SetDeadline(time.Now().Add(2 * time.Second))
			if err != nil {
				fmt.Println("Failed to set listener deadline", err.Error())
			}

			conn, err := tcpListener.Accept()
			if oppErr, ok := err.(*net.OpError); ok && oppErr.Timeout(){
				continue
			}

			if err != nil {
				fmt.Println("Failed to accept connection", err.Error())
			}

			write(conn, "Welcome to IMDB server")
			srv.connections[id] = conn
			go func(connID int){
				fmt.Println("Client with id", connID, "joined")
				srv.handleConn(conn)
				delete(srv.connections, connID)
				fmt.Println("Client with id", connID, "left")
			}(id)
			id++
		}
	}
}

func write(conn net.Conn, s string) {
	_, err := fmt.Fprintf(conn, "%s\n*->", s)
	if err != nil {
		log.Fatal(err)
	}
}

func (srv *Server) handleConn(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		l := strings.ToLower(strings.TrimSpace(scanner.Text()))
		values := strings.Split(l, " ")
	
	switch{
	case len(values) == 3 && values[0] == "set": 
		srv.db.set(values[1], values[2])
		write(conn, "OK")
	case len(values) == 2 && values[0] == "get": 
	    key := values[1]
		value, found := srv.db.get(key)
		if !found {
			write(conn, fmt.Sprintf("Key %v was not found", key))
		} else {
			write(conn, value)
		}
	case len(values) == 2 && values[0] == "delete": 
		srv.db.delete(values[1])
		write(conn, "DELETED")
	case len(values) == 1 && values[0] == "exit": 
		if err := conn.Close(); err != nil {
			fmt.Println("Could not close connection", err.Error())
		}
	default: 
		write(conn, fmt.Sprintf("UNKNOWN: %s", l))
	}
	}
}

func (srv *Server) warnConnections (timeout time.Duration) {
	for _, conn := range srv.connections {
		write(conn, fmt.Sprintf("Host wants to shut down the server in: %s", timeout.String()))
	}
}

func (srv *Server) closeConnections () {
	fmt.Println("Closing all connections")
	for id, conn := range srv.connections {
		err := conn.Close()
		if err != nil {
			fmt.Println("Could not close connection with id: ", id)
		}

	}
}
func (srv *Server) Stop() {
	fmt.Println("Stopping DB server")
	close(srv.quit)
	<-srv.exited
	fmt.Println("Saving IMDB records")
	srv.db.save()
	fmt.Println("Database server was successfully stopped")
}