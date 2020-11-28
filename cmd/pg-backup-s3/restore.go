package main

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mect/pg-backup-s3/pkg/pgdump"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/mect/pg-backup-s3/pkg/crypt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(NewRestoreCmd())
}

type restoreCmdOptions struct {
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

	name string
}

// NewRestoreCmd generates the `backup` command
func NewRestoreCmd() *cobra.Command {
	r := restoreCmdOptions{}
	c := &cobra.Command{
		Use:     "restore",
		Short:   "restore a backup",
		PreRunE: r.Validate,
		RunE:    r.RunE,
	}
	c.Flags().StringVar(&r.postgresHost, "postgres-host", "", "PostgreSQL hostname")
	c.Flags().IntVar(&r.postgresPort, "postgres-port", 5432, "PostgreSQL hostname")
	c.Flags().StringVar(&r.postgresUsername, "postgres-username", "", "PostgreSQL hostname")
	c.Flags().StringVar(&r.postgresPassword, "postgres-password", "", "PostgreSQL hostname")
	c.Flags().StringVar(&r.postgresDatabase, "postgres-database", "", "PostgreSQL hostname")

	c.Flags().StringVar(&r.encryptionKey, "encryption-key", "", "key to encrypt data with")

	c.Flags().StringVar(&r.s3Endpoint, "s3-endpoint", "", "S3 endpoint URL, leave empty for AWS")
	c.Flags().StringVar(&r.s3Bucket, "s3-bucket", "", "S3 Bucket name")
	c.Flags().StringVar(&r.s3Region, "s3-region", "", "S3 region name")
	c.Flags().StringVar(&r.s3AccessKey, "s3-access-key", "", "S3 Access Key")
	c.Flags().StringVar(&r.s3SecretKey, "s3-secret-key", "", "S3 Access Secret")

	c.Flags().StringVar(&r.name, "name", "", "Name of the the S3 file")

	c.MarkFlagRequired("postgres-host")
	c.MarkFlagRequired("postgres-username")
	c.MarkFlagRequired("postgres-password")
	c.MarkFlagRequired("postgres-database")
	c.MarkFlagRequired("encryption-key")
	c.MarkFlagRequired("s3-bucket")
	c.MarkFlagRequired("s3-region")
	c.MarkFlagRequired("name")

	// TODO: allow ambient mode
	c.MarkFlagRequired("s3-access-key")
	c.MarkFlagRequired("s3-secret-key")

	viper.BindPFlags(c.Flags())

	return c
}

func (r *restoreCmdOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (r *restoreCmdOptions) RunE(cmd *cobra.Command, args []string) error {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "pg-backup-s3")
	if err != nil {
		return err
	}

	log.Println("Downloading backup to", path.Join(tmpDir, "dump.tar.gz.enc"))
	if err := r.downloadTo(r.name, path.Join(tmpDir, "dump.tar.gz.enc")); err != nil {
		return err
	}

	log.Println("Decrypting backup to", path.Join(tmpDir, "dump.tar.gz"))
	if err := crypt.DecryptFile(path.Join(tmpDir, "dump.tar.gz.enc"), path.Join(tmpDir, "dump.tar.gz"), r.encryptionKey); err != nil {
		return err
	}

	if err := os.Remove(path.Join(tmpDir, "dump.tar.gz.enc")); err != nil {
		return err
	}

	d := pgdump.NewDump(r.postgresHost, r.postgresPort, r.postgresUsername, r.postgresPassword, r.postgresDatabase)
	if err := d.RestoreFrom(path.Join(tmpDir, "dump.tar.gz")); err != nil {
		return err
	}

	if err := os.Remove(path.Join(tmpDir, "dump.tar.gz")); err != nil {
		return err
	}

	log.Println("Restore finished, have a nice day")
	return nil
}

func (r *restoreCmdOptions) downloadTo(name, filePath string) error {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(r.s3AccessKey, r.s3SecretKey, ""),
		Endpoint:         aws.String(r.s3Endpoint),
		Region:           aws.String(r.s3Region),
		S3ForcePathStyle: aws.Bool(true),
	}

	s, err := session.NewSession(s3Config)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(s)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(r.s3Bucket),
			Key:    aws.String(name),
		})

	return err
}
