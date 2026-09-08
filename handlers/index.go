package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"

	"mongowapi/models"
)

// CreateIndex 创建单个索引
// POST /api/createindex
func (h *Handler) CreateIndex(c *gin.Context) {
	var req models.IndexCreateRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	name, err := h.mongo.Collection(req.Database, req.Collection).Indexes().CreateOne(ctx, mongo.IndexModel{Keys: req.Keys})
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "创建索引失败: "+err.Error())
		return
	}
	ok(c, gin.H{"index_name": name})
}

// ListIndexes 列出集合的所有索引
// POST /api/indexes
func (h *Handler) ListIndexes(c *gin.Context) {
	var req models.IndexListRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	cur, err := h.mongo.Collection(req.Database, req.Collection).Indexes().List(ctx)
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "列出索引失败: "+err.Error())
		return
	}
	defer cur.Close(ctx)

	var indexes []map[string]interface{}
	if err := cur.All(ctx, &indexes); err != nil {
		fail(c, http.StatusInternalServerError, 500, "解析索引失败: "+err.Error())
		return
	}
	ok(c, indexes)
}

// DropIndex 删除指定索引
// POST /api/dropindex
func (h *Handler) DropIndex(c *gin.Context) {
	var req models.IndexDropRequest
	if !bindJSON(c, &req) || !h.checkDB(c, req.Database) {
		return
	}
	ctx, cancel := h.ctx(c)
	defer cancel()

	if _, err := h.mongo.Collection(req.Database, req.Collection).Indexes().DropOne(ctx, req.Name); err != nil {
		fail(c, http.StatusInternalServerError, 500, "删除索引失败: "+err.Error())
		return
	}
	ok(c, gin.H{"dropped_index": req.Name})
}