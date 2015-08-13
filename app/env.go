package app

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/utils"
)

// SetupEnv will create pidfile and possibly change the workdir.
func SetupEnv(cfg *config.Config) error {

	if cfg.System.User != "" {
		// Get the current user
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("Could not get the current user: %s", err)
		}

		// If the current user is different than the wanted user, try to change it
		if currentUser.Username != cfg.System.User {

			wantedUser, err := user.Lookup(cfg.System.User)
			if err != nil {
				return err
			}

			uid, err := strconv.Atoi(wantedUser.Uid)
			if err != nil {
				return fmt.Errorf("Error converting UID [%s] to int: %s", wantedUser.Uid, err)
			}

			gid, err := strconv.Atoi(wantedUser.Gid)
			if err != nil {
				return fmt.Errorf("Error converting GID [%s] to int: %s", wantedUser.Gid, err)
			}

			if err = syscall.Setgid(gid); err != nil {
				return fmt.Errorf("Setting group id: %s", err)
			}

			if err = syscall.Setuid(uid); err != nil {
				return fmt.Errorf("Setting user id: %s", err)
			}
		}
	}

	if cfg.System.Workdir != "" {
		if err := os.Chdir(cfg.System.Workdir); err != nil {
			return fmt.Errorf("Could not chdir to '%s': %s", cfg.System.Workdir, err)
		}
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

// CleanupEnv has to be called on application shutdown. Will remove the pidfile.
//!TODO: see to it that fh.Close() is called properly
func CleanupEnv(cfg *config.Config) error {
	if !utils.FileExists(cfg.System.Pidfile) {
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
