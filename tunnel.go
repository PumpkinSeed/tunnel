package tunnel

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

// Endpoint ...
type Endpoint struct {
	Host string
	Port int
}

// SSHTunnel ...
type SSHTunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint

	Config *ssh.ClientConfig
}

// Tunnel -> Config container
type Tunnel struct {
	LocalHost         string
	LocalPort         int
	ServerHost        string
	ServerPort        int
	RemoteHost        string
	RemotePort        int
	Username          string
	Password          string
	UseRSAKeysForAuth bool
	Path              string
}

// New connection
func New(lh, sh, rh string, lp, sp, rp int) Tunnel {
	return Tunnel{
		LocalHost:  lh,
		LocalPort:  lp,
		ServerHost: sh,
		ServerPort: sp,
		RemoteHost: rh,
		RemotePort: rp,
		Username:   "",
	}
}

// AuthWithRSAKey -> Using RSA key for SSH auth
func (t *Tunnel) AuthWithRSAKey(un, path string) {
	t.UseRSAKeysForAuth = true
	t.Username = un
	t.Path = path
}

// AuthWithPassword -> Using password for SSH auth
func (t *Tunnel) AuthWithPassword(un, password string) {
	t.UseRSAKeysForAuth = false
	t.Username = un
	t.Password = password
}

// Setup connection
func (t Tunnel) Setup() error {
	var sshConfig *ssh.ClientConfig

	if t.Username == "" {
		return errors.New("Authentication not found")
	}

	if t.UseRSAKeysForAuth {
		key, err := getKeyFile(t.Path)
		if err != nil {
			return err
		}
		sshConfig = &ssh.ClientConfig{
			User: t.Username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(key),
			},
		}
	} else {
		sshConfig = &ssh.ClientConfig{
			User: t.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(t.Password),
			},
		}
	}

	tunnelConf := &SSHTunnel{
		Config: sshConfig,
		Local: &Endpoint{
			Host: t.LocalHost,
			Port: t.LocalPort,
		},
		Server: &Endpoint{
			Host: t.ServerHost,
			Port: t.ServerPort,
		},
		Remote: &Endpoint{
			Host: t.RemoteHost,
			Port: t.RemotePort,
		},
	}

	tunnelConf.start()

	return nil
}

func getKeyFile(file string) (key ssh.Signer, err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return
	}
	return
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

func (tunnel *SSHTunnel) start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			return err
		}
		go tunnel.forward(conn)
	}
}

func (tunnel *SSHTunnel) forward(localConn net.Conn) {
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		fmt.Printf("Server dial error: %s\n", err)
		return
	}

	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		fmt.Printf("Remote dial error: %s\n", err)
		return
	}

	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		if err != nil {
			fmt.Printf("io.Copy error: %s", err)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}
