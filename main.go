package main
//Main file for the Load Balancer project
import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type SimpleServer struct {
	address string
	proxy   *httputil.ReverseProxy
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewSimpleServer(address string) *SimpleServer {
	serverUrl, err := url.Parse(address)
	if err != nil {
		log.Fatal(err)
	}
	return &SimpleServer{
		address: address,
		proxy:   httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func (s *SimpleServer) Address() string {
	return s.address
}

func (s *SimpleServer) IsAlive() bool {
	return true
}

func (s *SimpleServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func (lb *LoadBalancer) GetNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount]
	for !server.IsAlive() {
		lb.roundRobinCount = (lb.roundRobinCount + 1) % len(lb.servers)
		server = lb.servers[lb.roundRobinCount]
	}
	lb.roundRobinCount = (lb.roundRobinCount + 1) % len(lb.servers)
	return server
}

func (lb *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	nextServer := lb.GetNextAvailableServer()
	fmt.Printf("Forwarding request to the given address %q\n", nextServer.Address())
	nextServer.Serve(w, r)
}

func main() {
	servers := []Server{
		NewSimpleServer("https://www.google.com"),
		NewSimpleServer("https://www.microsoft.com"),
		NewSimpleServer("https://www.amazon.com"),
	}
	lb := NewLoadBalancer("8000", servers)
	HandleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.ServeProxy(w, r)
	}
	http.HandleFunc("/", HandleRedirect)
	fmt.Printf("Listening on port number %s\n", lb.port)
	log.Fatal(http.ListenAndServe(":"+lb.port, nil))
}
