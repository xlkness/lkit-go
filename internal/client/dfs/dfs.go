package dfs

import (
	"fmt"
)

type TimeDays = int

type DFSHandler interface {
	TryMakeBucket() error
	PutObject(path, fileName string, payload []byte) error
	GetObject(path, fileName string) ([]byte, error)
	DelObject(path, fileName string) error
}

type DFSType = string

var (
	DFSType_Minio       DFSType = "minio"
	DFSType_S3          DFSType = "s3"
	DFSType_CloudStorge DFSType = "cloudStorage"
	DFSType_HUAWEI_Obs  DFSType = "huawei-obs"
)

type Config struct {
	Type         DFSType             `yaml:"type"` // s3/minio/cloud_storage
	Enable       bool                `yaml:"enable"`
	Bucket       string              `yaml:"bucket"`
	ExpireDays   TimeDays            `yaml:"expire_days"` // 桶数据过期时间
	S3           *S3Config           `yaml:"s3"`
	Minio        *MinioConfig        `yaml:"minio"`
	CloudStorage *CloudStorageConfig `yaml:"cloud_storage"`
	Huaweicloud  *HuaweiObsConfig    `yaml:"huawei_obs"`

	// deprecated:兼容旧版本配置
	Url string `yaml:"url"`
	// deprecated:兼容旧版本配置
	KeyID string `yaml:"key_id"`
	// deprecated:兼容旧版本配置
	Key string `yaml:"key"`
}

func NewHandler(config *Config) (DFSHandler, error) {
	if !config.Enable {
		return nil, fmt.Errorf("config set diable dfs but now invoke new")
	}
	switch config.Type {
	case DFSType_Minio:
		if config.Bucket == "" && config.Minio != nil && config.Minio.Bucket == "" {
			return nil, fmt.Errorf("bucket is nil")
		}
		if config.Minio != nil {
			return newMinioHandler(config.Bucket, config.ExpireDays, config.Minio)
		} else if config.Url != "" {
			return newMinioHandler(config.Bucket, config.ExpireDays, &MinioConfig{
				Url:    config.Url,
				KeyID:  config.KeyID,
				Key:    config.Key,
				Bucket: config.Bucket,
			})
		} else {
			return nil, fmt.Errorf("not found invalid config for minio")
		}
	case DFSType_S3:
		if config.Bucket == "" && config.S3 != nil && config.S3.Bucket == "" {
			return nil, fmt.Errorf("bucket is nil")
		}
		if config.S3 != nil {
			return newS3Handler(config.Bucket, config.ExpireDays, config.S3)
		} else if config.Url != "" {
			return newS3Handler(config.Bucket, config.ExpireDays, &S3Config{
				Url:    config.Url,
				KeyID:  config.KeyID,
				Key:    config.Key,
				Bucket: config.Bucket,
			})
		} else {
			return nil, fmt.Errorf("not found invalid config for s3")
		}
	case DFSType_CloudStorge:
		if config.Bucket == "" {
			return nil, fmt.Errorf("bucket is nil")
		}
		if config.CloudStorage != nil {
			return newCloudStorageHandler(config.Bucket, config.ExpireDays, config.CloudStorage)
		}
		return nil, fmt.Errorf("not found invalid config for cloud storage")
	case DFSType_HUAWEI_Obs:
		if config.Bucket == "" && config.Huaweicloud != nil && config.Huaweicloud.Bucket == "" {
			return nil, fmt.Errorf("bucket is nil")
		}
		if config.Huaweicloud != nil {
			return newHuaweiObsHandler(config.Bucket, config.ExpireDays, config.Huaweicloud)
		} else if config.Url != "" {
			return newHuaweiObsHandler(config.Bucket, config.ExpireDays, &HuaweiObsConfig{
				Url:    config.Url,
				KeyID:  config.KeyID,
				Key:    config.Key,
				Bucket: config.Bucket,
			})
		} else {
			return nil, fmt.Errorf("not found invalid config for huawei obs")
		}
	}
	return nil, fmt.Errorf("invalid dfs type:%v", config.Type)
}
