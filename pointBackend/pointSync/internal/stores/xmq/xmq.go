package xmq

import (
	"context"
	"fmt"
	"log"
	"pointSync/internal/config"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
)

const (
	TopicBalanceEvents = "balance-events"
	PartitionKeyUser   = "user_id"
)

type Producer struct {
	KqPusherClient *kq.Pusher
}

func New(cfg *config.Config) *Producer {
	// 3. 创建 Kafka 生产者客户端
	newQueue(cfg.KqPusherConf)
	client := kq.NewPusher(cfg.KqPusherConf.Brokers, cfg.KqPusherConf.Topic)
	return &Producer{
		KqPusherClient: client,
	}
}

// NewProducer 创建Kafka生产者
func newQueue(cfg *config.KqConf) error {

	config := sarama.NewConfig()

	admin, err := sarama.NewClusterAdmin(cfg.Brokers, config)
	if err != nil {
		log.Fatalf("创建 ClusterAdmin 失败: %v", err)
		return err
	}
	defer admin.Close()

	// 1. 检查 topic 是否存在
	if isTopicExists(admin, cfg.Topic) {
		fmt.Printf("Topic '%s' 已存在\n", cfg.Topic)
		return nil
	}

	// 2. 配置 topic 参数
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     cfg.Partitions,        // 指定分区数为 6
		ReplicationFactor: cfg.ReplicationFactor, // 指定副本数为 1（生产环境建议 3）
		ConfigEntries:     make(map[string]*string),
	}
	// 3. 可选：添加 topic 级别配置
	//retentionMs := "604800000" // 消息保留 7 天
	topicDetail.ConfigEntries["retention.ms"] = &cfg.RetentionMs
	//4. 创建 topic
	err = admin.CreateTopic(cfg.Topic, topicDetail, false)
	if err != nil {
		log.Fatalf("创建 topic 失败: %v", err)
		return err
	}
	fmt.Printf("Topic '%s' 创建成功，分区数: %d\n", cfg.Topic, topicDetail.NumPartitions)
	return nil
}

func isTopicExists(admin sarama.ClusterAdmin, topic string) bool {
	topics, err := admin.ListTopics()
	if err != nil {
		return false
	}
	_, exists := topics[topic]
	return exists
}

func Consumers(c *config.Config, ctx context.Context) []service.Service {

	return []service.Service{
		//Listening for changes in consumption flow status
		kq.MustNewQueue(c.KqConsumerConf, NewPointConsumer(ctx)),
		//.....
	}
}
