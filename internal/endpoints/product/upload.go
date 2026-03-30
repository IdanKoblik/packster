package product

import (
	"fmt"
	"net/http"
	"packster/internal/metrics"
	"packster/internal/utils"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

const mbInBytes = 1 << 20

// HandleUpload godoc
// @Summary      Upload a version artifact
// @Description  Uploads a file as a new version of a product. Requires Upload permission. Duplicate version names are rejected.
// @Tags         versions
// @Accept       multipart/form-data
// @Param        product    formData  string  true  "Product name"
// @Param        group_name formData  string  false "Product group (default: empty)"
// @Param        version    formData  string  true  "Version identifier"
// @Param        file       formData  file    true  "Artifact file"
// @Success      201  "Version uploaded"
// @Failure      400  {object}  map[string]string  "Invalid request or duplicate version"
// @Failure      403  {string}  string  "Permission denied"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/upload [post]
func (h *ProductHandler) HandleUpload(c *gin.Context) {
	var request types.UploadRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if h.FileSizeLimit > 0 && request.File.Size > int64(h.FileSizeLimit)*mbInBytes {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("file size exceeds the limit of %d MB", h.FileSizeLimit),
		})
		return
	}

	if err := utils.ValidateName(request.Product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.ValidateName(request.Version); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	safeFilename, err := utils.SafeFilename(request.File.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.Repo.FetchProduct(request.Product, request.GroupName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if product == nil {
		c.String(http.StatusBadRequest, "Product not found")
		return
	}

	permissions := product.Tokens[utils.Hash(c.GetString("token"))]
	if !c.GetBool("admin") && !permissions.Upload {
		c.String(http.StatusForbidden, "permission denied")
		return
	}

	_, ok := product.Versions[request.Version]
	if ok {
		c.String(http.StatusForbidden, "this version is already uploaded")
		return
	}

	var baseDir string
	if request.GroupName == "" {
		baseDir = request.Product
	} else {
		baseDir = request.GroupName + "/" + request.Product
	}
	location := fmt.Sprintf("./prodcuts/%s/%s/%s", baseDir, request.Version, safeFilename)

	err = c.SaveUploadedFile(request.File, location)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	checksum, err := utils.Checksum(request.File)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	version := types.Version{
		Path:     location,
		Checksum: checksum,
	}

	err = h.Repo.AddVersion(
		request.Product,
		request.GroupName,
		request.Version,
		c.GetString("token"),
		c.GetBool("admin"),
		version,
	)

	if err != nil {
		metrics.ArtifactUploadsTotal.WithLabelValues(request.Product, "error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	metrics.ArtifactUploadsTotal.WithLabelValues(request.Product, "success").Inc()
	metrics.ArtifactUploadBytesTotal.WithLabelValues(request.Product).Add(float64(request.File.Size))

	c.Status(http.StatusCreated)
}
