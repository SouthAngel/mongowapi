package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo MongoDB 连接管理器
type Mongo struct {
	Client *mongo.Client
}

// New 创建 MongoDB 连接并测试连通性
func New(ctx context.Context, uri string, connectTimeout time.Duration) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("连接 MongoDB 失败: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("Ping MongoDB 失败: %w", err)
	}

	return &Mongo{Client: client}, nil
}

// Collection 获取指定集合
func (m *Mongo) Collection(dbName, collName string) *mongo.Collection {
	return m.Client.Database(dbName).Collection(collName)
}

// Close 关闭 MongoDB 连接
func (m *Mongo) Close(ctx context.Context) error {
	if m.Client != nil {
		return m.Client.Disconnect(ctx)
	}
	return nil
}

// ListDatabases 列出所有数据库名称
func (m *Mongo) ListDatabases(ctx context.Context) ([]string, error) {
	res, err := m.Client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	return res, nil
}