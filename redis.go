package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/baagee/test-vgo/util"
	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
	"time"
)

const (
	// 自增长key
	URL_REDIS_KEY = "shorten:next:url:id"
	//短地址和长地址的映射
	ShortlinkKey = "shortlink:%s:url"
	// 长地址对应的hash 和短地址关联
	URLHashKey = "urlhash:%s:url"
	// 通过短地址获取对应详细信息
	ShortlinkDetailKey = "shortlink:%s:detail"
)

type RedisCli struct {
	Client *redis.Client
}

//url详细信息结构体
type UrlDetail struct {
	Url        string        `json:"url"`
	CreateTime string        `json:"create_time"`
	Expire     time.Duration `json:"expire"`
}

// 创建redis客户端
func NewRedisCli(address string, password string, db int) *RedisCli {
	client := redis.NewClient(&redis.Options{
		Addr:       address,
		Password:   password,
		DB:         db,
		PoolSize:   1000,
		MaxRetries: 3,
	})
	if _, err := client.Ping().Result(); err != nil {
		panic("connect redis error")
	}
	return &RedisCli{Client: client}
}

//expire 分钟
func (r *RedisCli) Shorten(url string, expire int64) (string, error) {
	var err error
	//获取url的hash
	h := toSHa1(url)
	//验证这个url是否处理过
	res1, err := r.Client.Get(fmt.Sprintf(URLHashKey, h)).Result()
	if err == redis.Nil {
		//为空
	} else if err != nil {
		return "", err
	} else {
		if res1 == "{}" {
			//过期了 nothing
		} else {
			//存在 返回url
			return res1, nil
		}
	}

	err = r.Client.Incr(URL_REDIS_KEY).Err()
	if err != nil {
		return "", err
	} else {
		id, err := r.Client.Get(URL_REDIS_KEY).Int64()
		if err != nil {
			return "", err
		} else {
			// 将ID转化为字符串
			eid := base62.EncodeInt64(id)
			//将短地址和url做映射
			err = r.Client.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Minute*time.Duration(expire)).Err()
			if err != nil {
				return "", err
			} else {
				//将长地址的hash和短地址做映射
				// int64转化为 时间 time.Duration(expire) 进行类型转化
				err = r.Client.Set(fmt.Sprintf(URLHashKey, h), eid, time.Minute*time.Duration(expire)).Err()
				if err != nil {
					return "", err
				} else {
					// 获取详情
					urlDetail := UrlDetail{
						Url:        url,
						CreateTime: time.Now().String(),
						Expire:     time.Duration(expire),
					}
					fmt.Println(urlDetail)
					detail, err := json.Marshal(urlDetail)
					if err != nil {
						return "", err
					} else {
						//将短地址详情保存
						err = r.Client.Set(fmt.Sprintf(ShortlinkDetailKey, eid), string(detail), time.Minute*time.Duration(expire)).Err()
						if err != nil {
							return "", err
						} else {
							return eid, nil
						}
					}
				}
			}
		}
	}
}

func toSHa1(str string) string {
	sha := sha1.New()
	return string(sha.Sum([]byte(str)))
}

func (r *RedisCli) ShortenInfo(shorten string) (UrlDetail, error) {
	result, err := r.Client.Get(fmt.Sprintf(ShortlinkDetailKey, shorten)).Result()
	if err == redis.Nil {
		//不为空
		return UrlDetail{}, util.StatusError{
			Code: 404,
			Err:  errors.New("unknow shorten"),
		}
	} else if err != nil {
		return UrlDetail{}, err
	} else {
		byt := []byte(result)
		var urlDetail UrlDetail
		json.Unmarshal(byt, &urlDetail)
		return urlDetail, nil
	}
}

func (r *RedisCli) UnShorten(shorten string) (string, error) {
	//通过短地址找到对应的原来的地址
	result, err := r.Client.Get(fmt.Sprintf(ShortlinkKey, shorten)).Result()
	if err == redis.Nil {
		//没找到
		return "", util.StatusError{
			Code: 404,
			Err:  errors.New("unknow shorten"),
		}
	} else if err != nil {
		return "", err
	} else {
		return result, nil
	}
}
