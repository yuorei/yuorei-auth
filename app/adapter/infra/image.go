package infra

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/kolesa-team/go-webp/webp"
)

func ConvertToWebp(imageFile multipart.File, id string) (*os.File, error) {
	if imageFile == nil {
		return nil, nil
	}

	buf := make([]byte, 512)
	_, err := imageFile.Read(buf)
	if err != nil {
		return nil, err
	}

	// コンテンツタイプを検出する
	contentType := http.DetectContentType(buf)
	fmt.Println("Content Type:", contentType)

	var image image.Image
	switch contentType {
	case "image/jpeg":
		// JPEG画像をデコード
		img, err := jpeg.Decode(imageFile)
		if err != nil {
			return nil, fmt.Errorf("Failed to decode JPEG image")
		}
		image = img
	case "image/png":
		// PNG画像をデコード
		img, err := png.Decode(imageFile)
		if err != nil {
			return nil, fmt.Errorf("Failed to decode PNG image")
		}
		image = img
	case "image/webp":
		// WEBP画像をデコード
		img, err := webp.Decode(imageFile, nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to decode WEBP image")
		}
		image = img
	default:
		return nil, fmt.Errorf("This file is not supported.")
	}

	imageTmp, err := os.Create(id + ".webp")
	if err != nil {
		return nil, err
	}
	defer imageTmp.Close()
	err = webp.Encode(imageTmp, image, nil)
	if err != nil {

		return nil, fmt.Errorf("Failed to encode to WEBP image")
	}

	return imageTmp, nil
}

func UploadImageForStorage(imageBuffer *os.File, id string) (string, error) {
	ctx := context.Background()

	imagePath := id + ".webp"
	defer os.Remove(imagePath)
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cred := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(cred))
	if err != nil {
		return "", err
	}

	// change object address style
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = true
		options.BaseEndpoint = aws.String(os.Getenv("AWS_S3_ENDPOINT"))
		options.Region = "ap-northeast-1"
	})

	// get buckets
	lbo, err := client.ListBuckets(ctx, nil)
	if err != nil {
		return "", err
	}
	buckets := make(map[string]struct{}, len(lbo.Buckets))
	for _, b := range lbo.Buckets {
		buckets[*b.Name] = struct{}{}
	}

	// create 'video-service' bucket if not exist
	bucketName := "keycloal-user"
	if _, ok := buckets[bucketName]; !ok {
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &bucketName,
			ACL:    types.BucketCannedACLPublicRead,
		})
		if err != nil {
			return "", err
		}
	}
	fmt.Println("bucketName: ", bucketName, imagePath)

	image, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer image.Close()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(imagePath),
		Body:   image,
		ACL:    types.ObjectCannedACLPublicRead,
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	log.Println("Successful upload: ", id)

	url := fmt.Sprintf("%s/video-service/%s.webp", os.Getenv("AWS_S3_ENDPOINT"), id)
	return url, nil
}
