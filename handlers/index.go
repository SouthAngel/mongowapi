package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"

	"mongowapi/models"
)

// CreateIndex 创建单个索引
// POST /api/:database/:collection/indexes
func (h *Handler) CreateIndex(c *gin.Context) {
	var req models.IndexCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	p := h.params(c)
	ctx, cancel := h.ctx(c)
	defer cancel()

	name, err := h.mongo.Collection(p.Database, p.Collection).Indexes().CreateOne(ctx, mongo.IndexModel{Keys: req.Keys})
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "创建索引失败: "+err.Error())
		return
	}
	ok(c, gin.H{"index_name": name})
}

// ListIndexes 列出集合的所有索引
// GET /api/:database/:collection/indexes
func (h *Handler) ListIndexes(c *gin.Context) {
	p := h.params(c)
	ctx, cancel := h.ctx(c)
	defer cancel()

	cur, err := h.mongo.Collection(p.Database, p.Collection).Indexes().List(ctx)
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
// DELETE /api/:database/:collection/indexes/:name
func (h *Handler) DropIndex(c *gin.Context) {
	p := h.params(c)
	name := c.Param("name")
	ctx, cancel := h.ctx(c)
	defer cancel()

	if _, err := h.mongo.Collection(p.Database, p.Collection).Indexes().DropOne(ctx, name); err != nil {
		fail(c, http.StatusInternalServerError, 500, "删除索引失败: "+err.Error())
		return
	}
	ok(c, gin.H{"dropped_index": name})
}