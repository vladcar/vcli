package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/fatih/color"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const defaultDotFile = ".zshenv"
const defaultAwsRegion = "eu-central-1"
const defaultAwsProfile = "default"

const accessKeyLine = "export AWS_ACCESS_KEY_ID="
const secretKeyLine = "export AWS_SECRET_ACCESS_KEY="
const sessionTokenLine = "export AWS_SESSION_TOKEN="
const regionLine = "export AWS_REGION="

var printRed = color.New(color.FgRed).Add(color.Bold).PrintlnFunc()
var printCyan = color.New(color.FgCyan).PrintlnFunc()
var printGreen = color.New(color.FgGreen).Add(color.Bold).PrintlnFunc()

var region string
var profile string
var dotFile string

var awsconfCmd = &cobra.Command{
	Use:   "awsconf",
	Short: "Assume IAM role using SSO credentials and export temporary credentials to your shell dotfile e.g. .zshenv. Prior SSO login required",
	Run: func(cmd *cobra.Command, args []string) {
		if err := viper.ReadInConfig(); err == nil {
			roleArn := viper.GetString(fmt.Sprintf("aws.%v.roleArn", profile))
			awsProfile := viper.GetString(fmt.Sprintf("aws.%v.awsProfile", profile))

			printCyan("Profile:", awsProfile)
			printCyan("Role:", roleArn)
			printCyan("Region:", region)
			printCyan("Shell dotfile:", dotFile)

			if err := assumeRole(roleArn, awsProfile); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(awsconfCmd)

	awsconfCmd.Flags().StringVar(&region, "region", defaultAwsRegion, "AWS region to use")
	awsconfCmd.Flags().StringVar(&profile, "profile", defaultAwsProfile, "AWS profile to use")
	awsconfCmd.Flags().StringVar(&dotFile, "dotfile", defaultDotFile, "Shell environment file. E.g. .zshenv or .bashprofile")
}

type STSAssumeRoleAPI interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

// TakeRole gets temporary security credentials to access resources.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If successful, an AssumeRoleOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to AssumeRole.
func TakeRole(c context.Context, api STSAssumeRoleAPI, input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	return api.AssumeRole(c, input)
}

func assumeRole(roleArn string, profile string) error {
	sharedConfig, err := loadSharedAwsProfileConfig(profile)
	if err != nil {
		printRed("unable to load aws profile config")
		return err
	}
	cfg, err := loadAwsConfig(sharedConfig)
	if err != nil {
		printRed("unable to initialize aws config")
		return err
	}
	client := sts.NewFromConfig(cfg)

	id, _ := nanoid.New()
	input := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &id,
	}

	result, err := TakeRole(context.TODO(), client, input)
	if err != nil {
		printRed("error assuming the role:")
		return err
	}

	printCyan("role assumed, session id:", id)
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dotFilePath := dirname + "/" + dotFile
	printCyan("exporting temporary AWS credentials to: ", dotFilePath)

	if err := modifyShellProfile(*result.Credentials, dotFilePath); err != nil {
		return err
	}
	printGreen("done")
	printGreen("To get started with using AWS you may need to restart your current shell.\n" +
		"This would reload your environment to include latest temporary AWS credentials.\n")
	printGreen(fmt.Sprintf("To configure your current shell, run:\n"+
		"source $HOME/%v", dotFile))

	return nil
}

func loadSharedAwsProfileConfig(profile string) (config.SharedConfig, error) {
	return config.LoadSharedConfigProfile(context.TODO(), profile)
}

func loadAwsConfig(conf config.SharedConfig) (cfg aws.Config, err error) {
	roleArn := conf.RoleARN
	mfaSerial := conf.MFASerial

	if strings.TrimSpace(roleArn) != "" && strings.TrimSpace(mfaSerial) != "" {
		// assume role with mfa
		return config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile(profile),
			config.WithAssumeRoleCredentialOptions(func(o *stscreds.AssumeRoleOptions) {
				o.RoleARN = roleArn
				o.SerialNumber = aws.String(mfaSerial)
				o.TokenProvider = stscreds.StdinTokenProvider
			}))
	} else if strings.TrimSpace(roleArn) != "" {
		// assume role without mfa
		return config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile(profile),
			config.WithAssumeRoleCredentialOptions(func(o *stscreds.AssumeRoleOptions) {
				o.RoleARN = roleArn
			}))
	} else {
		// other aws credentials
		return config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile(profile))
	}
}

func modifyShellProfile(credentials types.Credentials, dotFilePath string) error {
	accessKey := accessKeyLine + *credentials.AccessKeyId
	secretKey := secretKeyLine + *credentials.SecretAccessKey
	token := sessionTokenLine + *credentials.SessionToken
	awsRegion := regionLine + region

	f, err := os.OpenFile(dotFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()

	input, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")

	hasAccessKey := false
	hasSecretKey := false
	hasToken := false
	hasRegion := false

	for i, line := range lines {
		if strings.Contains(line, accessKeyLine) {
			lines[i] = accessKey
			hasAccessKey = true
		} else if strings.Contains(line, secretKeyLine) {
			lines[i] = secretKey
			hasSecretKey = true
		} else if strings.Contains(line, sessionTokenLine) {
			lines[i] = token
			hasToken = true
		} else if strings.Contains(line, regionLine) {
			lines[i] = awsRegion
			hasRegion = true
		}
	}

	bf := bufio.NewWriter(f)
	if hasAccessKey && hasSecretKey && hasToken && hasRegion {
		if err := f.Truncate(0); err != nil {
			return err
		}
		if _, err := f.WriteString(strings.Join(lines, "\n")); err != nil {
			return err
		}
	}
	if !hasAccessKey {
		writeToBuffer(bf, fmt.Sprintf("%v\n", accessKey))
	}
	if !hasSecretKey {
		writeToBuffer(bf, fmt.Sprintf("%v\n", secretKey))
	}
	if !hasToken {
		writeToBuffer(bf, fmt.Sprintf("%v\n", token))
	}
	if !hasRegion {
		writeToBuffer(bf, fmt.Sprintf("%v\n", awsRegion))
	}

	return bf.Flush()
}

func writeToBuffer(buf *bufio.Writer, line string) {
	if _, err := buf.WriteString(line); err != nil {
		log.Fatal(err)
	}
}
