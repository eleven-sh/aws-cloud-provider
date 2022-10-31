package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/eleven-sh/eleven/entities"
	"golang.org/x/crypto/ssh"
)

const (
	InstanceSSHPort  = 22
	InstanceRootUser = "ubuntu"
)

type RawInitInstanceScriptResults struct {
	ExitCode      string `json:"exit_code"`
	SSHHostKeys   string `json:"ssh_host_keys"`
	CloudInitLogs string `json:"cloud_init_logs"`
}

type InitInstanceScriptResults struct {
	ExitCode    string                   `json:"exit_code"`
	SSHHostKeys []entities.EnvSSHHostKey `json:"ssh_host_keys"`
}

func LookupInitInstanceScriptResults(
	ec2Client *ec2.Client,
	instancePublicIPAddress string,
	instanceSSHPort string,
	instanceLoginUser string,
	sshPrivateKeyContent string,
) (returnedInitScriptResults *InitInstanceScriptResults, returnedError error) {

	pollTimeoutChan := time.After(4 * time.Minute)
	pollSleepDuration := 4 * time.Second

	for {
		select {
		case <-pollTimeoutChan:
			cloudInitLogs, err := runCmdOnInstanceViaSSH(
				instancePublicIPAddress,
				instanceSSHPort,
				instanceLoginUser,
				sshPrivateKeyContent,
				"sudo cat /var/log/cloud-init-output.log",
			)

			if err != nil {
				cloudInitLogs = fmt.Sprintf(
					"<cannot_retrieve_cloud_init_logs> (\"%v\")",
					err,
				)
			}

			returnedError = entities.ErrEnvCloudInitError{
				Logs: cloudInitLogs,
				ErrorMessage: fmt.Sprintf(
					"Instance cloud init script timed out (\"%v\")",
					returnedError,
				),
			}
			return
		default:
			initScriptOutput, err := runCmdOnInstanceViaSSH(
				instancePublicIPAddress,
				instanceSSHPort,
				instanceLoginUser,
				sshPrivateKeyContent,
				"cat /tmp/eleven-init-results",
			)

			// Make sure timeout returns last error
			returnedError = err

			if err != nil {
				break // wait "pollSleepDuration" and retry until timeout
			}

			var initScriptResults *RawInitInstanceScriptResults
			err = json.Unmarshal([]byte(initScriptOutput), &initScriptResults)

			if err != nil {
				returnedError = entities.ErrEnvCloudInitError{
					Logs: initScriptOutput,
					ErrorMessage: fmt.Sprintf(
						"Instance cloud init script exited with invalid JSON (\"%v\").",
						err,
					),
				}
				return
			}

			if initScriptResults.ExitCode != "0" {
				returnedError = entities.ErrEnvCloudInitError{
					Logs: initScriptResults.CloudInitLogs,
					ErrorMessage: fmt.Sprintf(
						"Instance cloud init script exited with code \"%s\".",
						initScriptResults.ExitCode,
					),
				}
				return
			}

			parsedSSHHostKeys, err := entities.ParseSSHHostKeysForEnv(
				initScriptResults.SSHHostKeys,
			)

			if err != nil {
				returnedError = entities.ErrEnvCloudInitError{
					Logs: initScriptResults.SSHHostKeys,
					ErrorMessage: fmt.Sprintf(
						"Instance cloud init script exited with invalid SSH host keys (\"%v\").",
						err,
					),
				}
				return
			}

			returnedInitScriptResults = &InitInstanceScriptResults{
				ExitCode:    initScriptResults.ExitCode,
				SSHHostKeys: parsedSSHHostKeys,
			}
			return
		} // <- end of select

		time.Sleep(pollSleepDuration)
	} // <- end of for
}

func WaitForSSHAvailableInInstance(
	ec2Client *ec2.Client,
	instancePublicIPAddress string,
	instanceSSHPort string,
) (returnedError error) {

	pollTimeoutChan := time.After(4 * time.Minute)
	pollSleepDuration := 4 * time.Second

	SSHConnTimeout := 8 * time.Second

	for {
		select {
		case <-pollTimeoutChan:
			return
		default:
			conn, err := net.DialTimeout(
				"tcp",
				net.JoinHostPort(
					instancePublicIPAddress,
					instanceSSHPort,
				),
				SSHConnTimeout,
			)

			// Make sure timeout returns last error
			returnedError = err

			if err != nil {
				break // wait "pollSleepDuration" and retry until timeout
			}

			conn.Close()
			return
		}

		time.Sleep(pollSleepDuration)
	}
}

func runCmdOnInstanceViaSSH(
	instancePublicIPAddress string,
	instanceSSHPort string,
	loginUser string,
	privateKeyContent string,
	cmd string,
) (string, error) {

	signer, err := ssh.ParsePrivateKey([]byte(privateKeyContent))

	if err != nil {
		return "", err
	}

	SSHConnTimeout := time.Second * 8

	config := &ssh.ClientConfig{
		User: loginUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         SSHConnTimeout,
	}

	client, err := ssh.Dial(
		"tcp",
		net.JoinHostPort(
			instancePublicIPAddress,
			instanceSSHPort,
		),
		config,
	)

	if err != nil {
		return "", err
	}

	session, err := client.NewSession()

	if err != nil {
		return "", err
	}

	defer session.Close()

	var output bytes.Buffer
	session.Stdout = &output

	err = session.Run(cmd)

	if err != nil {
		return "", err
	}

	return output.String(), nil
}
