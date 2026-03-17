package product

import (
	"artifactor/internal/utils"
	"artifactor/pkg/types"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) HandleUpload(c *gin.Context) {
	var request types.UploadRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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

	product, err := h.Repo.FetchProduct(request.Product)
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
	if !c.GetBool("admin") || !permissions.Upload {
		c.String(http.StatusForbidden, "permission denied")
		return
	}

	_, ok := product.Versions[request.Version]
	if ok {
		c.String(http.StatusForbidden, "this version is already uploaded")
		return
	}

	location := fmt.Sprintf("./prodcuts/%s/%s/%s", request.Product, request.Version, safeFilename)
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
		request.Version,
		c.GetString("token"),
		c.GetBool("admin"),
		version,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusCreated)
}
