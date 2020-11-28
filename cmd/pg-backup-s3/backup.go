package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/mect/pg-backup-s3/pkg/crypt"

	"github.com/mect/pg-backup-s3/pkg/pgdump"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(NewBackupCmd())
}

type backupCmdOptions struct {
	postgresHost     string
	postgresPort     int
	postgresUsername string
	postgresDatabase string
	postgresPassword string

	encryptionKey string

	s3Endpoint  string
	s3Region    string
	s3AccessKey string
	s3SecretKey string
	s3Bucket    string

	namePrefix string
}

// NewBackupCmd generates the `backup` command
func NewBackupCmd() *cobra.Command {
	b := backupCmdOptions{}
	c := &cobra.Command{
		Use:     "backup",
		Short:   "take a backup",
		PreRunE: b.Validate,
		RunE:    b.RunE,
	}
	c.Flags().StringVar(&b.postgresHost, "postgres-host", "", "PostgreSQL hostname")
	c.Flags().IntVar(&b.postgresPort, "postgres-port", 5432, "PostgreSQL hostname")
	c.Flags().StringVar(&b.postgresUsername, "postgres-username", "", "PostgreSQL hostname")
	c.Flags().StringVar(&b.postgresPassword, "postgres-password", "", "PostgreSQL hostname")
	c.Flags().StringVar(&b.postgresDatabase, "postgres-database", "", "PostgreSQL hostname")

	c.Flags().StringVar(&b.encryptionKey, "encryption-key", "", "key to encrypt data with")

	c.Flags().StringVar(&b.s3Endpoint, "s3-endpoint", "", "S3 endpoint URL, leave empty for AWS")
	c.Flags().StringVar(&b.s3Bucket, "s3-bucket", "", "S3 Bucket name")
	c.Flags().StringVar(&b.s3Region, "s3-region", "", "S3 region name")
	c.Flags().StringVar(&b.s3AccessKey, "s3-access-key", "", "S3 Access Key")
	c.Flags().StringVar(&b.s3SecretKey, "s3-secret-key", "", "S3 Access Secret")

	c.Flags().StringVar(&b.namePrefix, "name-prefix", "", "Prefix for the S3 file")

	c.MarkFlagRequired("postgres-host")
	c.MarkFlagRequired("postgres-username")
	c.MarkFlagRequired("postgres-password")
	c.MarkFlagRequired("postgres-database")
	c.MarkFlagRequired("encryption-key")
	c.MarkFlagRequired("s3-bucket")
	c.MarkFlagRequired("s3-region")
	c.MarkFlagRequired("name-prefix")

	// TODO: allow ambient mode
	c.MarkFlagRequired("s3-access-key")
	c.MarkFlagRequired("s3-secret-key")

	viper.BindPFlags(c.Flags())

	return c
}

func (b *backupCmdOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (b *backupCmdOptions) RunE(cmd *cobra.Command, args []string) error {
	d := pgdump.NewDump(b.postgresHost, b.postgresPort, b.postgresUsername, b.postgresPassword, b.postgresDatabase)
	tmpDir, err := ioutil.TempDir(os.TempDir(), "pg-backup-s3")
	if err != nil {
		return err
	}

	log.Println("Writing backup to", path.Join(tmpDir, "dump.tar.gz"))
	if err := d.DumpToDisk(path.Join(tmpDir, "dump.tar.gz")); err != nil {
		return err
	}

	log.Println("Encrypting backup to", path.Join(tmpDir, "dump.tar.gz.enc"))
	if err := crypt.EncryptFile(path.Join(tmpDir, "dump.tar.gz"), path.Join(tmpDir, "dump.tar.gz.enc"), b.encryptionKey); err != nil {
		return err
	}

	if err := os.Remove(path.Join(tmpDir, "dump.tar.gz")); err != nil {
		return err
	}

	dst := strings.Join([]string{b.namePrefix, time.Now().Format(time.RFC3339), ".tar.gz.enc"}, "-")
	log.Println("Uploading backup to", dst)
	if err := b.upload(path.Join(tmpDir, "dump.tar.gz.enc"), dst); err != nil {
		return err
	}

	if err := os.Remove(path.Join(tmpDir, "dump.tar.gz.enc")); err != nil {
		return err
	}

	log.Println("Backup finished, have a nice day")
	return nil
}

func (b *backupCmdOptions) upload(filePath string, dst string) error {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(b.s3AccessKey, b.s3SecretKey, ""),
		Endpoint:         aws.String(b.s3Endpoint),
		Region:           aws.String(b.s3Region),
		S3ForcePathStyle: aws.Bool(true),
	}

	s, err := session.NewSession(s3Config)
	if err != nil {
		return err
	}

	// Open the file for use
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket: aws.String(b.s3Bucket),
		Key:    aws.String(dst),
		ACL:    aws.String("private"), // ALWAYS
		Body:   file,
	})

	return err
}
