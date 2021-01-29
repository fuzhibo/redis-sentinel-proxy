package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

var (
	masterAddrChan chan *net.TCPAddr = make(chan *net.TCPAddr, 10)
	raddr          *net.TCPAddr
	saddrs         []*net.TCPAddr

	localAddr    = flag.String("listen", ":9999", "local address")
	sentinelAddr = flag.String("sentinel", ":26379", "remote address, split with ','")
	masterName   = flag.String("master", "", "name of the master redis node")
)

func main() {
	flag.Parse()

	laddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		log.Fatal("Failed to resolve local address: ", err)
	}
	// check if "," in the sentinelAddr string
	if strings.Contains(*sentinelAddr, ",") {
		// split string with ","
		sentinelAddrs := strings.Split(*sentinelAddr, ",")
		for _, addr := range sentinelAddrs {
			saddr, err := net.ResolveTCPAddr("tcp", addr)
			if err != nil {
				log.Fatal("Failed to resolve sentinel address ", addr, " : ", err)
				continue
			}
			saddrs = append(saddrs, saddr)
		}
	} else {
		// take it as a single addr
		saddr, err := net.ResolveTCPAddr("tcp", *sentinelAddr)
		if err != nil {
			log.Fatal("Failed to resolve sentinel address ", *sentinelAddr, " : ", err)
		} else {
			saddrs = append(saddrs, saddr)
		}
	}
	if len(saddrs) == 0 {
		log.Fatal("Failed to get sentinel address.")
	}

	go master()

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		go proxy(conn)
	}
}

func master() {
	var inx = 0
	for {
		masterAddr, err := getMasterAddr(saddrs[inx%len(saddrs)], *masterName)
		if err != nil {
			log.Println("Failed to get master addr from ", saddrs[inx%len(saddrs)], " due to ", err)
			// try next addr
			inx += 1
			continue
		}
		masterAddrChan <- masterAddr
		time.Sleep(1 * time.Second)
		// balance
		inx += 1
	}
}

func pipe(r io.Reader, w io.WriteCloser) {
	io.Copy(w, r)
	w.Close()
}

func proxy(local io.ReadWriteCloser) {
	remoteAddr := <-masterAddrChan
	remote, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		log.Println("Failed to connect ", remoteAddr, " due to ", err)
		local.Close()
		return
	}
	go pipe(local, remote)
	go pipe(remote, local)
}

func getMasterAddr(sentinelAddress *net.TCPAddr, masterName string) (*net.TCPAddr, error) {
	conn, err := net.DialTCP("tcp", nil, sentinelAddress)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	conn.Write([]byte(fmt.Sprintf("sentinel get-master-addr-by-name %s\n", masterName)))

	b := make([]byte, 256)
	_, err = conn.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	parts := strings.Split(string(b), "\r\n")

	if len(parts) < 5 {
		err = errors.New("Couldn't get master address from sentinel")
		return nil, err
	}

	//getting the string address for the master node
	stringaddr := fmt.Sprintf("%s:%s", parts[2], parts[4])
	addr, err := net.ResolveTCPAddr("tcp", stringaddr)

	if err != nil {
		return nil, err
	}

	//check that there's actually someone listening on that address
	conn2, err := net.DialTCP("tcp", nil, addr)
	if err == nil {
		defer conn2.Close()
	}

	return addr, err
}
