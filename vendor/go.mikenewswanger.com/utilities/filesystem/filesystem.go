package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()
var verbosity = uint8(0)

// SetLogger sets up a logrus instance
func SetLogger(l *logrus.Logger) {
	logger = l
}

// SetVerbosity sets the verbosity for the filesystem package
func SetVerbosity(v uint8) {
	verbosity = v
}

// BuildAbsolutePathFromHome builds an absolute path (i.e. /home/user/example) from a home-based path (~/example)
func BuildAbsolutePathFromHome(path string) (string, error) {
	var err error
	var fields = logrus.Fields{
		"path":     path,
		"expanded": path,
	}

	logger.WithFields(fields).Debug("Expanding path")
	path, err = homedir.Expand(path)
	if err != nil {
		logger.WithFields(fields).Error("Could not expand path")
	}
	return path, err
}

// CheckExists checks to see if the provided path exists on the machine
func CheckExists(path string) bool {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return false
	}
	var fields = logrus.Fields{
		"path": path,
	}

	logger.WithFields(fields).Debug("Checking to see if path exists")
	_, e := os.Stat(path)
	return e == nil
}

// CreateDirectory creates a directory on the machine
//   All children will be created (behavior matches mkdir -p)
func CreateDirectory(path string) error {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return err
	}
	var fields = logrus.Fields{
		"path": path,
	}

	if isDirectory(path) {
		return nil
	}
	logger.WithFields(fields).Debug("Creating directory")
	err = os.MkdirAll(path, 0755)
	if err != nil {
		logger.WithFields(fields).Warn("Failed to create directory")
		return err
	}
	logger.WithFields(fields).Debug("Directory created successfully")
	return err
}

func DeleteFile(path string) error {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// ForceTrailingSlash forces a trailing slash at the end of the path
// It will add the trailing slash only if one does not already exist
func ForceTrailingSlash(path string) string {
	if len(path) == 0 {
		return "/"
	}

	if string(path[len(path)-1]) != "/" {
		path += "/"
	}
	return path
}

// GetDirectoryContents gets the files and folders inside the provided path
func GetDirectoryContents(path string) ([]string, error) {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return nil, err
	}
	var fields = logrus.Fields{
		"path": path,
	}
	var fileNames = []string{}
	var files []os.FileInfo

	logger.WithFields(fields).Debug("Listing directory contents")
	files, err = ioutil.ReadDir(path)
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}
	return fileNames, err
}

// GetFileExtension returns the extension for the file passed in
func GetFileExtension(path string) string {
	var ext = filepath.Ext(path)
	if len(ext) > 0 {
		return ext[1:]
	}
	return ""
}

// GetFileSHA256Checksum gets the SHA-256 checksum of the file as a hex string
//   Output matches sha256sum (Linux) / shasum -a 256 (OSX)
func GetFileSHA256Checksum(path string) (string, error) {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return "", err
	}
	var fields = logrus.Fields{
		"path": path,
	}

	if err == nil {
		if isFile(path) {
			if f, err := os.Open(path); err == nil {
				defer f.Close()

				hasher := sha256.New()
				if _, err := io.Copy(hasher, f); err == nil {
					checksumString := hex.EncodeToString(hasher.Sum(nil))
					fields = logrus.Fields{
						"path":     path,
						"checksum": checksumString,
					}
					logger.WithFields(fields).Debug("Computed file checksum")
					return checksumString, err
				}
			}
		} else {
			err = errors.New(path + " is not a file")
		}
	}
	logger.WithFields(fields).Warn("Failed to retreive file checksum")
	return "", err
}

// IsDirectory returns when path exists and is a directory
// supports ~ expansion
func IsDirectory(path string) bool {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return false
	}
	var fields = logrus.Fields{
		"path": path,
	}

	logger.WithFields(fields).Debug("Checking to see if path is a directory")
	return isDirectory(path)
}

// Check to see if the path provided is a directory
func isDirectory(path string) bool {
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && stat.IsDir()
}

// IsEmptyDirectory returns when path exists and is an empty directory
// supports ~ expansion
func IsEmptyDirectory(path string) bool {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return false
	}
	var fields = logrus.Fields{
		"path": path,
	}

	logger.WithFields(fields).Debug("Checking to see if path is an empty directory")

	if file, err := os.Open(path); err == nil {
		defer file.Close()
		if !isDirectory(path) {
			return false
		}

		contents, err := file.Readdir(1)
		return (err == nil || err == io.EOF) && len(contents) == 0
	}
	return false
}

// IsFile returns when path exists and is a file
// supports ~ expansion
func IsFile(path string) bool {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return false
	}
	var fields = logrus.Fields{
		"path": path,
	}

	logger.WithFields(fields).Debug("Checking to see if path is a file")
	return isFile(path)
}

// isFile checks to see if the file exists on the filesystem
func isFile(path string) bool {
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && !stat.IsDir()
}

// LoadFileBytes loads the contents of path into a []byte if the file exists
func LoadFileBytes(path string) ([]byte, error) {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return nil, err
	}
	var fields = logrus.Fields{
		"file": path,
	}

	logger.WithFields(fields).Debug("Attempting to load file")

	if isFile(path) {
		contents, err := ioutil.ReadFile(path)
		if err == nil {
			logger.WithFields(fields).Debug("File read successfully")
			return contents, err
		}
	} else {
		err = errors.New(path + " is not a file")
	}
	logger.WithFields(fields).Info("Could not read file")
	return []byte{}, err
}

// LoadFileIfExists is deprecated in favor of LoadFileString
func LoadFileIfExists(path string) (string, error) {
	return LoadFileString(path)
}

// LoadFileString loads the contents of path into a string if the file exists
func LoadFileString(path string) (string, error) {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return "", err
	}
	var fields = logrus.Fields{
		"file": path,
	}

	logger.WithFields(fields).Debug("Attempting to load file")
	if isFile(path) {
		contents, err := ioutil.ReadFile(path)
		if err == nil {
			logger.WithFields(fields).Debug("File read successfully")
			return string(contents), err
		}
	} else {
		err = errors.New(path + " is not a file")
	}
	logger.WithFields(fields).Info("Could not read file")
	return "", err
}

// RemoveDirectory removes the directory at path from the system
// If recursive is set to true, it will remove all children as well
func RemoveDirectory(path string, recursive bool) error {
	var err error
	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return err
	}

	var fields = logrus.Fields{
		"directory": path,
	}

	logger.WithFields(fields).Debug("Attempting to remove directory")
	if isDirectory(path) {
		if recursive {
			logger.WithFields(fields).Debug("Removing directory with recursion")
			err = os.RemoveAll(path)
		} else {
			logger.WithFields(fields).Debug("Removing directory without recursion")
			err = os.Remove(path)
		}
		if err == nil {
			logger.WithFields(fields).Debug("Directory was removed")
			return nil
		}
	} else {
		err = errors.New(path + " is not a directory")
	}
	logger.WithFields(fields).Warn("Failed to remove directory")
	return err
}

// WriteFile writes contents of data to path
func WriteFile(path string, data []byte, mode os.FileMode) error {
	var err error
	var fields = logrus.Fields{
		"filename": path,
		"mode":     mode,
	}

	path, err = BuildAbsolutePathFromHome(path)
	if err != nil {
		return err
	}

	logger.WithFields(fields).Debug("Writing file")
	err = ioutil.WriteFile(path, data, mode)
	if err == nil {
		logger.WithFields(fields).Debug("Successfully wrote file")
	} else {
		logger.WithFields(fields).Warn("Failed to write file")
	}
	return err
}
