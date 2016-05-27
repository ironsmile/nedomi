package app

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/testutils"
)

// A helper function that returns a full config file that contains the supplied
// system config
func getCfg(sysConfig config.System) *config.Config {
	cfg := new(config.Config)
	cfg.System = sysConfig
	return cfg
}

func TestProperEnvironmentCreation(t *testing.T) {
	t.Parallel()
	tempDir, cleanup := testutils.GetTestFolder(t)
	defer cleanup()

	tempFile := filepath.Join(tempDir, "test_pid_file.pid")
	currentUser, err := user.Current()

	if err != nil {
		t.Fatal("Was not able to find the current user")
	}

	cfg := getCfg(config.System{
		User:    currentUser.Username,
		Workdir: tempDir,
		Pidfile: tempFile,
	})

	if err := SetupEnv(cfg); err != nil {
		t.Fatalf("Error on creating environment. %s", err)
	}

	wd, err := os.Getwd()

	if err != nil {
		t.Errorf("Error getting current directory. %s", err)
	}

	if tempDir != wd {
		t.Errorf("SetupEnv did not change the current directory. %s", err)
	}

	pidfh, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Was not able to open the created pid file. %s", err)
	}

	scanner := bufio.NewScanner(pidfh)
	if !scanner.Scan() {
		t.Fatal("Pidfile was empty.")
	}

	pidInFile, err := strconv.Atoi(strings.Trim(scanner.Text(), "\n"))
	if err != nil {
		t.Fatalf("Was not able to convert pid to int from the pidfile. %s", err)
	}

	progPid := os.Getpid()
	if pidInFile != progPid {
		t.Error("Pidfile in the file was different than the one expected")
	}

	testutils.ShouldntFail(t, os.Remove(tempFile))
}

func TestWhenPidFileCreationFails(t *testing.T) {
	t.Parallel()

	targetPidFile := filepath.FromSlash("/this/place/does/not/exists")

	cfg := getCfg(config.System{Pidfile: targetPidFile})
	err := SetupEnv(cfg)

	if err == nil {
		t.Errorf("There was no error with pidfile `%s`", targetPidFile)

		// Remove the file in the off chance it has been created
		// for some reason
		testutils.ShouldntFail(t, os.Remove(targetPidFile))
	}

	if pathErr, ok := err.(*os.PathError); !ok || pathErr.Op != "open" {
		t.Errorf("Error was for creating the file. Not for writing in it: `%s`", err)
	}
}

func TestWithFullFilesystem(t *testing.T) {
	t.Parallel()

	targetPidFile := "/dev/full"

	// We will run this test only on operating systems which has the
	// /dev/full device
	if !utils.FileExists(targetPidFile) {
		t.Skip("This OS does not support /dev/full")
	}

	cfg := getCfg(config.System{Pidfile: targetPidFile})
	err := SetupEnv(cfg)

	if err == nil {
		t.Error("There was no error with pidfile in full filesystem")
	}

	if pathErr, ok := err.(*os.PathError); !ok || pathErr.Op != "write" {
		t.Errorf("Error was for creating the file. Not for writing in it: `%s`", err)
	}

}

func TestWithFakeUser(t *testing.T) {
	t.Parallel()

	tempDir, cleanup := testutils.GetTestFolder(t)
	defer cleanup()

	targetPidFile := filepath.Join(tempDir, "pidfile")
	notExistingUser := "thisuserdoesnotexists"

	cfg := getCfg(config.System{
		User:    notExistingUser,
		Pidfile: targetPidFile,
	})
	err := SetupEnv(cfg)

	if err == nil {
		t.Errorf("There was no error when user `%s` was used", notExistingUser)
	}

	if _, ok := err.(user.UnknownUserError); !ok {
		t.Errorf("The fake user's error was not UknownUserError. It was `%s`", err)
	}
}

func TestChangingTheUserWihtNobody(t *testing.T) {
	t.Parallel()

	//!TODO: find out if this test is possible at all.
	// If not, delete it from here.
	t.Skip("Setting tye UID and GID is not supported for some reason")

	nobody, err := user.Lookup("nobody")

	if err != nil {
		if _, ok := err.(user.UnknownUserError); ok {
			t.Skip("This system does not have the nobody user." +
				" Skipping the test since it requires it")
		} else {
			t.Errorf("Error getting the nobody user: %s", err)
		}
	}

	tempDir, cleanup := testutils.GetTestFolder(t)
	defer cleanup()

	targetPidFile := filepath.Join(tempDir, "pidfile")

	cfg := getCfg(config.System{
		User:    nobody.Name,
		Pidfile: targetPidFile,
	})

	err = SetupEnv(cfg)
	if err != nil {
		t.Errorf("There was an error when setting gid and uit to %s's. %s",
			nobody.Name, err)
	}

	currentEuid := os.Geteuid()
	uidOfNobody, err := strconv.Atoi(nobody.Uid)

	if err != nil {
		t.Errorf("Error converting UID [%s] to int: %s", nobody.Uid, err)
	}

	if uidOfNobody != currentEuid {
		t.Errorf("The current user id was not set to nobody's. "+
			"Expected %d but it was %d",
			uidOfNobody, currentEuid)
	}

	currentEgid := os.Getegid()
	gidOfNobody, err := strconv.Atoi(nobody.Gid)

	if err != nil {
		t.Errorf("Error converting GID [%s] to int: %s", nobody.Gid, err)
	}

	if gidOfNobody != currentEgid {
		t.Errorf("The current group id was not set to nobody's. "+
			"Expected %d but it was %d", gidOfNobody, currentEgid)
	}
	testutils.ShouldntFail(t, os.Remove(targetPidFile))
}

func TestCleaningUpErrors(t *testing.T) {
	t.Parallel()

	targetPidFile := filepath.FromSlash("/this/place/does/not/exists")

	cfg := getCfg(config.System{Pidfile: targetPidFile})

	if err := CleanupEnv(cfg); err == nil {
		t.Errorf("There was not an error for missing pidfile")
	}

	wrongPidFile, err := ioutil.TempFile("", "wrong_pid_in_file_test")

	if err != nil {
		t.Fatalf("Failed to create a temporray file: %s", err)
	}

	defer func() {
		testutils.ShouldntFail(t, os.Remove(wrongPidFile.Name()))
	}()

	fmt.Fprintf(wrongPidFile, "%d", os.Getpid()+1)
	testutils.ShouldntFail(t, wrongPidFile.Close())

	cfg.System = config.System{
		Pidfile: wrongPidFile.Name(),
	}

	if err := CleanupEnv(cfg); err == nil {
		t.Error("There was not an error with pidfile with different process id")
	}

}

func TestCleaningUpSuccesful(t *testing.T) {
	t.Parallel()
	testPidFile, err := ioutil.TempFile("", "pidfile")

	if err != nil {
		t.Fatalf("Failed to create a temporray file: %s", err)
	}

	defer func() {
		if err := os.Remove(testPidFile.Name()); err == nil {
			t.Fatalf("Expected error but received none")
		}
	}()

	fmt.Fprintf(testPidFile, "%d", os.Getpid())
	testutils.ShouldntFail(t, testPidFile.Close())

	cfg := getCfg(config.System{Pidfile: testPidFile.Name()})

	if err := CleanupEnv(cfg); err != nil {
		t.Error("Error cleaning up the pidfile")
	}
}
