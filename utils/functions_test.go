package utils

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
)

func TestFileExistsFunction(t *testing.T) {
	tmpDir := os.TempDir()

	defer os.Remove(tmpDir)

	if exists := FileExists(tmpDir); exists {
		t.Errorf("Expected false when calling FileExists with directory: %s", tmpDir)
	}

	tmpFile, err := ioutil.TempFile(tmpDir, "functest")

	if err != nil {
		t.Fatalf("Creating a temporary file filed: %s", err)
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if exists := FileExists(tmpFile.Name()); !exists {
		t.Errorf("Expected true when calling FileEists with a file %s", tmpFile.Name())
	}
}

func TestProperEnvironmentCreation(t *testing.T) {
	tempDir := os.TempDir()
	defer os.Remove(tempDir)

	tempFile := filepath.Join(tempDir, "test_pid_file.pid")
	currentUser, err := user.Current()

	if err != nil {
		t.Fatal("Was not able to find the current user")
	}

	cfg := &config.Config{
		System: config.SystemSection{
			User:    currentUser.Name,
			Workdir: tempDir,
			Pidfile: tempFile,
		},
	}

	if err := SetupEnv(cfg); err != nil {
		t.Fatalf("Error on creating environment. %s", err)
	}

	defer os.Remove(tempFile)

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
}

func TestWhenPidFileCreationFails(t *testing.T) {

	targetPidFile := filepath.FromSlash("/this/place/does/not/exists")

	cfg := &config.Config{
		System: config.SystemSection{
			Pidfile: targetPidFile,
		},
	}

	err := SetupEnv(cfg)

	if err == nil {
		t.Errorf("There was no error with pidfile `%s`", targetPidFile)

		// Remove the file in the off chance it has been created
		// for some reason
		os.Remove(targetPidFile)
	}

	if pathErr, ok := err.(*os.PathError); !ok || pathErr.Op != "open" {
		t.Errorf("Error was for creating the file. Not for writing in it: `%s`", err)
	}
}

func TestWithFullFilesystem(t *testing.T) {

	targetPidFile := "/dev/full"

	// We will run this test only on operating sytems which has the
	// /dev/full device
	if !FileExists(targetPidFile) {
		t.Skip("This OS does not support /dev/full")
	}

	cfg := &config.Config{
		System: config.SystemSection{
			Pidfile: targetPidFile,
		},
	}

	err := SetupEnv(cfg)

	if err == nil {
		t.Error("There was no error with pidfile in full filesystem")
	}

	if pathErr, ok := err.(*os.PathError); !ok || pathErr.Op != "write" {
		t.Errorf("Error was for creating the file. Not for writing in it: `%s`", err)
	}

}

func TestWithFakeUser(t *testing.T) {

	tempDir := os.TempDir()
	defer os.Remove(tempDir)

	targetPidFile := filepath.Join(tempDir, "pidfile")
	notExistingUser := "thisuserdoesnotexists"

	cfg := &config.Config{
		System: config.SystemSection{
			User:    notExistingUser,
			Pidfile: targetPidFile,
		},
	}

	err := SetupEnv(cfg)

	defer os.Remove(targetPidFile)

	if err == nil {
		t.Errorf("There was no error when user `%s` was used", notExistingUser)
	}

	if _, ok := err.(user.UnknownUserError); !ok {
		t.Errorf("The fake user's error was not UknownUserError. It was `%s`", err)
	}
}

func TestChangingTheUserWihtNobody(t *testing.T) {

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

	tempDir := os.TempDir()
	defer os.Remove(tempDir)

	targetPidFile := filepath.Join(tempDir, "pidfile")

	cfg := &config.Config{
		System: config.SystemSection{
			User:    nobody.Name,
			Pidfile: targetPidFile,
		},
	}

	err = SetupEnv(cfg)

	defer os.Remove(targetPidFile)

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

}

func TestCleaningUpErrors(t *testing.T) {

	targetPidFile := filepath.FromSlash("/this/place/does/not/exists")

	cfg := &config.Config{
		System: config.SystemSection{
			Pidfile: targetPidFile,
		},
	}

	if err := CleanupEnv(cfg); err == nil {
		t.Errorf("There was not an error for missing pidfile")
	}

	wrongPidFile, err := ioutil.TempFile("", "wrong_pid_in_file_test")

	if err != nil {
		t.Fatalf("Failed to create a temporray file: %s", err)
	}

	defer os.Remove(wrongPidFile.Name())

	fmt.Fprintf(wrongPidFile, "%d", os.Getpid()+1)
	wrongPidFile.Close()

	cfg.System = config.SystemSection{
		Pidfile: wrongPidFile.Name(),
	}

	if err := CleanupEnv(cfg); err == nil {
		t.Error("There was not an error with pidfile with different process id")
	}

}

func TestCleaningUpSuccesful(t *testing.T) {
	testPidFile, err := ioutil.TempFile("", "pidfile")

	if err != nil {
		t.Fatalf("Failed to create a temporray file: %s", err)
	}

	defer os.Remove(testPidFile.Name())

	fmt.Fprintf(testPidFile, "%d", os.Getpid())
	testPidFile.Close()

	cfg := &config.Config{
		System: config.SystemSection{
			Pidfile: testPidFile.Name(),
		},
	}

	if err := CleanupEnv(cfg); err != nil {
		t.Error("Error cleaning up the pidfile")
	}
}
