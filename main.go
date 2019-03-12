package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func startTest(name string) {
	fmt.Println("starting test ", name)
	o := exec.Command("./tt/tt.test", "-test.run", "Test")
	oo, _ := o.StdoutPipe()
	o.Start()
	defer o.Wait()
	go io.Copy(os.Stdout, oo)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Write "Hello, world!" to the response body
	//fmt.Fprintf(w, "Hello, world!: %#v", r.TLS.PeerCertificates[0].DNSNames)
	fmt.Printf("%v", r.TLS.PeerCertificates[0].Subject)
	fmt.Fprintf(w, "Hello, world! %v", r.TLS.PeerCertificates[0].Subject)
}

func startServer() {
	// Set up a /hello resource handler
	http.HandleFunc("/hello", helloHandler)
	log.Println("starting server")

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile("certs/ca/certs/cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	_ = caCert
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair("certs/server/cert.pem", "certs/server/key.pem")
	if err != nil {
		log.Println(err)
		return
	}

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
	}
	tlsConfig.BuildNameToCertificate()
	ll := log.New(os.Stdout, "", 0)

	// Create a Server instance to listen on port 8443 with the TLS config
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		ErrorLog:  ll,
	}
	aa := make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	server.TLSNextProto = aa

	// Listen to HTTPS connections with the server certificate and wait
	log.Fatal(server.ListenAndServeTLS("certs/server/cert.pem", "certs/server/key.pem"))

}

func execClient() {
	// Read the key pair to create certificate
	cert, err := tls.LoadX509KeyPair("certs/client/client.crt", "certs/client/key.pem")
	if err != nil {
		log.Fatal(err)
	}

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile("certs/server/client.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create a HTTPS client and supply the created CA pool and certificate
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	// Request /hello via the created HTTPS client over port 8443 via GET
	r, err := client.Get("https://localhost:8443/hello")
	if err != nil {
		log.Fatal(err)
	}

	// Read the response body
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response body to stdout
	fmt.Printf("%s\n", body)
}

func main() {
	isServer := os.Getenv("SERVER") != ""
	if isServer {
		startServer()
		return
	}
	execClient()
}
