package filesystem

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

func TestHomedirExpansion(t *testing.T) {
	SetVerbosity(4)
	color.Yellow("Testing ~ expansion functionality")
	var expandedPath, err = BuildAbsolutePathFromHome("~/test-dir")
	if err != nil {
		t.Error(err)
	}
	if strings.Index(expandedPath, "/") != 0 {
		t.Error("Absolute path should start with /")
	}

	badPath := "~~/test"
	if _, err := BuildAbsolutePathFromHome(badPath); err == nil {
		t.Error("BuildAbsolutePathFromHome succeeded with bad path")
	}
	if CheckExists(badPath) {
		t.Error("CheckExists succeeded with bad path")
	}
	if err := CreateDirectory(badPath); err == nil {
		t.Error("CreateDirectory succeeded with bad path")
	}
	if err := CreateDirectory("/root/test"); err == nil {
		t.Error("CreateDirectory succeeded with at path without permissions")
	}
	if err := DeleteFile(badPath); err == nil {
		t.Error("DeleteFile succeeded with bad path")
	}
	if _, err := GetDirectoryContents(badPath); err == nil {
		t.Error("GetDirectoryContents succeeded with bad path")
	}
	if _, err := GetFileSHA256Checksum(badPath); err == nil {
		t.Error("GetFileSHA256Checksum succeeded with bad path")
	}
	if IsDirectory(badPath) {
		t.Error("IsDirectory succeeded with bad path")
	}
	if IsEmptyDirectory(badPath) {
		t.Error("IsEmptyDirectory succeeded with bad path")
	}
	if IsFile(badPath) {
		t.Error("IsFile succeeded with bad path")
	}
	if _, err := LoadFileBytes(badPath); err == nil {
		t.Error("LoadFileBytes succeeded with bad path")
	}
	if _, err := LoadFileString(badPath); err == nil {
		t.Error("LoadFileString succeeded with bad path")
	}
	if err := RemoveDirectory(badPath, false); err == nil {
		t.Error("RemoveDirectory succeeded with bad path")
	}
	if err := WriteFile(badPath, []byte{}, 755); err == nil {
		t.Error("WriteFile succeeded with bad path")
	}
	color.Yellow("Test complete")
	println()
}

func TestFilesystemOperations(t *testing.T) {
	color.Yellow("Testing Filesystem Operations")
	// Create a temp directory
	var tempDir, err = ioutil.TempDir("/tmp/", ".filesystem-test-")
	if err != nil {
		t.Error(err)
	}

	// Test CheckExists - the temp directory now exists since err == nil above
	if !CheckExists(tempDir) {
		t.Error("Check exists failed; returned false:", tempDir)
	}

	// Test Remaining Directory Functions
	testFilesystemOperations(t, tempDir+"/does-not-exist")
	testFilesystemOperations(t, tempDir+"/this/has/subfolders/that/dont/exist")

	// Verify the loading a non-existent files / folders returns properly
	color.Yellow("Test failure handling")
	if c, err := LoadFileString(tempDir + "/file-dne"); err == nil || c != "" {
		t.Error("Load non-existent file string test failed")
	}
	if c, err := LoadFileBytes(tempDir + "/file-dne"); err == nil || !reflect.DeepEqual(c, []byte{}) {
		t.Error("Load non-existent file bytes test failed")
	}
	if IsEmptyDirectory(tempDir + "/dne/") {
		t.Error("Non-existent directory says it's an empty directory")
	}
	if c, err := GetFileSHA256Checksum(tempDir + "/file-dne"); err == nil || c != "" {
		t.Error("Non-existent file checksum test failed")
	}
	if err := WriteFile(tempDir+"/dne/expect-error", []byte("test"), 0644); err == nil {
		t.Error("Write should have failed but did not return an error")
	}

	RemoveDirectory(tempDir, true)
	if CheckExists(tempDir) {
		t.Error("Recursive directory removal failed; returned true:", tempDir)
	}
	color.Yellow("Test Complete")
	println()
}

