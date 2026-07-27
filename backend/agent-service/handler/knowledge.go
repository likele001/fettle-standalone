package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ai-platform/agent-service/models"
	"ai-platform/agent-service/service"
	"ai-platform/agent-service/storage"
	"ai-platform/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const uploadDir = "data/uploads/documents"

// KnowledgeHandler 知识库 API 处理器
type KnowledgeHandler struct {
	kbService   *service.KnowledgeService
	minioClient *storage.MinioClient
}

func NewKnowledgeHandler(kbService *service.KnowledgeService, minioClient *storage.MinioClient) *KnowledgeHandler {
	return &KnowledgeHandler{kbService: kbService, minioClient: minioClient}
}

// CreateKB 创建知识库
func (h *KnowledgeHandler) CreateKB(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	var req service.CreateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}

	kb, err := h.kbService.Create(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": kb})
}

// GetKB 获取知识库详情
func (h *KnowledgeHandler) GetKB(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	kbID := c.Param("id")

	kb, err := h.kbService.GetByID(tenantID, kbID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1004, "message": "knowledge base not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": kb})
}

// ListKB 获取知识库列表
func (h *KnowledgeHandler) ListKB(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	kbs, total, err := h.kbService.List(tenantID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     kbs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// UpdateKB 更新知识库
func (h *KnowledgeHandler) UpdateKB(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	kbID := c.Param("id")

	var req service.UpdateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": err.Error()})
		return
	}

	kb, err := h.kbService.Update(tenantID, kbID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": kb})
}

// DeleteKB 删除知识库
func (h *KnowledgeHandler) DeleteKB(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	kbID := c.Param("id")

	if err := h.kbService.Delete(tenantID, kbID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// UploadDocument 上传文档到知识库
func (h *KnowledgeHandler) UploadDocument(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1002, "message": "unauthorized"})
		return
	}
	kbID := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "file is required"})
		return
	}

	fileType := ""
	switch {
	case len(file.Filename) >= 4 && file.Filename[len(file.Filename)-4:] == ".pdf":
		fileType = "pdf"
	case len(file.Filename) >= 5 && file.Filename[len(file.Filename)-5:] == ".docx":
		fileType = "docx"
	case len(file.Filename) >= 4 && file.Filename[len(file.Filename)-4:] == ".txt":
		fileType = "txt"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "unsupported file type"})
		return
	}

	doc := &models.KnowledgeDocument{
		KnowledgeBaseID: uuid.MustParse(kbID),
		FileName:        file.Filename,
		FileType:        fileType,
		FileSize:        file.Size,
		Status:          "pending",
	}

	if parsed, err := uuid.Parse(tenantID); err == nil {
		doc.TenantID = parsed
	}

	if err := h.kbService.AddDocument(tenantID, kbID, doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	// 保存文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "open file failed"})
		return
	}
	defer src.Close()

	objectName := fmt.Sprintf("%s/%s/%s_%d%s", tenantID, kbID, doc.ID.String(), time.Now().UnixNano(), filepath.Ext(file.Filename))

	if h.minioClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fileURL, err := h.minioClient.Upload(ctx, objectName, src, file.Size, "application/octet-stream")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "upload to storage failed"})
			return
		}
		doc.FileURL = fileURL
	} else {
		tenantDir := filepath.Join(uploadDir, tenantID, kbID)
		if err := os.MkdirAll(tenantDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "create upload dir failed"})
			return
		}

		filePath := filepath.Join(tenantDir, fmt.Sprintf("%s_%d%s", doc.ID.String(), time.Now().UnixNano(), filepath.Ext(file.Filename)))
		dst, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "save file failed"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "write file failed"})
			return
		}
		doc.FileURL = filePath
	}

	h.kbService.UpdateDocumentStatus(doc.ID.String(), "uploaded")

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": doc})
}

// ListDocuments 获取知识库文档列表
func (h *KnowledgeHandler) ListDocuments(c *gin.Context) {
	kbID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	docs, total, err := h.kbService.ListDocuments(kbID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     docs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
