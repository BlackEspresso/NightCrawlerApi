package main

import (
	"log"
	"mime"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

type Settings struct {
	PhantomPath  string
	MailSettings map[string]string
	S3Buckets    map[string]string
}

var appSettings Settings = Settings{}

func main() {

	appSettings.PhantomPath = "./"
	appSettings.S3Buckets = map[string]string{
		"Screenshots": "nightcrawlerlinks",
	}

	r := gin.Default()

	r.GET("crawl/siteinfo", siteinfo)
	r.GET("crawld/screenshot", screenshot)
	r.GET("crawld/siteinfo", siteinfodyn)
	r.GET("crawld/bucketinfo", bucketinfo)
	r.GET("crawld/pageload", siteinfodyn)
	r.GET("crawl/task/add", crawltask)
	r.GET("crawl/task/info", crawltask)
	r.GET("crawl/task/stop", crawltask)
	r.GET("crawl/task/delete", crawltask)
	r.GET("/tasks", tasks)
	r.Run(":8076")
}

func bucketinfo(g *gin.Context) {
	b := GetBucketUrl("nightcrawlerlinks")
	g.String(200, b)
}

func getMimeType(fileName string) string {
	ext := path.Ext(fileName)
	if ext == ".jpeg" {
		ext = ".jpg"
	}
	cType := mime.TypeByExtension(ext)
	if cType == "" {
		cType = "binary/octet-stream"
	}
	return cType
}

func GetBucketUrl(bucketName string) string {
	return "https://s3.amazonaws.com/" + bucketName + "/"
}

func uploadToS3(fileName string, key string, meta map[string]*string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	svc := s3.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

	if err != nil {
		log.Println("GetBucketLocation", err)
	}

	bucketName := appSettings.S3Buckets["Screenshots"]
	bucketUrl := GetBucketUrl(bucketName)

	cType := getMimeType(fileName)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Body:        f,
		Bucket:      &bucketName,
		Key:         &key,
		ContentType: &cType,
		Metadata:    meta,
	})

	if err != nil {
		log.Println(err)
	}
	return bucketUrl + key, nil
}

func siteinfo(g *gin.Context) {

}

func siteinfodyn(g *gin.Context) {

}

func screenshot(g *gin.Context) {
	queryUrl := g.Query("url")
	format := g.Query("format")
	if format == "" {
		format = "jpeg"
	}
	if queryUrl == "" {
		g.String(403, "needs url parameter")
		return
	}

	url, err := url.Parse(queryUrl)
	if err != nil {
		g.String(403, "invalid url")
		return
	}
	if !url.IsAbs() {
		g.String(403, "url not absolute")
		return
	}
	if url.Host == "localhost" || url.Host == "127.0.0.1" || url.Host == ":::1" {
		g.String(403, "cant crawl localhost")
		return
	}

	fileUUID := uuid.NewV4()
	out, err := runPhantom("screen-capture.js", queryUrl, fileUUID.String(), format)
	if err != nil {
		log.Println(err)
		g.String(500, err.Error()+","+string(out))
		return
	}

	fname := fileUUID.String() + "." + format
	meta := map[string]*string{
		"URL": &queryUrl,
	}
	downloadUrl, _ := uploadToS3("./"+fileUUID.String(), fname, meta)
	os.Remove(fileUUID.String())
	g.String(200, downloadUrl)
}

func runPhantom(args ...string) ([]byte, error) {
	out, err := exec.Command(appSettings.PhantomPath+"phantomjs", args...).Output()
	return out, err
}

func crawltask(g *gin.Context) {

}

func tasks(g *gin.Context) {

}