func testFilesystemOperations(t *testing.T, dir string) {
	color.Yellow("Make sure the directory doesn't exist before starting")
	if CheckExists(dir) {
		t.Error("Directory already exists:" + dir)
	}

	color.Yellow("Attempt to create the directory")
	if err := CreateDirectory(dir); err != nil {
		t.Error("Create directory failed:", dir, err)
	}
	if !CheckExists(dir) {
		t.Error("Create diretory did not create the directory properly:", dir)
	}

	color.Yellow("Try creating again now that the directory exists")
	if err := CreateDirectory(dir); err != nil {
		t.Error("Create directory failed:", dir, err)
	}
	if c, err := GetDirectoryContents(dir); err != nil || len(c) > 0 {
		t.Error("Directory is not empty:", dir)
	}

	color.Yellow("Remove the directory")
	if err := RemoveDirectory(dir, false); err != nil {
		t.Error("Directory could not be deleted:", dir)
	}
	if CheckExists(dir) {
		t.Error("Directory should have been removed but was found:", dir)
	}

	color.Yellow("Recreate the directory")
	if err := CreateDirectory(dir); err != nil {
		t.Error("Create directory failed:", dir, err)
	}
	if !CheckExists(dir) {
		t.Error("Create diretory did not create the directory properly:", dir)
	}
	if !IsEmptyDirectory(dir) {
		t.Error("Directory is reported as not empty but has no contents:", dir)
	}

	color.Yellow("Write a file inside the directory")
	var testFile = dir + "/test.file"
	var testFileContents = "test"
	var testFileBytes = []byte(testFileContents)
	if err := WriteFile(testFile, testFileBytes, 0644); err != nil {
		t.Error("Error occured while writing file:", testFile, err)
	}
	if !IsDirectory(dir) {
		t.Error("IsDirectory test failed: ", dir, " is a directory")
	}
	if IsDirectory(testFile) {
		t.Error("IsDirectory test failed: ", testFile, " is a file")
	}
	if IsFile(dir) {
		t.Error("IsFile test failed: ", dir, " is a directory")
	}
	if !IsFile(testFile) {
		t.Error("Test file should exist: ", testFile, " is a file")
	}
	// Try IsEmptyDirectory() on a file
	if IsEmptyDirectory(testFile) {
		t.Error("IsEmptyDirectory on a file tested as empty directory")
	}
	if err := DeleteFile(testFile); err != nil {
		t.Error("Delete file failed: ", testFile)
	}
	if IsFile(testFile) {
		t.Error("Test file should have been removed: ", testFile, " is a file")
	}
	WriteFile(testFile, testFileBytes, 0644)

	// Make sure the folder now has contents
	if c, err := GetDirectoryContents(dir); err != nil || len(c) == 0 {
		t.Error("Directory is empty and should not be: ", dir)
	}
	if IsEmptyDirectory(dir) {
		t.Error("Directory is reported as empty but has contents:", dir)
	}

	// Verify the contents of the file match what was intended
	if c, err := LoadFileString(testFile); err != nil || c != testFileContents {
		t.Error("File contents (string) don't match what was saved: ", testFile, "Got:", c, "Wanted:", testFileContents)
	}
	if c, err := LoadFileBytes(testFile); err != nil || !reflect.DeepEqual(c, testFileBytes) {
		t.Error("File contents (bytes) don't match what was saved: ", testFile, "Got:", c, "Wanted:", testFileBytes)
	}
	// Test deprecated function call
	if c, err := LoadFileIfExists(testFile); err != nil || c != testFileContents {
		t.Error("File contents don't match what was saved: ", testFile, "Got:", c, "Wanted:", testFileContents)
	}

	// Verify the file checksum
	if c, err := GetFileSHA256Checksum(testFile); err != nil || c != "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" {
		t.Error("File checksum incorrect for:", testFile, "Got:", c, "Wanted:", "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08")
	}

	color.Yellow("Remove the directory")
	if err := RemoveDirectory(dir, true); err != nil {
		t.Error("Directory could not be deleted:", dir)
	}
	if CheckExists(dir) {
		t.Error("Directory should have been removed but was found: " + dir)
	}

	// Attempt to remove the directory again
	if err := RemoveDirectory(dir, false); err == nil {
		t.Error("Directory could not be deleted:", dir)
	}
}

func TestTrailingSlash(t *testing.T) {
	color.Yellow("Testing force trailing slash functionality")
	var testData = map[string]string{
		"":        "/",
		"/test/":  "/test/",
		"/test":   "/test/",
		"/test//": "/test//",
	}

	var got string
	for k, v := range testData {
		got = ForceTrailingSlash(k)
		if got != v {
			t.Error("ForceTrailingSlash failed:", "Got:", got, "Wanted:", v)
		}
	}
	color.Yellow("Test Complete")
	println()
}

func TestLoggingOptions(t *testing.T) {
	color.Yellow("Testing logging options")
	var err error
	// Try different verbosity levels
	for i := uint8(0); i <= 3; i++ {
		SetVerbosity(i)
		_, err = BuildAbsolutePathFromHome("~/test/")
		if err != nil {
			t.Error(err)
		}
	}
	// Try passing in
	var logger = logrus.New()
	SetLogger(logger)
	BuildAbsolutePathFromHome("~/test")
	color.Yellow("Test Complete")
	println()
}

func TestFileExtensionFunctionality(t *testing.T) {
	var extensionTestData = map[string]string{
		"none":                "",
		"file.ext":            "ext",
		"file.bk.ext":         "ext",
		"/full/path.txt":      "txt",
		"~/relative/path.pdf": "pdf",
		"test.":               "",
	}

	var got string
	for value, expected := range extensionTestData {
		got = GetFileExtension(value)
		if got != expected {
			t.Error("Got back unexpected extension.", "Expected:", expected, "Got:", got)
		}
	}
}
