package telnetssh

import (
	"errors"
	"io"
	"strconv"

	"github.com/muxache/telnetssh/model"
	"golang.org/x/crypto/ssh"
)

type SshConnect struct {
	conn    *ssh.Client
	session *ssh.Session
	reader  io.Reader
	writer  io.WriteCloser
}

func Connect(c *model.Config) (*SshConnect, error) {

	var sshconfig ssh.Config
	sshconfig.SetDefaults()

	sshconfig.Ciphers = append(sshconfig.Ciphers, "aes256-cbc", "aes128-cbc", "3des-cbc", "des-cbc")
	sshconfig.KeyExchanges = append(sshconfig.KeyExchanges, "diffie-hellman-group14-sha1", "diffie-hellman-group-exchange-sha1", "diffie-hellman-group1-sha1")

	config := &ssh.ClientConfig{
		Config:          sshconfig,
		User:            c.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         c.Timeout,
		Auth: []ssh.AuthMethod{
			ssh.RetryableAuthMethod(ssh.Password(c.Password), 1),
		},
	}
	address := c.Host + ":" + strconv.Itoa(c.Port)
	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return &SshConnect{}, err
	}
	session, err := conn.NewSession()
	if err != nil {
		return &SshConnect{}, err
	}

	w, h := 255, 255
	err = session.RequestPty("xterm", h, w, ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	})
	if err != nil {
		return &SshConnect{}, err
	}
	reader, err := session.StdoutPipe()
	if err != nil {
		return &SshConnect{}, err
	}
	writer, err := session.StdinPipe()
	if err != nil {
		return &SshConnect{}, err
	}
	err = session.Shell()
	if err != nil {
		return &SshConnect{}, err
	}
	return &SshConnect{
		conn:    conn,
		session: session,
		reader:  reader,
		writer:  writer,
	}, nil
}

func (ssh *SshConnect) Read(b []byte) (n int, err error) {
	if ssh.conn == nil {
		return 0, errors.New("connection was closed")
	}

	return ssh.reader.Read(b)
}

func (ssh *SshConnect) Write(b []byte) (n int, err error) {
	if ssh.conn == nil {
		return 0, errors.New("connection was closed")
	}
	return ssh.writer.Write(b)
}

func (ssh *SshConnect) Close() {
	ssh.session.Close()
	ssh.conn.Close()

}
