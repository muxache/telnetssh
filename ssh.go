package telnetssh

import (
	"errors"
	"regexp"
	"time"

	"github.com/muxache/telnetssh/model"
	"github.com/muxache/telnetssh/ssh"
)

type SSH struct {
	host           string
	port           int
	username       string
	password       string
	conn           *ssh.SshConnect
	timeout        time.Duration
	timeoutCommand time.Duration
	isAuth         bool
	banerstring    string
}

/* NewSSHConnectionAndAuth create a new connection and authentificated with login and password.
Default values(Port = 23, Timeout = 5 * time.Second)*/
func newSSHConnectionAndAuth(c *model.Config) (Authentication, error) {
	s, err := ssh.Connect(c)
	if err != nil {
		return nil, err
	}

	ssh := &SSH{
		conn:           s,
		host:           c.Host,
		port:           c.Port,
		username:       c.Username,
		password:       c.Password,
		timeout:        c.Timeout,
		timeoutCommand: c.TimeoutCommand,
	}
	if err := ssh.setBannerName(); err != nil {
		return nil, err
	}

	return ssh, nil
}

func (s *SSH) EnterCommand(command, expect string) error {
	_, err := s.get(command, expect)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSH) GetData(command string) ([]byte, error) {
	res, err := s.get(command, s.GetBanner())
	if err != nil {
		return res, err
	}
	return res, nil
}

func (s *SSH) get(command, expect string) ([]byte, error) {
	if s.conn == nil {
		return nil, errors.New("ssh connection was closed")
	}
	var (
		resbuf      []byte
		commandSent bool
	)

	ch := make(chan error)

	var n int
	go func() {
		_, err := s.conn.Write([]byte(command + "\n"))
		if err != nil {
			ch <- err
			return
		}
		for {
			buf := make([]byte, 1024)
			n, err = s.conn.Read(buf)

			if err != nil {
				ch <- err
				return
			}
			resbuf = append(resbuf, buf[:n]...)
			if regexp.MustCompile(command).Match(resbuf) {
				commandSent = true
			}
			if regexp.MustCompile(expect).Match(buf[:n]) {
				if commandSent {
					ch <- nil
					return
				}
			}

		}

	}()
	select {
	case err := <-ch:
		if err != nil {
			switch err.Error() {
			case "EOF":
				return resbuf, nil
			default:
				return resbuf, err
			}
		}
		return resbuf, err
	case <-time.After(s.timeoutCommand):
		return resbuf, errors.New(command + " command timeout")
	}
}

func (s *SSH) GetBanner() string {
	return s.banerstring
}

func (s *SSH) Close() {
	s.conn.Close()
	s.isAuth = false
}

func (s *SSH) setBannerName() error {
	if s.conn == nil {
		return errors.New("ssh connection was closed")
	}
	ch := make(chan error)
	var resbuf []byte
	go func() {
		for {
			buf := make([]byte, 512)
			n, err := s.conn.Read(buf)
			if err != nil {
				ch <- err
				return
			}
			resbuf = append(resbuf, buf[:n]...)
			if regexp.MustCompile("(?m)^(.*(>|#))").Match(buf) {
				s.isAuth = true
				s.banerstring = regexp.MustCompile("(?m)^(.*(>|#))").FindString(string(resbuf))
				ch <- nil
				return
			}
		}
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(s.timeout):
		return errors.New("set banner timeout")
	}
}
