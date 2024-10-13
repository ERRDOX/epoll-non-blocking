package main

import (
	"fmt"
	"syscall"
)

func createNonBlockingSocket() (int, error) {
	serverFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return 0, fmt.Errorf("error creating socket: %v", err)
	}

	err = syscall.SetsockoptInt(serverFd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		syscall.Close(serverFd)
		return 0, fmt.Errorf("error setting socket options: %v", err)
	}

	err = syscall.SetNonblock(serverFd, true)
	if err != nil {
		syscall.Close(serverFd)
		return 0, fmt.Errorf("error setting non-blocking mode: %v", err)
	}

	addr := &syscall.SockaddrInet4{Port: 8000}
	copy(addr.Addr[:], []byte{0, 0, 0, 0})
	err = syscall.Bind(serverFd, addr)
	if err != nil {
		syscall.Close(serverFd)
		return 0, fmt.Errorf("error binding socket: %v", err)
	}

	err = syscall.Listen(serverFd, syscall.SOMAXCONN)
	if err != nil {
		syscall.Close(serverFd)
		return 0, fmt.Errorf("error listening on socket: %v", err)
	}

	return serverFd, nil
}

func main() {
	serverFd, err := createNonBlockingSocket()
	if err != nil {
		fmt.Println("Error creating socket:", err)
		return
	}
	defer syscall.Close(serverFd)

	epollFd, err := syscall.EpollCreate1(0)
	if err != nil {
		fmt.Println("Error creating epoll:", err)
		return
	}
	defer syscall.Close(epollFd)

	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFd),
	}
	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, serverFd, &event); err != nil {
		fmt.Println("Error adding server socket to epoll:", err)
		return
	}

	events := make([]syscall.EpollEvent, 3)

	for {
		// epollWait: Waits for events on the epoll file descriptor.
		n, err := syscall.EpollWait(epollFd, events, -1)
		if err != nil {
			fmt.Println("Error waiting for epoll events:", err)
			return
		}
		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)
			if events[i].Events&(syscall.EPOLLERR|syscall.EPOLLRDHUP|syscall.EPOLLHUP) != 0 {
				fmt.Printf("Client disconnected or error on fd %d\n", fd)
				syscall.Close(fd)
				syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
				continue
			}
			// checks if the event is on the server's listening socket, indicating a new incoming connection.
			if fd == serverFd {
				// 1. get new connection and
				// 2.set it to non-blocking
				// 3.add it to epoll list with EPOLL_CTL_ADD
				fmt.Println("New connection on server socket")
				clientFd, _, err := syscall.Accept(serverFd)
				if err != nil {
					fmt.Println("Error accepting connection:", err)
					continue
				}
				err = syscall.SetNonblock(clientFd, true)
				if err != nil {
					fmt.Println("Error setting client socket to non-blocking mode:", err)
					syscall.Close(clientFd)
					continue
				}
				// Add clientFd to epoll
				event.Fd = int32(clientFd)
				if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, clientFd, &event); err != nil {
					fmt.Println("Error adding client socket to epoll:", err)
					syscall.Close(clientFd)
					continue
				}
			} else {
				// Handle client socket events (read/write)
				handleClientEvent(fd, epollFd)
			}
		}
	}
}

func handleClientEvent(fd, epollFd int) {
	buf := make([]byte, 1024)

	n, err := syscall.Read(fd, buf)
	if err != nil {
		fmt.Println("Error reading from socket:", err)
		syscall.Close(fd)
		syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
		return
	}

	fmt.Printf("Received %d bytes from client: %s\n", n, buf)
	responseString := fmt.Sprintf("Response from server to client with fd : %d !\n", fd)
	responsebyte := []byte(responseString)
	_, err = syscall.Write(fd, responsebyte)
	if err != nil {
		fmt.Println("Error writing to socket:", err)
		syscall.Close(fd)
		syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
		return
	}
}
