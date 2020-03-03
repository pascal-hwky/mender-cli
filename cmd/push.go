/*
Copyright © 2020 Pascal den Boef (pascal@hwky.ai)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os/exec"
	"time"

	"github.com/mendersoftware/mender-cli/client/deploy"
	"github.com/mendersoftware/mender-cli/client/deployments"
	"github.com/mendersoftware/mender-cli/client/inventory"
	"github.com/mendersoftware/mender-cli/log"
	"github.com/spf13/cobra"
)

const (
	argPushGroup = "group"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push GROUP",
	Short: "Create new release and deploy to device group.",
	Long: `Packages current directory as artifact using directory-artifact-gen.
This artifact is uploaded to the server and deployed to the devices that are in group GROUP.
The generated artifact is deleted locally after upload.`,
	Run: func(c *cobra.Command, args []string) {
		cmd, err := NewPushCmd(c, args)
		CheckErr(err)

		CheckErr(cmd.Run())
	},
}

func init() {
}

type PushCmd struct {
	server     string
	skipVerify bool
	group      string
	tokenPath  string
	deviceIds  []string
}

func NewPushCmd(cmd *cobra.Command, args []string) (*PushCmd, error) {
	server, err := cmd.Flags().GetString(argRootServer)
	if err != nil {
		return nil, err
	}

	skipVerify, err := cmd.Flags().GetBool(argRootSkipVerify)
	if err != nil {
		return nil, err
	}

	var group string
	if len(args) == 1 {
		group = args[0]
	} else {
		return nil, nil
	}

	token, err := cmd.Flags().GetString(argRootToken)
	if err != nil {
		return nil, err
	}

	if token == "" {
		token, err = getDefaultAuthTokenPath()
		if err != nil {
			return nil, err
		}
	}

	return &PushCmd{
		server:     server,
		group:      group,
		tokenPath:  token,
		skipVerify: skipVerify,
	}, nil
}

func (c *PushCmd) Run() error {

	// Get list of devices
	client := inventory.NewClient(c.server, c.skipVerify)
	devices, err := client.ListDevices(c.group, c.tokenPath)
	if err != nil {
		return err
	}
	var deviceIds []string
	for _, device := range devices {
		deviceIds = append(deviceIds, device.Id)
	}

	c.deviceIds = deviceIds
	if len(c.deviceIds) == 0 {
		log.Info("No devices in group " + c.group)
		return nil
	}

	// Create artifact TODO: infer device-type from []Devices
	artifactName := c.group + "-" + time.Now().Format(time.RFC3339)
	artifactPath := "/tmp/" + artifactName + ".mender"
	cmd := exec.Command("directory-artifact-gen", "--artifact-name", artifactName, "--device-type", "generic-armv6", "--dest-dir", "/opt/install-by-directory", "--output-path", artifactPath, ".")
	cmdErr := cmd.Run()
	if cmdErr != nil {
		log.Err("Error when creating artifact. Ensure directory-artifact-gen is installed and try again.")
		return cmdErr
	}
	log.Info("Created artifact " + artifactName + " at " + artifactPath)

	// Upload artifact
	artifactDescription := "Uploaded by mender-cli for device group " + c.group
	uploadClient := deployments.NewClient(c.server, c.skipVerify)
	uploadErr := uploadClient.UploadArtifact(artifactDescription, artifactPath, c.tokenPath, false)
	if uploadErr != nil {
		log.Err("Error when uploading artifact.")
		return uploadErr
	}

	// Deploy artifact
	deployName := "Deploy"
	deployClient := deploy.NewClient(c.server, c.skipVerify)
	deployErr := deployClient.DeployRelease(artifactName, c.deviceIds, deployName, c.tokenPath)
	if deployErr != nil {
		log.Err("Error when deploying release.")
		return deployErr
	}

	log.Info("push successful")

	return nil
}
