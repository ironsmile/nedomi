/*
   Package utils exports few handy functions
*/
package utils

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/ironsmile/nedomi/config"
)

// SetupEnv will create pidfile and possibly change the workdir.
//!TODO: add SetupEnv and CleanupEnv into Application?
func SetupEnv(cfg *config.Config) error {

	if cfg.System.User != "" {
		user, err := user.Lookup(cfg.System.User)
		if err != nil {
			return err
		}

		uid, err := strconv.Atoi(user.Uid)

		if err != nil {
			return fmt.Errorf("Error converting UID [%s] to int: %s", user.Uid, err)
		}

		gid, err := strconv.Atoi(user.Gid)

		if err != nil {
			return fmt.Errorf("Error converting GID [%s] to int: %s", user.Gid, err)
		}

		if err = syscall.Setgid(gid); err != nil {
			return fmt.Errorf("Setting group id: %s", err)
		}

		if err = syscall.Setuid(uid); err != nil {
			return fmt.Errorf("Setting user id: %s", err)
		}
	}

	if cfg.System.Workdir != "" {
		os.Chdir(cfg.System.Workdir)
	}

	pFile, err := os.Create(cfg.System.Pidfile)

	if err != nil {
		return err
	}
	defer pFile.Close()

	_, err = pFile.WriteString(fmt.Sprintf("%d", os.Getpid()))

	if err != nil {
		return err
	}

	return nil
}

// Function to be called on application shutdown. Will remove the pidfile.
//!TODO: see to it that fh.Close() is called properly
func CleanupEnv(cfg *config.Config) error {
	if !FileExists(cfg.System.Pidfile) {
		return fmt.Errorf("Pidfile %s does not exists.", cfg.System.Pidfile)
	}
	fh, err := os.Open(cfg.System.Pidfile)
	if err != nil {
		return err
	}
	var pid int
	_, err = fmt.Fscanf(fh, "%d", &pid)
	if err != nil {
		return err
	}
	if pid != os.Getpid() {
		return fmt.Errorf("File had different pid: %d", pid)
	}
	fh.Close()
	return os.Remove(cfg.System.Pidfile)
}

// FileExists returns true if filePath is already existing regular file. It is a
// directory FileExists will return false.
func FileExists(filePath string) bool {
	st, err := os.Stat(filePath)
	return err == nil && !st.IsDir()
}
