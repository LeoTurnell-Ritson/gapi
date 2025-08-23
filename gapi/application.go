package autocrud

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func API[T any](r *gin.Engine, db *gorm.DB, path string) {
	// Create
	r.POST(path, func(c *gin.Context) {
		var obj T
		if err := c.ShouldBindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&obj).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, obj)
	})

	// Get many
	r.GET(path, func(c *gin.Context) {
		var objs []T
		if err := db.Find(&objs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, objs)
	})

	// Get one
	r.GET(path+"/:id", func(c *gin.Context) {
		var obj T
		id := c.Param("id")
		if err := db.First(&obj, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, obj)
	})

	// Update
	r.PUT(path+"/:id", func(c *gin.Context) {
		var obj T
		id := c.Param("id")
		if err := db.First(&obj, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if err := c.ShouldBindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Save(&obj).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, obj)
	})

	// Delete
	r.DELETE(path+"/:id", func(c *gin.Context) {
		var obj T
		id := c.Param("id")
		if err := db.First(&obj, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if err := db.Delete(&obj).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
