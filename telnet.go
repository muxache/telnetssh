package telnetssh

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/muxache/telnetssh/model"
	"github.com/muxache/telnetssh/telnet"
)

type Telnet struct {
	host           string
	port           int
	username       string
	password       string
	conn           *telnet.TelnetConnection
	timeout        time.Duration
	timeoutCommand time.Duration
	isAuth         bool
	banerstring    string
}

type Commands struct {
	Command string
	Expect  string
}

/* NewTelnetConnectionAndAuth create a new connection and authentificated with login and password.
Default values(Port = 23, Timeout = 5 * time.Second)*/
func newTelnetConnectionAndAuth(c *model.Config) (Authentication, error) {
	t, err := newTelnetConnection(c.Host, c.Port, c.Timeout)
	if err != nil {
		return nil, err
	}

	t.username = c.Username
	t.password = c.Password
	t.timeout = c.Timeout
	t.timeoutCommand = c.TimeoutCommand

	if err = t.auth(); err != nil {
		return &Telnet{}, err
	}

	return t, nil
}

func (t *Telnet) EnterCommand(command, expect string) error {
	_, err := t.get(command, expect)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telnet) GetData(command string) ([]byte, error) {
	res, err := t.get(command, t.GetBanner())
	if err != nil {
		return res, err
	}
	return res, nil
}

func (t *Telnet) GetBanner() string {
	return t.banerstring
}

func (t *Telnet) Close() {
	t.conn.Close()
	t.isAuth = false
}

func (t *Telnet) get(command, expect string) ([]byte, error) {
	if t.conn == nil {
		return nil, errors.New("ssh onnection was closed")
	}
	var (
		resbuf      []byte
		commandSent bool
	)

	ch := make(chan error)

	var n int
	go func() {
		_, err := t.conn.Write([]byte(command + "\n"))
		if err != nil {
			ch <- err
			return
		}
		for {
			buf := make([]byte, 1024)
			n, err = t.conn.Read(buf)

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
	case <-time.After(t.timeoutCommand):
		return resbuf, errors.New("command timeout")
	}
}

func (t *Telnet) auth() error {
	if t.conn == nil {
		return errors.New("connection was closed")
	}

	ch := make(chan error)
	var resbuf []byte
	go func() {
		var enterLogin bool
		var enterPass bool
		for {
			buf := make([]byte, 512)
			n, err := t.conn.Read(buf)
			if err != nil {
				ch <- err
				return
			}
			resbuf = append(resbuf, buf[:n]...)
			if !enterLogin && regexp.MustCompile("sername|login").Match(buf) {
				t.conn.Write([]byte(t.username + "\n"))

			}
			if !enterPass && regexp.MustCompile("assword").Match(buf) {
				enterLogin = true
				t.conn.Write([]byte(t.password + "\n"))

			}
			if regexp.MustCompile("#|>").Match(buf) {
				enterPass = true
				t.isAuth = true
				t.banerstring = regexp.MustCompile("(?m)^(.*(>|#))").FindString(string(resbuf))
				ch <- nil
				return
			}
			if regexp.MustCompile("Error|fail|incorrect|Fail").Match(buf) {
				ch <- errors.New("authentication error")
				return
			}
		}
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(t.timeout):
		return errors.New("timeout")
	}
}

func newTelnetConnection(host string, port int, timeout time.Duration) (*Telnet, error) {
	address := host + ":" + strconv.Itoa(port)

	connection, err := telnet.Connect(address, timeout)
	if err != nil {
		fmt.Print(err)
		return &Telnet{}, err
	}

	t := &Telnet{
		conn:    connection,
		host:    host,
		port:    port,
		timeout: timeout,
	}
	return t, nil
}
