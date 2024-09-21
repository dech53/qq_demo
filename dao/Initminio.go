package dao

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
)

var (
	MinioClient *minio.Client
)
// 初使化 minio
func Initminio() {
	endpoint := "localhost:9000"
	accessKeyID := "h0tTGdv6MkYpm9E0Hwgg"
	secretAccessKey := "lsIIYs0kDkldB2WDjQsXGiHPG6YQSaXwY9U5bjBp"
	useSSL := false
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln("MinIO 初始化失败:", err)
	}
	MinioClient = minioClient
	log.Println("MinIO 初始化成功:", minioClient) // MinIO 初始化成功
	// err = minioClient.MakeBucket(context.Background(), "dech53", minio.MakeBucketOptions{
	// 	Region: "cn-north-1",
	// })
	// if err != nil {
	// 	log.Println("创建Bucket失败:", err)
	// } else {
	// 	log.Println("Bucket创建成功")
	// }
}
