package projects

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

type versionDTO struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
	Product  int    `json:"product"`
}

func toVersionDTO(v types.Version) versionDTO {
	return versionDTO{
		ID:       v.ID,
		Name:     v.Name,
		Path:     originalFilename(v.Path),
		Checksum: v.Checksum,
		Product:  v.Product,
	}
}

func validateVersionName(name string) error {
	if name == "" {
		return errors.New("version name is required")
	}
	if strings.ContainsAny(name, "/\\") || name == "." || name == ".." {
		return errors.New("invalid version name")
	}
	return nil
}

// HandleListVersions returns the versions under a product. Caller needs read
// access on the parent project.
func (h *ProjectsHandler) HandleListVersions(c *gin.Context) {
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

	if _, _, _, ok := h.authorize(c, product.Project, accessRead); !ok {
		return
	}

	versions, err := h.VersionRepo.ListByProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]versionDTO, 0, len(versions))
	for _, v := range versions {
		out = append(out, toVersionDTO(v))
	}
	c.JSON(http.StatusOK, out)
}

// HandleUploadVersion stores an uploaded artifact as a new version of the
// product. Validates the version name, enforces the per-server upload limit,
// and computes a sha256 checksum during the copy. Caller needs upload access.
func (h *ProjectsHandler) HandleUploadVersion(c *gin.Context) {
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

	if _, _, _, ok := h.authorize(c, product.Project, accessUpload); !ok {
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if err := validateVersionName(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if existing, err := h.VersionRepo.GetByName(productID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "version name already exists for this product"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if h.Cfg.FileUploadLimit > 0 {
		max := int64(h.Cfg.FileUploadLimit) * 1024 * 1024
		if fh.Size > max {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file exceeds upload limit"})
			return
		}
	}

	storedPath, checksum, err := storeBlob(h.Cfg.Storage.Path, product.Project, product.ID, fh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	version, err := h.VersionRepo.Create(product.ID, name, storedPath, checksum)
	if err != nil {
		removeBlob(storedPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toVersionDTO(*version))
}

// HandleDownloadVersion streams a version blob by its numeric id. Caller
// needs download access on the parent project.
func (h *ProjectsHandler) HandleDownloadVersion(c *gin.Context) {
	versionID, err := strconv.Atoi(c.Param("version_id"))
	if err != nil || versionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version id"})
		return
	}

	version, err := h.VersionRepo.GetByID(versionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if version == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	h.streamVersion(c, version)
}

// HandleDownloadByName resolves project_id + product_name + version_name to a
// version row and streams it. Caller needs download access on the project.
func (h *ProjectsHandler) HandleDownloadByName(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	productName := c.Param("product_name")
	versionName := c.Param("version_name")

	product, err := h.ProductRepo.GetByName(projectID, productName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	version, err := h.VersionRepo.GetByName(product.ID, versionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if version == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	h.streamVersion(c, version)
}

func (h *ProjectsHandler) streamVersion(c *gin.Context, version *types.Version) {
	product, err := h.ProductRepo.GetByID(version.Product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	_, _, perm, ok := h.authorize(c, product.Project, accessRead)
	if !ok {
		return
	}
	if !perm.CanDownload {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	c.Header("X-Checksum-SHA256", version.Checksum)
	c.Header("X-Version-Name", version.Name)
	c.FileAttachment(version.Path, originalFilename(version.Path))
}

// HandleDeleteVersion removes a version row and its blob on disk. Caller
// needs delete access on the parent project.
func (h *ProjectsHandler) HandleDeleteVersion(c *gin.Context) {
	versionID, err := strconv.Atoi(c.Param("version_id"))
	if err != nil || versionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version id"})
		return
	}

	version, err := h.VersionRepo.GetByID(versionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if version == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	product, err := h.ProductRepo.GetByID(version.Product)
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

	if err := h.VersionRepo.Delete(version.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	removeBlob(version.Path)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
