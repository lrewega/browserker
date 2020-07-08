package browser

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// LeaserService for a browser
type LeaserService interface {
	Acquire() (string, error) // returns port number
	SetHeadless()
	Return(port string) error
	Cleanup() (string, error)
	Count() (string, error)
}

func randPort() string {
	l, err := net.Listen("tcp", ":0")

	if err != nil {
		log.Warn().Err(err).Msg("unable to get port using default 9022")
		return "9022"
	}
	_, randPort, _ := net.SplitHostPort(l.Addr().String())
	l.Close()
	return randPort
}

func randProfile(tmp string) string {
	os.MkdirAll(tmp, 0766)
	profile, err := ioutil.TempDir(tmp, "gcd")
	if err != nil {
		log.Fatal().Err(err).Msg("profile returned empty which could delete system files on termination")
	}

	if profile == "" {
		log.Fatal().Msg("profile returned empty which could delete system files on termination")
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to create temporary profile directory")
		return "tmp"
	}

	return profile
}

// RemoveTmpContents that the browser created
func RemoveTmpContents(tmp string) error {
	files, err := filepath.Glob(filepath.Join(tmp, "gcd*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// KillOldProcesses with a vengence
func KillOldProcesses() error {
	killer := FindKill("google-chrome")
	cmd := exec.Command(killer[0], killer[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warn().Msgf("google-chrome %s:%s", err.Error(), string(output))
	}

	killer = FindKill("chrome")
	cmd = exec.Command(killer[0], killer[1:]...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Warn().Msgf("chrome %s:%s", err.Error(), string(output))
	}
	return nil
}
