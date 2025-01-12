# Non-Blocking Socket Example

## Problem
Bring up a socket in non-blocking mode and listen to its events using a notification system like epoll or io_uring, and perform read and write operations. If we increase the number of sockets, will it become more complex? More complex but not efficient.

## Workflow
1. **Create a Non-Blocking Socket**: Use `syscall.SetNonblock()` to create a non-blocking socket. This allows the socket to return immediately from read or write operations.
2. **Add Socket to Epoll Instance**: Use `syscall.EpollCtl()` to add the socket to the epoll instance. Epoll will monitor this socket and notify your program when the socket has data to read or is ready to send data.
3. **Wait for Events**: Use `syscall.EpollWait()` to wait for events. When it returns, it tells your program which sockets are ready. You can then read or write data without blocking, because the socket is in non-blocking mode.

## Epoll (syscall.Epoll)
**What it is**: Epoll is a system call (primarily available on Linux) that allows programs to monitor multiple file descriptors (such as sockets) and be notified when they are ready for reading, writing, or when errors occur. Unlike traditional methods like `poll()` or `select()`, which require scanning all file descriptors to find ready ones, epoll is more efficient for large numbers of file descriptors.

Epoll waits for events on a set of file descriptors (like incoming data on a socket or a new connection on a listening socket) and only notifies the program when an event occurs. This means you don't have to continuously check every socket to see if it's ready for reading or writing—epoll does that efficiently for you.

You create an epoll instance with `syscall.EpollCreate1()`, and then you can add file descriptors (like sockets) to the epoll instance using `syscall.EpollCtl()`. The program then waits for events using `syscall.EpollWait()`.

Epoll provides an efficient way to manage and monitor multiple file descriptors, making it perfect for handling large numbers of network connections without requiring a separate thread or process for each connection.

## Non-Blocking I/O
Non-blocking I/O ensures that I/O operations (like reading or writing data) do not block the execution of the program. If a socket isn't ready for reading or writing, the program can immediately continue working on other tasks.

## Test
1. Run the program:
  ```sh
  go run .
  ```
2. Connect using telnet:
  ```sh
  telnet localhost 8000
  ```

## Next Steps
In the next step, I will try to implement the same functionality using io_uring instead of epoll.
