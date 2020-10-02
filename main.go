package goPSRemoting

import (
	"bytes"
	"errors"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh"
)

func runCommand(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)

	var out bytes.Buffer
	var err bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &err
	cmd.Run()

	// convert err to an error type if there is an error returned
	var e error
	if err.String() != "" {
		e = errors.New(err.String())
	}

	return strings.TrimRight(out.String(), "\r\n"), e
}

func RunPowershellCommand(username string, password string, server string, command string, usessl string, usessh string) (string, error) {
	var pscommand string
	var out string
	var err error

	if runtime.GOOS == "windows" {
		pscommand = "powershell.exe"
	} else {
		pscommand = "pwsh"
	}

	var winRMPre string

	if usessh == "1" {
		winRMPre = "$s = New-PSSession -HostName " + server + " -Username " + username + " -SSHTransport"
	} else {
		winRMPre = "$SecurePassword = '" + password + "' | ConvertTo-SecureString -AsPlainText -Force; $cred = New-Object System.Management.Automation.PSCredential -ArgumentList '" + username + "', $SecurePassword; $s = New-PSSession -ComputerName " + server + " -Credential $cred -Authentication Negotiate"
	}

	var winRMPost string
	winRMPost = "; Invoke-Command -Session $s -Scriptblock { " + command + " }; Remove-PSSession $s"

	var winRMCommand string

	if usessl == "1" {
		winRMCommand = winRMPre + " -UseSSL" + winRMPost
	} else {
		winRMCommand = winRMPre + winRMPost
	}

	if pscommand == "pwsh" && usessh == "1" {
		var errconn error
		var output []byte
		client, session, errconn := connectToHost(username, password, server)
		if errconn != nil {
			panic(errconn)
		}
		output, err = session.CombinedOutput("pwsh -Command " + command)
		out = string(output)
		client.Close()
	} else {
		out, err = runCommand(pscommand, "-command", winRMCommand)
	}

	return out, err
}

func connectToHost(user string, pass string, host string) (*ssh.Client, *ssh.Session, error) {

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(pass)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}
