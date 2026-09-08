package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mongowapi/database"
	"mongowapi/models"
)

// Handler 所有 HTTP 处理器的载体，持有 MongoDB 连接、请求超时与数据库白名单
type Handler struct {
	mongo     *database.Mongo
	timeout   time.Duration
	whiteList map[string]struct{} // 允许访问的数据库白名单，为空表示不限制
}

// NewHandler 创建处理器，whiteList 为允许访问的数据库列表（为空表示不限制）
func NewHandler(m *database.Mongo, timeout time.Duration, whiteList []string) *Handler {
	wl := make(map[string]struct{}, len(whiteList))
	for _, db := range whiteList {
		if db != "" {
			wl[db] = struct{}{}
		}
	}
	return &Handler{mongo: m, timeout: timeout, whiteList: wl}
}

// checkDB 校验数据库是否在白名单内；未配置白名单时始终放行
func (h *Handler) checkDB(c *gin.Context, db string) bool {
	if len(h.whiteList) == 0 {
		return true
	}
	if _, ok := h.whiteList[db]; ok {
		return true
	}
	fail(c, http.StatusForbidden, 403, "数据库不在白名单内，禁止访问: "+db)
	return false
}

// ctx 返回带超时的上下文
func (h *Handler) ctx(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), h.timeout)
}

// InsertOne 插入单条文档
// POST /api/insert
func (h *Handler) InsertOne(c *gin.Context) {
	var req models.InsertOneRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	res, err := h.mongo.Collection(req.Database, req.Collection).InsertOne(ctx, req.Document)
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "插入文档失败: "+err.Error())
		return
	}
	ok(c, gin.H{"inserted_id": res.InsertedID})
}

// InsertMany 批量插入文档
// POST /api/insertmany
func (h *Handler) InsertMany(c *gin.Context) {
	var req models.InsertManyRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	docs := make([]interface{}, len(req.Documents))
	for i := range req.Documents {
		docs[i] = req.Documents[i]
	}
	res, err := h.mongo.Collection(req.Database, req.Collection).InsertMany(ctx, docs)
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "批量插入失败: "+err.Error())
		return
	}
	ok(c, gin.H{"inserted_ids": res.InsertedIDs})
}

// FindOne 查询单条文档
// POST /api/findone
func (h *Handler) FindOne(c *gin.Context) {
	var req models.FindRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	opts := options.FindOne().SetProjection(req.Projection)
	var result bson.M
	err := h.mongo.Collection(req.Database, req.Collection).FindOne(ctx, req.Filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fail(c, http.StatusNotFound, 404, "未找到文档")
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "查询失败: "+err.Error())
		return
	}
	ok(c, result)
}

// Find 查询文档列表（支持分页、排序、投影）
// POST /api/find
func (h *Handler) Find(c *gin.Context) {
	var req models.FindRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 10
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	opts := options.Find().
		SetSort(req.Sort).
		SetProjection(req.Projection).
		SetSkip(req.Skip).
		SetLimit(req.Limit)

	cur, err := h.mongo.Collection(req.Database, req.Collection).Find(ctx, req.Filter, opts)
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "查询失败: "+err.Error())
		return
	}
	defer cur.Close(ctx)

	var results []bson.M
	if err := cur.All(ctx, &results); err != nil {
		fail(c, http.StatusInternalServerError, 500, "解析结果失败: "+err.Error())
		return
	}
	ok(c, results)
}

// UpdateOne 更新单条文档
// POST /api/update
func (h *Handler) UpdateOne(c *gin.Context) {
	var req models.UpdateRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	opts := options.Update().SetUpsert(req.Upsert)
	res, err := h.mongo.Collection(req.Database, req.Collection).UpdateOne(ctx, req.Filter, req.Update, opts)
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "更新失败: "+err.Error())
		return
	}
	ok(c, models.UpdateResult{
		MatchedCount:  res.MatchedCount,
		ModifiedCount: res.ModifiedCount,
		UpsertedID:    res.UpsertedID,
	})
}

// UpdateMany 更新多条文档
// POST /api/updatemany
func (h *Handler) UpdateMany(c *gin.Context) {
	var req models.UpdateRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	opts := options.Update().SetUpsert(req.Upsert)
	res, err := h.mongo.Collection(req.Database, req.Collection).UpdateMany(ctx, req.Filter, req.Update, opts)
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "批量更新失败: "+err.Error())
		return
	}
	ok(c, models.UpdateResult{
		MatchedCount:  res.MatchedCount,
		ModifiedCount: res.ModifiedCount,
	})
}

// DeleteOne 删除单条文档
// POST /api/delete
func (h *Handler) DeleteOne(c *gin.Context) {
	var req models.DeleteRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	res, err := h.mongo.Collection(req.Database, req.Collection).DeleteOne(ctx, req.Filter)
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "删除失败: "+err.Error())
		return
	}
	ok(c, gin.H{"deleted_count": res.DeletedCount})
}

// DeleteMany 删除多条文档
// POST /api/deletemany
func (h *Handler) DeleteMany(c *gin.Context) {
	var req models.DeleteRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	res, err := h.mongo.Collection(req.Database, req.Collection).DeleteMany(ctx, req.Filter)
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "批量删除失败: "+err.Error())
		return
	}
	ok(c, gin.H{"deleted_count": res.DeletedCount})
}