package main

import (
	"text/tabwriter"
	"time"
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
	return fmt.Sprintf("%s.sh", strings.ReplaceAll(j.Name, " ", "-"))
}

type BuildConfig struct {
	Jobs []BuildJob
}

type JobStatus struct {
	Name string
	Status string
	Duration string
}

type BuildContext struct {
	WorkDir       string
	TempFolderName string
	TempFolder    string
	DockerWorkDir string

	currentJob string
	jobStartTime time.Time
	JobStatus map[string]*JobStatus
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

	buildContext, err := NewBuildContext(buildConfig.Jobs)
	if err != nil {
		log.Printf("ERROR: Could not create build context: %v\r\n", err)
		os.Exit(1)
	}

	log.Printf("Starting build execution\r\n")
	err = buildContext.ExecuteBuild(&buildConfig)
	if err != nil {
		log.Printf("ERROR: Execution failed for build configuration %s: %v\r\n", configFile, err)
		buildContext.PrintJobStatus()
		os.Exit(1)
	}
	buildContext.PrintJobStatus()
}

func NewBuildContext(jobs []BuildJob) (*BuildContext, error) {
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
		JobStatus: make(map[string]*JobStatus),
	}

	for _, job := range jobs {
		buildContext.JobStatus[job.Name] = &JobStatus{
			Name: job.Name,
			Status: "NotRun",
			Duration: "None",
		}
	}

	err = os.MkdirAll(buildContext.TempFolder, 666)
	if err != nil {
		return nil, err
	}

	return &buildContext, nil
}

func (c *BuildContext) StartJob(jobName string) {
	c.currentJob = jobName
	c.jobStartTime = time.Now()
}

func (c *BuildContext) FinishJob(status string) {
	jobStatus := c.JobStatus[c.currentJob]

	jobStatus.Status = status
	jobStatus.Duration = time.Since(c.jobStartTime).String()
}

func (c *BuildContext) PrintJobStatus() {
	
	LogSeparator()	
	log.Println()
	
	logWriter := &LogWriter{}
	writer := tabwriter.NewWriter(logWriter, 0, 0, 4, ' ', 0)
	fmt.Fprintf(writer, "Job\tStatus\tDuration\n")
	fmt.Fprintf(writer, "---\t------\t--------\n")
	for _, v := range c.JobStatus {		
		fmt.Fprintf(writer, "%s\t%s\t%s\n", v.Name, v.Status, v.Duration)
	}
	writer.Flush()
	
	log.Println()
	LogSeparator()
}

func (c *BuildContext) ExecuteBuild(buildConfig *BuildConfig) error {
	for _, job := range buildConfig.Jobs {
		LogSeparator()
		log.Printf("Execute job: '%s'\r\n", job.Name)
		log.Println()
		err := c.ExecuteDockerJob(&job)
		log.Println()
		if err != nil {
			return err
		}
		log.Println("SUCCESS!")
	}

	return nil
}

func (c *BuildContext) ExecuteDockerJob(job *BuildJob) error {
	status := "OK"
	c.StartJob(job.Name)
	defer func() {
		c.FinishJob(status)
	}()


	err := c.CreateEntryPointScript(job)
	if err != nil {
		status = "EntryPointCreationError"
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

	logWriter := &LogWriter{}
	logWriter.SetDockerJobName(job.Name)
	dockerCmd := exec.Command("docker", dockerArgs...)
	dockerCmd.Stdout = logWriter
	dockerCmd.Stderr = logWriter
	dockerCmd.Env = os.Environ()

	err = dockerCmd.Run()
	if err != nil {
		status = "Failed"
		return err
	}

	return nil
}

func (c *BuildContext) CreateEntryPointScript(job *BuildJob) error {
	file := path.Join(c.TempFolder, job.GetEntryPointScriptName())

	builder := strings.Builder{}

	builder.WriteString("#!/bin/sh\n")
	builder.WriteString("set -e\n")

	builder.WriteString(fmt.Sprintf("echo /# cd %s\n", c.DockerWorkDir))
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

func LogSeparator() {
	log.Println("-------------------------------------------------------------------")
}