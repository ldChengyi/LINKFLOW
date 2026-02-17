package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type ModuleHandler struct {
	repo *repository.ModuleRepository
}

func NewModuleHandler(repo *repository.ModuleRepository) *ModuleHandler {
	return &ModuleHandler{repo: repo}
}

// List 获取所有可用模块
func (h *ModuleHandler) List(c *gin.Context) {
	modules, err := h.repo.List(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to list modules")
		return
	}
	Success(c, modules)
}

// Get 获取模块详情
func (h *ModuleHandler) Get(c *gin.Context) {
	id := c.Param("id")
	module, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		Fail(c, http.StatusNotFound, "module not found")
		return
	}
	Success(c, module)
}
