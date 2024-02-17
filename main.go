package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"

	"github.com/go-ini/ini"

	"acloudguru-sandbox/authproviders"
)

const acgSandbox = "acloudguru-sandbox"
const acgSandboxesUrl = "https://learn.acloud.guru/cloud-playground/cloud-sandboxes"

func main() {
	log.SetPrefix(fmt.Sprintf("%s: ", acgSandbox))
	command, authproviderid := CheckSubCommand(acgSandbox, os.Args)

	log.Printf("%s [AuthProvider: %s]", command, authproviderid)

	authprovider := AuthProviderFactory(authproviderid)

	log.Printf("AuthId: %s", authprovider.AuthId())

	username, password := GetGitCredentials(acgSandboxesUrl)
	page := authprovider.Login(acgSandboxesUrl, username, password)
	switch command {
	case "current":
		command = DetectSandbox(page)
	case "stop":
		page = StopSandbox(page)
	case "aws":
		page = StartSandbox(page, "AWS")
	case "azure":
		page = StartSandbox(page, "Azure")
	case "gcloud":
		page = StartSandbox(page, "Google Cloud")
	}
	switch command {
	case "stop":
		log.Printf("Stopped\n")
	case "aws":
		ConfigureAwsSandbox(page)
	case "azure":
		ConfigureAzureSandbox(page)
	case "gcloud":
		ConfigureGcloudSandbox(page)
	}
	Logout(page)
}

type AuthProvider interface {
	AuthId() string
	Login(acgSandboxesUrl string, username string, password string) *rod.Page
}

func AuthProviderFactory(authproviderid string) AuthProvider {

	// All auth providers
	authproviders := [2]AuthProvider{
		&authproviders.AuthGuru{},
		&authproviders.AuthGoogle{},
	}

	for _, authprovider := range authproviders {
		if authproviderid == authprovider.AuthId() {
			return authprovider
		}
	}
	return nil
}

func CheckSubCommand(commandexec string, args []string) (command string, authproviderid string) {
	paramsSyntax := "<current|stop|aws|azure|gcloud> [-rod=...]"
	if len(args) < 2 {
		log.Fatalf("missing sandbox command:\n\t%s %s\n", commandexec, paramsSyntax)
	}
	if len(args) > 4 {
		log.Fatalf("unexpected arguments: %s\n\t%s %s\n", strings.Join(args[1:], " "), commandexec, paramsSyntax)
	}
	if len(args) == 3 && !strings.HasPrefix(args[2], "-rod=") && !strings.HasPrefix(args[2], "-auth=") {
		log.Fatalln("2")
	}
	if len(args) == 4 && !strings.HasPrefix(args[2], "-rod=") && !strings.HasPrefix(args[2], "-auth=") && !strings.HasPrefix(args[3], "-rod=") && !strings.HasPrefix(args[3], "-auth=") {
		log.Fatalln("2")
	}
	// Get the auth provider
	authproviderid = "guru"
	for _, arg := range args {
		if strings.HasPrefix(arg, "-auth=") {
			authproviderid = strings.Split(arg, "=")[1]
		}
	}
	// Get the command
	switch args[1] {
	case "current", "stop", "aws", "azure", "gcloud":
		return args[1], authproviderid
	default:
		log.Fatalf("unknown sandbox command: %s\n\t%s %s\n", args[1], commandexec, paramsSyntax)
		// unreached
		return "unknown", authproviderid
	}
}

// https://pkg.go.dev/golang.org/x/tools/cmd/auth/gitauth
func GetGitCredentials(url string) (username string, password string) {
	log.Printf("git credentials for url: %s\n", url)
	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("url=%s\n", url))
	out := RunCmd(cmd, 3)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		frags := strings.SplitN(line, "=", 2)
		if len(frags) != 2 {
			continue // Ignore unrecognized response lines.
		}
		switch strings.TrimSpace(frags[0]) {
		case "username":
			username = frags[1]
		case "password":
			password = frags[1]
		}
	}
	log.Printf("git username: %s\n", username)
	return username, password
}

