package telnetssh

import (
	"errors"
	"net"
	"time"

	"github.com/muxache/telnetssh/model"
)

type Config struct {
	Host           string
	Port           int
	Username       string
	Password       string
	Timeout        time.Duration
	TimeoutCommand time.Duration
}

func NewConnection(conf *Config) (Authentication, error) {
	c := &model.Config{
		Host:           conf.Host,
		Username:       conf.Username,
		Password:       conf.Password,
		Port:           conf.Port,
		Timeout:        conf.Timeout,
		TimeoutCommand: conf.TimeoutCommand,
	}
	if len(c.Username) == 0 {
		return nil, errors.New("username is not defined")
	}
	if len(c.Password) == 0 {
		return nil, errors.New("password is not defined")
	}
	if len(c.Host) == 0 {
		return nil, errors.New("host is not defined")
	}

	if c.Timeout == 0 {
		c.Timeout = time.Second * 5
	}
	if c.Port == 0 {
		port, err := findPort(c.Host)
		if err != nil {
			return nil, err
		}
		c.Port = port
	}

	if c.Port == 23 {
		conn, err := newTelnetConnectionAndAuth(c)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	if c.Port == 22 {
		conn, err := newSSHConnectionAndAuth(c)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	return nil, errors.New("error")
}

func findPort(host string) (int, error) {
	d := net.Dialer{
		Timeout: time.Millisecond * 200,
	}
	ssh, err := d.Dial("tcp", host+":22")
	if err != nil {
		err = nil
	} else {
		ssh.Close()
		return 22, nil
	}
	telnet, err := d.Dial("tcp", host+":23")
	if err != nil {
		err = nil
	} else {
		telnet.Close()
		return 23, nil
	}
	return 0, errors.New("no avalible ports to connection")
}
