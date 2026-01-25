package handlers

import (
	"net/http"
	"strconv"

	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/service"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
		return
	}

	if err := h.svc.ValidateParentID(c.Request.Context(), req.ParentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.svc.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分类失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category.ToResponse())
}

func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类ID"})
		return
	}

	category, err := h.svc.GetCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "分类不存在"})
		return
	}

	c.JSON(http.StatusOK, category.ToResponse())
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类ID"})
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
		return
	}

	if req.ParentID != nil {
		if err := h.svc.ValidateParentID(c.Request.Context(), req.ParentID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	category, err := h.svc.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新分类失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category.ToResponse())
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类ID"})
		return
	}

	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除分类失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *CategoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	categories, total, err := h.svc.ListCategories(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询分类列表失败", "details": err.Error()})
		return
	}

	response := make([]*models.CategoryResponse, len(categories))
	for i, category := range categories {
		response[i] = category.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *CategoryHandler) GetChildren(c *gin.Context) {
	var parentID *int64
	if c.Param("id") != "" {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类ID"})
			return
		}
		parentID = &id
	}

	categories, err := h.svc.GetChildren(c.Request.Context(), parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询子分类失败", "details": err.Error()})
		return
	}

	response := make([]*models.CategoryResponse, len(categories))
	for i, category := range categories {
		response[i] = category.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}