func RunCmd(cmd *exec.Cmd, logArgsCount int) []byte {
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot get user HOME: %v\n", err)
	}
	cmd.Dir = userHome
	cmd.Stderr = os.Stderr
	var out []byte = nil

	if cmd.Stdout != nil {
		err = cmd.Run()
	} else {
		out, err = cmd.Output()
	}

	if err != nil {
		logCmd := strings.Join(cmd.Args[:logArgsCount], " ")
		log.Fatalf("'%s' failed: %v\n", logCmd, err)
	}
	return out
}

func Logout(page *rod.Page) *rod.Page {
	page.MustNavigate("https://learn.acloud.guru/logout")
	return page
}

func DetectSandbox(page *rod.Page) string {
	current := "stop"
	page.Race().
		ElementR("button", "Start").MustHandle(func(e *rod.Element) {
		current = "stop"
	}).
		ElementR("h3", "AWS Sandbox").MustHandle(func(e *rod.Element) {
		current = "aws"
	}).
		ElementR("h3", "Azure Sandbox").MustHandle(func(e *rod.Element) {
		current = "azure"
	}).
		ElementR("h3", "Google Cloud Sandbox").MustHandle(func(e *rod.Element) {
		current = "gcloud"
	}).
		MustDo()
	return current
}

func StartSandbox(page *rod.Page, target string) *rod.Page {
	sandboxHeading := fmt.Sprintf("%s Sandbox", target)
	startButtonText := fmt.Sprintf("Start %s Sandbox", target)
	for targetReached := false; !targetReached; {
		page.Race().
			ElementR("h3", sandboxHeading).MustHandle(func(e *rod.Element) {
			targetReached = true
		}).
			ElementR("button", startButtonText).MustHandle(func(e *rod.Element) {
			e.MustClick()
		}).
			ElementR("button", "Delete Sandbox").MustHandle(func(e *rod.Element) {
			DeleteSandbox(e, page)
		}).
			MustDo()
	}
	return page
}

func DeleteSandbox(e *rod.Element, page *rod.Page) {
	e.MustClick()
	buttons := page.MustElements("button")
	buttons.Last().MustClick()
	page.MustWaitStable()
}

func StopSandbox(page *rod.Page) *rod.Page {
	for targetReached := false; !targetReached; {
		page.Race().
			ElementR("button", "Start").MustHandle(func(e *rod.Element) {
			targetReached = true
		}).
			ElementR("button", "Delete Sandbox").MustHandle(func(e *rod.Element) {
			DeleteSandbox(e, page)
		}).
			MustDo()
	}
	return page
}

func ConfigureAwsSandbox(page *rod.Page) {
	awsAccessKeyId, awsSecretAccessKey := GetAwsSandboxCredentials(page)
	WriteAwsCredentialsFile(acgSandbox, awsAccessKeyId, awsSecretAccessKey)
	CheckAwsCredentials(acgSandbox)
}

func GetAwsSandboxCredentials(page *rod.Page) (string, string) {
	elements := page.MustElements("input[aria-label='Copy to clipboard']")
	aws_username := elements[0].MustText()
	aws_password := elements[1].MustText()
	aws_url := elements[2].MustText()
	aws_access_key_id := elements[3].MustText()
	aws_secret_access_key := elements[4].MustText()
	log.Printf("Aws username: %s\n", aws_username)
	log.Printf("Aws password: %s\n", aws_password)
	log.Printf("Aws URL: %s\n", aws_url)
	log.Printf("Aws access key ID: %s\n", aws_access_key_id)
	log.Printf("Aws secret access key: %s\n", aws_secret_access_key)
	return aws_access_key_id, aws_secret_access_key
}

