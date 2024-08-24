package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
)

func wsKeyFromRequest(request string) string {

	var wsKey = ""
	var lines = strings.Split(request, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		var header = strings.Split(line, ": ")

		if len(header) == 2 && header[0] == "Sec-WebSocket-Key" {
			wsKey = header[1]
			break
		}
	}
	return wsKey
}
func wsUpgradeResponse(b64hash string) string {
	var res = "HTTP/1.1 101 Switching Protocols\r\n"
	res += "Connection: Upgrade\r\n"
	res += "Upgrade: websocket\r\n"
	res += "Sec-WebSocket-Accept: " + b64hash + "\r\n"
	res += "\r\n"
	return res
}
func wsMessage(message []byte) []byte {

	var lenToByte = []byte{0, 0}
	binary.LittleEndian.PutUint16(lenToByte, uint16(cap(message)))
	var firstByte = []byte{0x81} // 8 final frame, 1 textual
	var length = lenToByte[0]

	fmt.Println(message)
	fmt.Println(length)

	var command = message
	var msg = append(firstByte, length)
	msg = append(msg, command...)
	return msg
}

func sendMessage(id int, conn net.Conn, msg []byte) {

	var sid = fmt.Sprintf("#%d:", id)
	fmt.Println(sid + "sending message")
	_, err := conn.Write(wsMessage(msg))
	if err != nil {
		log.Print("Write: ", err)
		return
	}
}
func handshake(conn net.Conn) {

	// read request
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Print("Handshake error: ", err)
	}
	var request = string(buffer)

	// handshake
	// ws key
	var wsKey = wsKeyFromRequest(request)

	if len(wsKey) == 0 {
		fmt.Println("invalid ws request, closing connection")
		_, err := conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\nError missing Sec-WebSocket-Accept header\r\n"))
		if err != nil {
			return
		}
		err = conn.Close()
		if err != nil {
			return
		}
		return
	}

	fmt.Println("handshake..")

	// ws magic
	var magic = []byte(`258EAFA5-E914-47DA-95CA-C5AB0DC85B11`)

	// calculate sha1
	h := sha1.New()
	h.Write([]byte(wsKey))
	h.Write(magic)
	var hash = h.Sum(nil)
	// base64
	var b64hash = base64.StdEncoding.EncodeToString(hash)

	// send response
	_, err = conn.Write([]byte(wsUpgradeResponse(b64hash)))
	if err != nil {
		return
	}

}

func writePipe(){


    file, err := os.OpenFile(pipeFileName, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		log.Fatal("Open named pipe file error:", err)
	}

    writer := bufio.NewWriter(file)

	fmt.Fprint(writer, "refresh\n")
	writer.Flush()

}
func readPipe() {

	os.Remove(pipeFileName) // remove if exists

	err := syscall.Mkfifo(pipeFileName, 0666)
	if err != nil {
		log.Fatal("Make pipe error:", err)
		return
	}
	// Reader
	file, err := os.OpenFile(pipeFileName, os.O_CREATE|os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		log.Fatal("Open named pipe file error:", err)
	}
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) == 0 {
			continue
		}
		if err == nil {
			fmt.Print("PIPE new string:" + string(line))
			for i, c := range clients {
				sendMessage(i, c, line)
			}

			// fmt.Print(".")
		} else {
			_ = line
			fmt.Printf("READ err: %v\n", err)
		}
	}
}

func processArgs() bool{
    if(len(os.Args)>1){
        if(os.Args[1] == "refresh"){
            writePipe()
            return true
        }
    }

    if(len(os.Args)>1){
        fmt.Println("invalid input")
        return true
    }
    return false
}

var addr = "127.0.0.1:8888"
var clients []net.Conn
var clients_status []bool
var pipeFileName string

func main() {

    // pipe name
    home, err := os.UserHomeDir()
    if err != nil {
    		log.Fatal(err)
    		return
    }
    pipeFileName = home+"/.plugin_server_message_pipe"


    // process cmd args
    cmd := processArgs()
    if cmd {
        return
    }

	// handle messages
	go readPipe()


	// server
	fmt.Println("Starting plugin websocket server (" + addr + ")... ")

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
		return
	}

	// handle clients
	for {
		conn, err := listen.Accept()
		clients = append(clients, conn)
		if err != nil {
			log.Print(err)
			return
		}
		fmt.Printf("New client")
		handshake(conn)
	}

}
