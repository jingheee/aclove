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

	c.JSON(http.StatusCreated, models.CategoryToResponse(category))
}

func (h *CategoryHandler) Get(c *gin.Context) (*models.CategoryResponse, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	category, err := h.svc.GetCategory(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	return models.CategoryToResponse(category), nil
}

func (h *CategoryHandler) Update(c *gin.Context) (*models.CategoryResponse, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	if req.ParentID != nil {
		if err := h.svc.ValidateParentID(c.Request.Context(), req.ParentID); err != nil {
			return nil, err
		}
	}

	category, err := h.svc.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		return nil, err
	}

	return models.CategoryToResponse(category), nil
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
		response[i] = models.CategoryToResponse(category)
	}

	c.JSON(http.StatusOK, response)
}

func (h *CategoryHandler) GetTree(c *gin.Context) {
	tree, err := h.svc.GetCategoryTree(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分类树失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tree)
}
