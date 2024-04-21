package main

import (
  "bufio"
  "fmt"
  "net"
  "os"
)

func main() {
  ln, err := net.Listen("tcp", ":5001")
  if err != nil {
    fmt.Println("Error listening:", err.Error())
    os.Exit(1)
  }

  defer ln.Close()

  fmt.Println("TCP server listening on port 5001")

  for {
    conn, err := ln.Accept()
    if err != nil {
      fmt.Println("Error accepting: ", err.Error())
      os.Exit(1)
    }
    fmt.Println("Connected to client: ", conn.RemoteAddr().String())

    go handleRequest(conn)
  }
}

func handleRequest(conn net.Conn) {
  defer conn.Close()

  reader := bufio.NewReader(conn)
  writer := bufio.NewWriter(conn)

  for {
    message, err := reader.ReadString('\n')
    if err != nil {
      fmt.Println("Error reading: ", err.Error())
      break
    }
    fmt.Printf("Received: %s", message)


    writer.WriteString(message)
    writer.Flush()
    fmt.Printf("Echoed back: %s", message)
  }
}
