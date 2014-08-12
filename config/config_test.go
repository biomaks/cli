package config

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDemoDir(t *testing.T) {
	path, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	os.Chdir(path)

	path, err = filepath.EvalSymlinks(path)
	assert.NoError(t, err)

	path = filepath.Join(path, "exercism-demo")

	demoDir, err := demoDirectory()
	assert.NoError(t, err)
	assert.Equal(t, demoDir, path)
}

func TestExpandsTildeInExercismDirectory(t *testing.T) {
	expandedDir := ReplaceTilde("~/exercism/directory")
	assert.NotContains(t, "~", expandedDir)
}

func TestReadingWritingConfig(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	filename := Filename(tmpDir)
	assert.NoError(t, err)

	c := &Config{
		GithubUsername:    "user",
		APIKey:            "MyKey",
		ExercismDirectory: "/exercism/directory",
		Hostname:          "localhost",
	}

	c.ToFile(filename)

	loadedConfig, err := FromFile(filename)
	assert.NoError(t, err)

	assert.Equal(t, c, loadedConfig)
}

func TestDecodingConfig(t *testing.T) {
	unsanitizedJSON := `{"githubUsername":"user ","apiKey":"MyKey  ","exercismDirectory":"/exercism/directory\r\n","hostname":"localhost \r\n"}`
	sanitizedConfig := &Config{
		GithubUsername:    "user",
		APIKey:            "MyKey",
		ExercismDirectory: "/exercism/directory",
		Hostname:          "localhost",
	}
	b := bytes.NewBufferString(unsanitizedJSON)
	c, err := Decode(b)

	assert.NoError(t, err)
	assert.Equal(t, sanitizedConfig, c)
}

func TestEncodingConfig(t *testing.T) {
	currentConfig := Config{
		GithubUsername:    "user\r\n",
		APIKey:            "MyKey ",
		ExercismDirectory: "/home/user name  ",
		Hostname:          "localhost  ",
	}
	sanitizedJSON := `{"githubUsername":"user","apiKey":"MyKey","exercismDirectory":"/home/user name","hostname":"localhost"}
`

	buf := new(bytes.Buffer)
	err := currentConfig.Encode(buf)

	assert.NoError(t, err)
	assert.Equal(t, sanitizedJSON, buf.String())
}

func TestNormalizeGoPresent(t *testing.T) {
	withPreparedConfigDir(t, false, true, func(confDir, jsonPath, goPath string) {
		err := normalizeFilename(confDir)
		assert.NoError(t, err)

		assertFileExists(t, jsonPath)
		assertNoFileExists(t, goPath)
	})
}

func TestNormalizeJsonPresent(t *testing.T) {
	withPreparedConfigDir(t, true, false, func(confDir, jsonPath, goPath string) {
		err := normalizeFilename(confDir)
		assert.NoError(t, err)

		assertFileExists(t, jsonPath)
		assertNoFileExists(t, goPath)
	})
}

func TestNormalizeBothPresent(t *testing.T) {
	withPreparedConfigDir(t, true, true, func(confDir, jsonPath, goPath string) {
		err := normalizeFilename(confDir)
		assert.NoError(t, err)

		assertFileExists(t, jsonPath)
		assertFileExists(t, goPath)
	})
}

func TestNormalizeNeitherPresent(t *testing.T) {
	withPreparedConfigDir(t, false, false, func(confDir, jsonPath, goPath string) {
		err := normalizeFilename(confDir)
		assert.NoError(t, err)

		assertNoFileExists(t, jsonPath)
		assertNoFileExists(t, goPath)
	})
}

func withPreparedConfigDir(t *testing.T, jsonExists, goExists bool, fn func(configPath, goPath, jsonPath string)) {
	tmpDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)

	jsonPath := filepath.Join(tmpDir, ".exercism.json")
	goPath := filepath.Join(tmpDir, ".exercism.go")

	if jsonExists {
		f, err := os.Create(jsonPath)
		assert.NoError(t, err)
		f.Close()
	}
	if goExists {
		f, err := os.Create(goPath)
		assert.NoError(t, err)
		f.Close()
	}

	fn(tmpDir, jsonPath, goPath)

	os.Remove(tmpDir)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func assertFileExists(t *testing.T, path string) {
	if !fileExists(path) {
		t.Error("expected", path, "to exist")
	}
}

func assertNoFileExists(t *testing.T, path string) {
	if fileExists(path) {
		t.Error("expected", path, "to exist")
	}
}
