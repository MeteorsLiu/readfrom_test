package readfrom_test

import (
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func newServer(l net.Listener, wg *sync.WaitGroup, f func(net.Conn)) {
	wg.Done()
	c, _ := l.Accept()
	f(c)
}

func TestReadFrom(t *testing.T) {
	var wg sync.WaitGroup
	l1, _ := net.Listen("tcp", "127.0.0.1:9998")
	l2, _ := net.Listen("tcp", "127.0.0.1:9999")
	wg.Add(2)
	// reader
	go newServer(l1, &wg, func(c net.Conn) {
		b := make([]byte, 1024)
		for {
			_, err := c.Read(b)
			if err != nil {
				log.Println("Reader: ", err)
				return
			}
		}
	})
	// writer
	go newServer(l2, &wg, func(c net.Conn) {
		b := make([]byte, 1024)
		for {
			time.Sleep(6 * time.Second)
			_, err := c.Write(b)
			if err != nil {
				log.Println("writer: ", err)
				return
			}
		}
	})

	wg.Wait()
	t.Log("start to dial")
	c1, _ := net.Dial("tcp", "127.0.0.1:9998")
	c2, _ := net.Dial("tcp", "127.0.0.1:9999")
	rc, _ := c2.(*net.TCPConn).SyscallConn()
	rc.Control(func(fd uintptr) {
		unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_USER_TIMEOUT, 3*1000)
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(c2, c1)
		log.Println(err)
	}()

	time.Sleep(10 * time.Second)
	log.Println("shutdown c2")
	c1.SetReadDeadline(time.Now())
	c2.SetReadDeadline(time.Now())
	wg.Wait()
}
