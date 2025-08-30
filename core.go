package gorest

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type Config struct {
	Handlers []gin.HandlerFunc
}

func GormHandlerFunc(db *gorm.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		tx := db.Session(&gorm.Session{NewDB: true})
		c.Set("db", tx)
		c.Next()
	}
}

func Get[T any](r *gin.Engine, path string, cfg *Config) {
	r.GET(path, append(cfg.Handlers, func(c *gin.Context) {
		var objs []T
		db := c.MustGet("db").(*gorm.DB)
		if err := db.Find(&objs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, objs)
	})...)

	r.GET(path+"/:id", append(cfg.Handlers, func(c *gin.Context) {
		var obj T
		db := c.MustGet("db").(*gorm.DB)
		id := c.Param("id")
		if err := db.First(&obj, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, obj)
	})...)
}

func Create[T any](r *gin.Engine, path string, cfg *Config) {
	r.POST(path, append(cfg.Handlers, func(c *gin.Context) {
		var obj T
		db := c.MustGet("db").(*gorm.DB)
		if err := c.ShouldBindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&obj).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, obj)
	})...)
}

func Update[T any](r *gin.Engine, path string, cfg *Config) {
	r.PUT(path+"/:id", append(cfg.Handlers, func(c *gin.Context) {
		var obj T
		db := c.MustGet("db").(*gorm.DB)
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
	})...)
}

func Delete[T any](r *gin.Engine, path string, cfg *Config) {
	r.DELETE(path+"/:id", append(cfg.Handlers, func(c *gin.Context) {
		var obj T
		db := c.MustGet("db").(*gorm.DB)
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
	})...)
}

func Rest[T any](r *gin.Engine, path string, cfg *Config) {
	Get[T](r, path, cfg)
	Create[T](r, path, cfg)
	Update[T](r, path, cfg)
	Delete[T](r, path, cfg)
}
