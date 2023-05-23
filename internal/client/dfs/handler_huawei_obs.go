package dfs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type HuaweiObsConfig struct {
	Url    string `yaml:"url"`
	Bucket string `yaml:"bucket"`
	KeyID  string `yaml:"key_id"`
	Key    string `yaml:"key"`
}

type huaweiObsHandler struct {
	bucket          string
	expireDays      TimeDays
	config          *HuaweiObsConfig
	huaweiObsClient *obs.ObsClient
	once            *sync.Once
}

func newHuaweiObsHandler(bucket string, expireDays TimeDays, config *HuaweiObsConfig) (DFSHandler, error) {
	huaweiObsClient, err := obs.New(config.KeyID, config.Key, config.Url)
	if err != nil {
		return nil, err
	}

	if bucket == "" {
		bucket = config.Bucket
	}

	handler := &huaweiObsHandler{bucket: bucket, expireDays: expireDays, config: config, huaweiObsClient: huaweiObsClient, once: new(sync.Once)}
	handler.huaweiObsClient = huaweiObsClient
	return handler, nil
}

func (h *huaweiObsHandler) TryMakeBucket() error {
	var err error
	makeFun := func() {

		bucketInput := &obs.CreateBucketInput{}
		bucketInput.Bucket = h.config.Bucket
		bucketInput.ACL = obs.AclPrivate
		bucketInput.StorageClass = obs.StorageClassStandard

		_, err := h.huaweiObsClient.CreateBucket(bucketInput)
		if err != nil {
			// Check to see if we already own this bucket (which happens if you run this twice)

			_, err := h.huaweiObsClient.HeadBucket(h.bucket)
			if err == nil {
				return
			}

			if obsError, ok := err.(obs.ObsError); ok {
				err = fmt.Errorf("obs errcode:%v errmessage:%v", obsError.Code, obsError.Message)
			}

			return
		} else {

			if h.expireDays > 0 {

				lifeConfig := &obs.SetBucketLifecycleConfigurationInput{
					Bucket: h.config.Bucket,
					BucketLifecyleConfiguration: obs.BucketLifecyleConfiguration{
						LifecycleRules: make([]obs.LifecycleRule, 0),
					},
				}
				lifeConfig.LifecycleRules = append(lifeConfig.LifecycleRules, obs.LifecycleRule{
					Status: obs.RuleStatusEnabled,
					ID:     "rule1",
					Expiration: obs.Expiration{
						Days: h.expireDays,
					},
				})

				_, err = h.huaweiObsClient.SetBucketLifecycleConfiguration(lifeConfig)
				if err != nil {
					return
				}
			}
		}
		return
	}

	h.once.Do(makeFun)

	return err
}

func (h *huaweiObsHandler) PutObject(path, fileName string, payload []byte) error {
	// ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	// defer cf()
	contentType := "text/plain"

	if path[len(path)-1] != '/' {
		path += "/"
	}

	objectInput := &obs.PutObjectInput{}
	objectInput.Bucket = h.bucket
	objectInput.Key = path + fileName
	objectInput.Body = bytes.NewReader(payload)
	objectInput.ContentType = contentType

	_, err := h.huaweiObsClient.PutObject(objectInput)

	// info, err := h.huaweiObsClient.PutObject(ctx, h.bucket, path+fileName, bytes.NewReader(payload), int64(len(payload)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	return nil
}

func (h *huaweiObsHandler) GetObject(path, fileName string) ([]byte, error) {
	// ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	// defer cf()
	if path[len(path)-1] != '/' {
		path += "/"
	}
	objectInput := &obs.GetObjectInput{}
	objectInput.Bucket = h.bucket
	objectInput.Key = path + fileName

	obj, err := h.huaweiObsClient.GetObject(objectInput)

	// obj, err := h.huaweiObsClient.GetObject(ctx, h.bucket, path+fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	defer obj.Body.Close()

	payload, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (h *huaweiObsHandler) DelObject(path, fileName string) error {
	if path[len(path)-1] != '/' {
		path += "/"
	}

	input := &obs.DeleteObjectInput{
		Bucket: h.bucket,
		Key:    path + fileName,
	}
	input.Bucket = h.bucket
	input.Key = path + fileName
	_, err := h.huaweiObsClient.DeleteObject(input)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			errStr := fmt.Sprintf("code:%v message:%v", obsError.Code, obsError.Message)
			return fmt.Errorf(errStr)
		}
		return err
	}
	return nil
}
