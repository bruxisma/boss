package main

import "strings"
import "bufio"
import "os"

import "os/exec"
import "path/filepath"

import log "github.com/sirupsen/logrus"
import "github.com/spf13/viper"

var CACHE = filepath.Join(os.ExpandEnv("$HOME"), ".cache", "boss")

func initConfig() {
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("hooks")
	viper.SetConfigName("boss")
	viper.SetEnvPrefix("boss")
	viper.AutomaticEnv()
	// Optional settings
	viper.SetDefault("build-cache", "$HOME/.cache/hugo")
	viper.SetDefault("branch", "master")
	viper.SetDefault("clean-destination-dir", true)
	viper.SetDefault("minify", true)
	viper.SetDefault("gc", true)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file '%s'", err)
	}
	log.Infof("Using config '%s'", viper.ConfigFileUsed())
	log.Infof("Destination is: %s", viper.Get("destination"))
	log.Infof("Build-Cache is: %s", viper.Get("build-cache"))
}

func cache(path string) string {
	return filepath.Join(CACHE, path)
}

func isDeployBranch(ref interface{}) bool {
	var deployBranch strings.Builder
	deployBranch.WriteString("refs/heads")
	deployBranch.WriteString(viper.GetString("branch"))
	return ref == deployBranch.String()
}

func receive() string {
	var new string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		values := strings.Fields(text)
		length := len(values)
		if length < 3 {
			log.Fatalf("Expected 3 parameters, received %d from input '%s'", length, text)
		}
		fields := log.Fields{
			"old": values[0][:7],
			"new": values[1][:7],
			"ref": values[2],
		}
		logger := log.WithFields(fields)
		logger.Info("payload received")
		if isDeployBranch(fields["ref"]) {
			logger.Infof("Deployment branch '%s' does not match. Skipping...", fields["ref"])
			continue
		}
		new = values[1]
	}
	return new
}

func hugoArgs(path string) []string {
	buildCache := os.ExpandEnv(viper.GetString("build-cache"))
	cacheDir := filepath.Join(buildCache, filepath.Base(path))
	destination := os.ExpandEnv(viper.GetString("destination"))
	args := []string{
		"--source", path,
		"--cacheDir", cacheDir,
		"--destination", destination,
	}
	if viper.GetBool("minify") {
		args = append(args, "--minify")
	}
	if viper.GetBool("clean-destination-dir") {
		args = append(args, "--cleanDestinationDir")
	}
	if viper.GetBool("gc") {
		args = append(args, "--gc")
	}
	return args
}

func build(path string) {
	args := hugoArgs(path)
	cmd := exec.Command("hugo", args...)
	output, err := cmd.CombinedOutput()
	slicedout := strings.Split(strings.TrimSpace(string(output[:])), "\n")
	fields := log.Fields{
		"cmd": "hugo",
		"src": filepath.Base(path),
		"dst": viper.GetString("destination"),
	}
	logger := log.WithFields(fields)
	for _, line := range slicedout {
		logger.Info(line)
	}
	if err != nil {
		logger.Fatal(err)
	}
}

func cleanEnvironment() {
	envvars := []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_QUARANTINE_PATH"}
	for _, v := range envvars {
		fields := log.Fields{"env": v}
		logger := log.WithFields(fields)
		logger.Debug("Unsetting environment variable")
		if err := os.Unsetenv(v); err != nil {
			logger.Errorf("Could not unset environment variable: %s", err)
		}
	}
}

// TODO: Need to move anything *git* related into its own module
func checkout(revision string) string {
	log.Debugf("creating directory '%s' in worktree cache", revision)
	worktree := filepath.Join(CACHE, revision)
	os.RemoveAll(worktree)
	if err := os.MkdirAll(worktree, os.ModePerm); err != nil {
		log.Fatalf("could not create '%s'", worktree, err)
	}
	cleanEnvironment()
	args := []string{
		"worktree",
		"add",
		worktree,
		revision,
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	slicedout := strings.Split(strings.TrimSpace(string(output[:])), "\n")
	fields := log.Fields{"git": "worktree-add", "revision": revision[:7]}
	logger := log.WithFields(fields)
	for _, line := range slicedout {
		logger.Info(line)
	}
	if err != nil {
		log.Fatalf("could not create worktree for '%s': %s", revision, err)
	}
	return worktree
}

func main() {
	formatter := &log.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true, // We're only ever used as a git push
	}
	log.SetFormatter(formatter)
	initConfig()
  // This is where the magic happens :D
	revision := receive()
	worktree := checkout(revision)
	build(worktree)
}
