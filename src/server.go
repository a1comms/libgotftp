package tftp

import (
	"fmt"
	"net"
)

type Server struct {
	conn       *net.UDPConn
	listenAddr *net.UDPAddr
	replyAddr  *net.UDPAddr
	portArray  []int
	portMap    map[int]bool
	portMapMutex sync.Mutex
	buffer     []byte
}

func (server *Server) Accept() (*RRQresponse, error) {

	written, addr, err := server.conn.ReadFrom(server.buffer)
	if err != nil {
		return nil, fmt.Errorf("Failed to read data from client: %v", err)
	}

	request, err := ParseRequest(server.buffer[:written])
	if err != nil {
		return nil, fmt.Errorf("Failed to parse request: %v", err)
	}
	request.Addr = &addr

	if request.Opcode != RRQ {
		return nil, fmt.Errorf("Unkown opcode %v", request.Opcode)
	}

	raddr, err := net.ResolveUDPAddr("udp", addr.String())
	if err != nil {
		return nil, fmt.Errorf("Failed to resolve client address: %v", err)
	}

	response, err := NewRRQresponse(server, raddr, request, false)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (server *Server) AllocatePort() (int, error) {
	server.portMapMutex.Lock()
	defer server.portMapMutex.Unlock()

	for _, port := range server.portArray {
		if _, ok := server.portMap[port]; ! ok {
			server.portMap[port] = true
			return port, nil
		}
	}

	return 0, fmt.Errorf("Unable to allocate reply port, non available from pool")
}

func (server *Server) FreePort(port int) (error) {
	server.portMapMutex.Lock()
	defer server.portMapMutex.Unlock()

	delete(server.portMap, port)
}

func NewTFTPServer(addr *net.UDPAddr, replyAddr *net.UDPAddr, portArray []int) (*Server, error) {
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("Failed listen UDP %v", err)
	}

	return &Server{
		conn,
		addr,
		replyAddr,
		portArray,
		make([]byte, 2048),
	}, nil

}
