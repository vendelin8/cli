package cmd

import (
	"fmt"
	"time"

	client "github.com/semaphoreci/cli/api/client"
	models "github.com/semaphoreci/cli/api/models"
	"github.com/semaphoreci/cli/cmd/ssh"
	"github.com/semaphoreci/cli/cmd/utils"
	"github.com/semaphoreci/cli/config"
	"github.com/spf13/cobra"
)

func NewDebugProjectCmd() *cobra.Command {
	var DebugProjectCmd = &cobra.Command{
		Use:     "project [NAME]",
		Short:   "Debug a project",
		Long:    ``,
		Aliases: []string{"prj", "projects"},
		Args:    cobra.ExactArgs(1),
		Run:     RunDebugProjectCmd,
	}

	DebugProjectCmd.Flags().Duration(
		"duration",
		60*time.Minute,
		"duration of the debug session in seconds")

	DebugProjectCmd.Flags().String(
		"machine-type",
		"e1-standard-2",
		"machine type to use for debugging; default: e1-standard-2")

	return DebugProjectCmd
}

func RunDebugProjectCmd(cmd *cobra.Command, args []string) {
	publicKey, err := config.GetPublicSshKeyForDebugSession()
	utils.Check(err)

	machineType, err := cmd.Flags().GetString("machine-type")
	utils.Check(err)

	duration, err := cmd.Flags().GetDuration("duration")

	utils.Check(err)

	projectName := args[0]
	pc := client.NewProjectV1AlphaApi()
	project, err := pc.GetProject(projectName)

	utils.Check(err)

	jobName := fmt.Sprintf("Debug Session for %s", projectName)
	job := models.NewJobV1Alpha(jobName)

	job.Spec = &models.JobV1AlphaSpec{}
	job.Spec.Agent.Machine.Type = machineType
	job.Spec.Agent.Machine.OsImage = "ubuntu1804"
	job.Spec.ProjectId = project.Metadata.Id

	job.Spec.Commands = []string{
		fmt.Sprintf("echo '%s' >> .ssh/authorized_keys", publicKey),
		fmt.Sprintf("sleep %d", int(duration.Seconds())),
	}

	fmt.Printf("* Creating debug session for project '%s'\n", projectName)
	fmt.Printf("* Setting duration to %d minutes\n", int(duration.Minutes()))

	sshIntroMessage := `
Semaphore CI Debug Session.

  - Checkout your code with ` + "`checkout`" + `
  - Leave the session with ` + "`exit`" + `

Documentation: https://docs.semaphoreci.com/article/75-debugging-with-ssh-access.
`

	ssh.StartDebugSession(job, sshIntroMessage)
}
