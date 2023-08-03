package executor

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type S3 struct {
	mc         *minio.Client
	bucketName string
	storageDir string
}

func CreateS3(endpoint string, region string, accessKey string, secretAccessKey string, secure bool, insecureSkipVerify bool, bucketName string, storageDir string) (*S3, error) {
	transport := http.DefaultTransport
	transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKey, secretAccessKey, ""),
		Secure:    secure,
		Region:    region,
		Transport: transport,
	})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	context := context.Background()
	bucket, err := minioClient.BucketExists(context, bucketName)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Bucket": bucket, "BucketName": bucketName}).Fatal("Failed to check if bucket exists")
		return nil, err
	}
	if !bucket {
		err = minioClient.MakeBucket(context, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.WithFields(log.Fields{"Error": err, "Bucket": bucket}).Fatal("Failed to create bucket")
			return nil, err
		}
	}

	return &S3{mc: minioClient, bucketName: bucketName, storageDir: storageDir}, nil
}

func (s3 *S3) Download(filepath string) (*[]byte, error) {
	file, err := s3.mc.GetObject(context.Background(), s3.bucketName, filepath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	defer file.Close()

	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	b := buffer.Bytes()
	return &b, err
}

func (s3 *S3) List(directory string) ([]string, error) {
	objs := s3.mc.ListObjects(context.Background(), s3.bucketName, minio.ListObjectsOptions{Prefix: directory})
	keys := []string{}
	for obj := range objs {
		if obj.Err != nil {
			return keys, obj.Err
		}
		keys = append(keys, obj.Key)
	}
	return keys, nil
}
