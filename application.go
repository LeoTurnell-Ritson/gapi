package gapi

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type FilterFunc[T any] func(db *gorm.DB, c *gin.Context) *gorm.DB

type Config struct {
	Handlers []gin.HandlerFunc
	Filters []FilterFunc[any]
}

func getQueryFromConfig[T any](db *gorm.DB, c *gin.Context, cfg *Config) *gorm.DB {
	query := db.Model(new(T))
	if cfg.Filters != nil {
		for _, filter := range cfg.Filters {
			query = filter(db, c)
		}
	}

	return query
}

func getHandlersFromConfig(cfg *Config) []gin.HandlerFunc {
	handlers := []gin.HandlerFunc{}

	if cfg.Handlers != nil {
		handlers = append(handlers, cfg.Handlers...)
	}

	return handlers
}

func GET[T any](r *gin.Engine, db *gorm.DB, path string, cfg *Config) {
	r.GET(path, append(getHandlersFromConfig(cfg), func(c *gin.Context) {
		var objs []T
		query := getQueryFromConfig[T](db, c, cfg)

		if err := query.Find(&objs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, objs)
	})...)

	r.GET(path+"/:id", append(getHandlersFromConfig(cfg), func(c *gin.Context) {
		var obj T
		query := getQueryFromConfig[T](db, c, cfg)

		id := c.Param("id")
		if err := query.First(&obj, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, obj)
	})...)
}

func POST[T any](r *gin.Engine, db *gorm.DB, path string, cfg *Config) {
	r.POST(path, append(getHandlersFromConfig(cfg), func(c *gin.Context) {
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
	})...)
}

func PUT[T any](r *gin.Engine, db *gorm.DB, path string, cfg *Config) {
	r.PUT(path+"/:id", append(getHandlersFromConfig(cfg), func(c *gin.Context) {
		var obj T
		query := getQueryFromConfig[T](db, nil, cfg)

		id := c.Param("id")
		if err := query.First(&obj, id).Error; err != nil {
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

func DELETE[T any](r *gin.Engine, db *gorm.DB, path string, cfg *Config) {
	r.DELETE(path+"/:id", append(getHandlersFromConfig(cfg), func(c *gin.Context) {
		var obj T
		query := getQueryFromConfig[T](db, nil, cfg)

		id := c.Param("id")
		if err := query.First(&obj, id).Error; err != nil {
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
