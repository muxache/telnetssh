package telnetssh

type Authentication interface {
	EnterCommand(command, expect string) error
	GetData(command string) ([]byte, error)
	GetBanner() string
	Close()
}
