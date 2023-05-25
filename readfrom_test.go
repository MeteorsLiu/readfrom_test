package readfrom_test

import (
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"
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
			}
		}
	})
	// writer
	go newServer(l2, &wg, func(c net.Conn) {
		b := make([]byte, 1024)
		for {
			_, err := c.Write(b)
			if err != nil {
				log.Println("writer: ", err)
			}
		}
	})

	wg.Wait()
	t.Log("start to dial")
	c1, _ := net.Dial("tcp", "127.0.0.1:9998")
	c2, _ := net.Dial("tcp", "127.0.0.1:9999")

	go func() {
		_, err := io.Copy(c1, c2)
		log.Println(err)
	}()

	time.Sleep(10 * time.Second)
	log.Println("shutdown c2")
	c2.Close()
}
