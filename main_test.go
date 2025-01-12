package main

import (
	"fmt"
	"syscall"
	"testing"
)

func TestHandleClientEvent(t *testing.T) {
	// Create a pair of connected sockets
	sockets, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("Error creating socket pair: %v", err)
	}
	defer syscall.Close(sockets[0])
	defer syscall.Close(sockets[1])

	// Write data to the first socket to simulate client input
	clientMessage := "Client Connected [123456789012345678901234567890]"
	_, err = syscall.Write(sockets[0], []byte(clientMessage))
	if err != nil {
		t.Fatalf("Error writing to socket: %v", err)
	}

	handleClientEvent(sockets[1], 0)

	// Read the response from the second socket
	buf := make([]byte, 1024)
	n, err := syscall.Read(sockets[0], buf)
	if err != nil {
		t.Fatalf("Error reading from socket: %v", err)
	}

	// Verify the response
	expectedResponse := fmt.Sprintf("Response from server to client with fd : %d !\n", sockets[1])
	if string(buf[:n]) != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, string(buf[:n]))
	}
}

func BenchmarkHandleClientEvent(b *testing.B) {
	// Create a pair of connected sockets
	// socket pair use for test, it create 2 socket, one for read, one for write
	sockets, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		b.Fatalf("Error creating socket pair: %v", err)
	}
	defer syscall.Close(sockets[0])
	defer syscall.Close(sockets[1])

	clientMessage := "Client Connected [123456789012345678901234567890]"
	for i := 0; i < b.N; i++ {
		_, err = syscall.Write(sockets[0], []byte(clientMessage))
		if err != nil {
			b.Fatalf("Error writing to socket: %v", err)
		}

		handleClientEvent(sockets[1], 0)

		buf := make([]byte, 1024)
		_, err = syscall.Read(sockets[0], buf)
		if err != nil {
			b.Fatalf("Error reading from socket: %v", err)
		}
	}
}
