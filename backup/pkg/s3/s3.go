package s3

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type S3 struct {
	mc         *minio.Client
	bucketName string
	storageDir string
}

func CreateS3Client(endpoint string, region string, accessKey string, secretAccessKey string, secure bool, insecureSkipVerify bool, bucketName string, storageDir string) (*S3, error) {
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

func (s3 *S3) Upload(path string, filename string, size int64) (int64, int64, error) {
	start := time.Now()
	f, err := os.Open(path + "/" + filename)

	res, err := s3.mc.PutObject(context.Background(), s3.bucketName, s3.storageDir+"/"+filename, f, size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.Errorln(err)
		return 0, -1, err
	}
	end := time.Now()
	execTime := end.Unix() - start.Unix()

	return execTime, res.Size, nil
}

func (s3 *S3) GetFile(filepath string) (*[]byte, error) {
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

func (s3 *S3) FileExists(filepath string) bool {
	_, err := s3.mc.StatObject(context.Background(), s3.bucketName, filepath, minio.StatObjectOptions{})
	if err != nil {
		log.Errorln(err)
		return false
	}

	return true
}

func (s3 *S3) RemoveFile(filepath string) error {
	return s3.mc.RemoveObject(context.Background(), s3.bucketName, filepath, minio.RemoveObjectOptions{})
}

func (s3 *S3) RemoveDirectory(directory string) error {
	objs := s3.mc.ListObjects(context.Background(), s3.bucketName, minio.ListObjectsOptions{Prefix: directory})
	for obj := range objs {
		err := s3.RemoveFile(obj.Key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s3 *S3) ListDirectory(directory string) ([]string, error) {
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
