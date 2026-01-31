package handlers

import (
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
		models.JSONBadRequest(c, "无效的请求数据")
		return
	}

	if err := h.svc.ValidateParentID(c.Request.Context(), req.ParentID); err != nil {
		models.JSONBadRequest(c, err.Error())
		return
	}

	category, err := h.svc.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		models.JSONInternalError(c, "创建分类失败")
		return
	}

	models.JSONCreated(c, models.CategoryToResponse(category))
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
		models.JSONBadRequest(c, "无效的分类ID")
		return
	}

	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		models.JSONInternalError(c, "删除分类失败")
		return
	}

	models.JSONSuccessWithMsg(c, "删除成功", nil)
}


func (h *CategoryHandler) GetChildren(c *gin.Context) {
	var parentID *int64
	if c.Param("id") != "" {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			models.JSONBadRequest(c, "无效的分类ID")
			return
		}
		parentID = &id
	}

	categories, err := h.svc.GetChildren(c.Request.Context(), parentID)
	if err != nil {
		models.JSONInternalError(c, "查询子分类失败")
		return
	}

	response := make([]*models.CategoryResponse, len(categories))
	for i, category := range categories {
		response[i] = models.CategoryToResponse(category)
	}

	models.JSONSuccess(c, response)
}

func (h *CategoryHandler) GetTree(c *gin.Context) {
	tree, err := h.svc.GetCategoryTree(c.Request.Context())
	if err != nil {
		models.JSONInternalError(c, "获取分类树失败")
		return
	}

	models.JSONSuccess(c, tree)
}