func WriteAwsCredentialsFile(profile string, awsAccessKeyId string, awsSecretAccessKey string) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot get user HOME: %v\n", err)
	}

	awsFolder := filepath.Join(userHome, ".aws")
	err = os.MkdirAll(awsFolder, os.ModePerm)
	if err != nil {
		log.Fatalf("Cannot access AWS config folder: %v\n", err)
	} else {
		log.Printf("Aws config folder: %s\n", awsFolder)
	}

	awsCredentialsFile := filepath.Join(awsFolder, "credentials")
	awsCredentials, err := ini.LooseLoad(awsCredentialsFile)
	if err != nil {
		log.Fatalf("Cannot read AWS credentials file: %v\n", err)
	}
	log.Printf("Aws proflie: %s\n", profile)
	awsCredentials.Section(profile).Key("aws_access_key_id").SetValue(awsAccessKeyId)
	awsCredentials.Section(profile).Key("aws_secret_access_key").SetValue(awsSecretAccessKey)
	awsCredentials.SaveTo(awsCredentialsFile)
}

func CheckAwsCredentials(profile string) {
	cmd := exec.Command("aws", "iam", "get-user", "--profile", profile)
	cmd.Stdout = os.Stdout
	_ = RunCmd(cmd, 3)
}

func ConfigureAzureSandbox(page *rod.Page) {
	azure_username, azure_password := GetAzureSandboxCredentials(page)
	LoginAzureCli(azure_username, azure_password)
}

func GetAzureSandboxCredentials(page *rod.Page) (string, string) {
	elements := page.MustElements("input[aria-label='Copy to clipboard']")
	azure_username := elements[0].MustText()
	azure_password := elements[1].MustText()
	azure_url := elements[2].MustText()
	azure_application_client_id := elements[3].MustText()
	azure_secret := elements[4].MustText()

	if true {
		log.Printf("Azure username: %s\n", azure_username)
		log.Printf("Azure password: %s\n", azure_password)
		log.Printf("Azure URL: %s\n", azure_url)
		log.Printf("Azure application client ID: %s\n", azure_application_client_id)
		log.Printf("Azure secret: %s\n", azure_secret)
	}
	return azure_username, azure_password
}

func LoginAzureCli(azure_username string, azure_password string) {
	cmd := exec.Command("az", "login", "--user", azure_username, "--password", azure_password)
	cmd.Stdout = os.Stdout
	_ = RunCmd(cmd, 2)
}

func ConfigureGcloudSandbox(page *rod.Page) {
	gcloud_service_account_credentials := GetGoogleCloudSandboxCredentials(page)
	LoginGoogleCloudCli(gcloud_service_account_credentials)
}

func GetGoogleCloudSandboxCredentials(page *rod.Page) string {
	elements := page.MustElements("input[aria-label='Copy to clipboard']")
	gcloud_username := elements[0].MustText()
	gcloud_password := elements[1].MustText()
	gcloud_url := elements[2].MustText()
	gcloud_service_account_credentials := elements[3].MustText()

	if true {
		log.Printf("Google Cloud username: %s\n", gcloud_username)
		log.Printf("Google Cloud password: %s\n", gcloud_password)
		log.Printf("Google Cloud URL: %s\n", gcloud_url)
		log.Printf("Google Cloud service account credentials: %s\n", gcloud_service_account_credentials)
	}
	return gcloud_service_account_credentials
}

func LoginGoogleCloudCli(gcloud_service_account_credentials string) {
	tempFile, err := os.CreateTemp("", "sample")
	if err != nil {
		log.Fatalf("cannot create temporary file: %v\n", err)
	}
	defer os.Remove(tempFile.Name())
	_, err = tempFile.WriteString(gcloud_service_account_credentials)
	if err != nil {
		log.Fatalf("cannot write service accout credentials: %v\n", err)
	}
	keyArgument := fmt.Sprintf("--key=%s", tempFile.Name())
	cmd := exec.Command("gcloud", "auth", "activate-service-account", keyArgument)
	cmd.Stdout = os.Stdout
	_ = RunCmd(cmd, 3)
}
