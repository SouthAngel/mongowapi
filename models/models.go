package models

import "go.mongodb.org/mongo-driver/bson"

// APIResponse 统一响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// DocumentParams 文档操作的基础参数（URL 路径中提取）
type DocumentParams struct {
	Database string
	Collection string
}

// InsertOneRequest 插入单文档请求体
type InsertOneRequest struct {
	Document bson.M `json:"document" binding:"required"`
}

// InsertManyRequest 批量插入请求体
type InsertManyRequest struct {
	Documents []bson.M `json:"documents" binding:"required"`
}

// FindRequest 查询请求体
type FindRequest struct {
	Filter     bson.M `json:"filter"`               // 过滤条件，可选，默认为空
	Sort       bson.D `json:"sort"`                 // 排序，可选
	Projection bson.M `json:"projection"`           // 投影字段，可选
	Skip       int64  `json:"skip"`                 // 跳过的文档数
	Limit      int64  `json:"limit"`                // 返回的最大条数，默认 10
}

// UpdateRequest 更新请求体
type UpdateRequest struct {
	Filter bson.M     `json:"filter" binding:"required"` // 过滤条件
	Update bson.M     `json:"update" binding:"required"` // 更新操作（如 $set）
	Upsert bool       `json:"upsert"`                    // 不存在时是否插入
}

// UpdateResult 更新结果
type UpdateResult struct {
	MatchedCount  int64 `json:"matched_count"`
	ModifiedCount int64 `json:"modified_count"`
	UpsertedID    interface{} `json:"upserted_id,omitempty"`
}

// IndexCreateRequest 创建索引请求体
type IndexCreateRequest struct {
	Keys bson.D `json:"keys" binding:"required"` // 索引字段及排序，如 [{name:1}, {age:-1}]
}