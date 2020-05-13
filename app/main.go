package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

type BuildJob struct {
	Name    string
	Image   string
	Scripts []string
}

func (j *BuildJob) GetEntryPointScriptName() string {
	return fmt.Sprintf("%s.sh", j.Name)
}

type BuildConfig struct {
	Jobs []BuildJob
}

type BuildContext struct {
	WorkDir       string
	TempFolderName string
	TempFolder    string
	DockerWorkDir string
}

func main() {
	configFile := ".gobuild.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	log.Printf("Reading config file from %s\r\n", configFile)
	configYaml, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("ERROR: Could not read config file %s: %v\r\n", configFile, err)
		os.Exit(1)
	}

	log.Printf("Parsing yaml configuration\r\n")
	buildConfig := BuildConfig{}
	err = yaml.Unmarshal(configYaml, &buildConfig)
	if err != nil {
		log.Printf("ERROR: Could not parse yaml in config file %s: %v\r\n", configFile, err)
		os.Exit(1)
	}

	buildContext, err := NewBuildContext()
	if err != nil {
		log.Printf("ERROR: Could not create build context: %v\r\n", err)
		os.Exit(1)
	}

	log.Printf("Starting build execution\r\n")
	err = buildContext.ExecuteBuild(&buildConfig)
	if err != nil {
		log.Printf("ERROR: Execution of build configuration %s: %v\r\n", configFile, err)
		os.Exit(1)
	}
}

func NewBuildContext() (*BuildContext, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	tempFolderName := ".gobuild"

	buildContext := BuildContext{
		WorkDir:       workDir,
		TempFolderName: tempFolderName,
		TempFolder:    path.Join(workDir, tempFolderName),
		DockerWorkDir: "/var/gobuild",
	}

	err = os.MkdirAll(buildContext.TempFolder, 666)
	if err != nil {
		return nil, err
	}

	return &buildContext, nil
}

func (c *BuildContext) ExecuteBuild(buildConfig *BuildConfig) error {
	for _, job := range buildConfig.Jobs {
		log.Printf("Execute job %s\r\n", job.Name)
		err := c.ExecuteDockerJob(&job)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *BuildContext) ExecuteDockerJob(job *BuildJob) error {
	err := c.CreateEntryPointScript(job)
	if err != nil {
		return err
	}

	dockerArgs := make([]string, 0)

	dockerArgs = append(dockerArgs, "run")

	// share volume with build folder
	dockerArgs = append(dockerArgs, "-v")
	dockerArgs = append(dockerArgs, fmt.Sprintf("%s:%s", c.WorkDir, c.DockerWorkDir))

	// use prepared entry point
	dockerArgs = append(dockerArgs, "--entrypoint")
	dockerArgs = append(dockerArgs, fmt.Sprintf("%s/%s/%s", c.DockerWorkDir, c.TempFolderName, job.GetEntryPointScriptName()))

	dockerArgs = append(dockerArgs, job.Image)

	dockerCommandText := "docker " + strings.Join(dockerArgs, " ")
	log.Printf("Executing docker command: %s", dockerCommandText)

	dockerCmd := exec.Command("docker", dockerArgs...)
	dockerCmd.Stdout = LogWriter{ JobName: job.Name }
	dockerCmd.Stderr = LogWriter{ JobName: job.Name }
	dockerCmd.Env = os.Environ()

	err = dockerCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (c *BuildContext) CreateEntryPointScript(job *BuildJob) error {
	file := path.Join(c.TempFolder, job.GetEntryPointScriptName())

	builder := strings.Builder{}

	builder.WriteString("#!/bin/sh\n")
	builder.WriteString("set -e\n")

	builder.WriteString(fmt.Sprintf("echo Switch workdir to %s\n", c.DockerWorkDir))
	builder.WriteString(fmt.Sprintf("cd %s\n", c.DockerWorkDir))

	for _, script := range job.Scripts {
		builder.WriteString(fmt.Sprintf("echo /# %s\n", script))
		builder.WriteString(script)
		builder.WriteString("\n")
	}

	fileContent := builder.String()
	err := ioutil.WriteFile(file, []byte(fileContent), 666)
	if err != nil {
		return err
	}

	return nil
}
