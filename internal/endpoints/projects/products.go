package projects

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type productDTO struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Project int    `json:"project"`
}

type createProductRequest struct {
	Name string `json:"name"`
}

// HandleListProducts returns the products under a project. Caller needs read
// access (any of download/upload/delete).
func (h *ProjectsHandler) HandleListProducts(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	if _, _, _, ok := h.authorize(c, projectID, accessRead); !ok {
		return
	}

	products, err := h.ProductRepo.ListByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]productDTO, 0, len(products))
	for _, p := range products {
		out = append(out, productDTO{ID: p.ID, Name: p.Name, Project: p.Project})
	}
	c.JSON(http.StatusOK, out)
}

// HandleCreateProduct creates a product under the project. Caller needs
// upload access. Names must be unique within the project.
func (h *ProjectsHandler) HandleCreateProduct(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	if _, _, _, ok := h.authorize(c, projectID, accessUpload); !ok {
		return
	}

	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	product, err := h.ProductRepo.Create(projectID, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, productDTO{ID: product.ID, Name: product.Name, Project: product.Project})
}

// HandleDeleteProduct deletes the product, every version under it, and the
// version blobs from disk. Caller needs delete access on the parent project.
func (h *ProjectsHandler) HandleDeleteProduct(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := h.ProductRepo.GetByID(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	if _, _, _, ok := h.authorize(c, product.Project, accessDelete); !ok {
		return
	}

	versions, err := h.VersionRepo.ListByProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, v := range versions {
		removeBlob(v.Path)
		if err := h.VersionRepo.Delete(v.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := h.ProductRepo.Delete(productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
