package browser

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

const SOCK = "browserker.sock"

type SocketLeaser struct {
	leaserClient http.Client
}

func NewSocketLeaser() *SocketLeaser {
	s := &SocketLeaser{}
	s.leaserClient = http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", SOCK)
			},
		},
	}
	return s
}

func (s *SocketLeaser) Cleanup() (string, error) {
	resp, err := s.leaserClient.Get("http://unix/cleanup")
	if err != nil {
		return "", err
	}

	response, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	if resp.StatusCode == 500 {
		return "", errors.New(string(response))
	}

	return string(response), nil
}

func (s *SocketLeaser) Acquire() (string, error) {
	resp, err := s.leaserClient.Get("http://unix/acquire")
	if err != nil {
		return "", err
	}

	port, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	return string(port), nil
}

func (s *SocketLeaser) Count() (string, error) {
	resp, err := s.leaserClient.Get("http://unix/count")
	if err != nil {
		return "", err
	}

	count, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	return string(count), nil
}

func (s *SocketLeaser) Return(port string) error {
	resp, err := s.leaserClient.Get("http://unix/return?port=" + port)
	if err != nil {
		return err
	}

	_, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode == 404 {
		return errors.New("browser not found")
	}
	return nil
}
