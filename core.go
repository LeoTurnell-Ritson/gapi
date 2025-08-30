package gorest

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"reflect"
)

type Config struct {
	Handlers []gin.HandlerFunc
	AddQueryParams bool
}

func GormHandlerFunc(db *gorm.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		tx := db.Session(&gorm.Session{NewDB: true})
		c.Set("db", tx)
		c.Next()
	}
}

func QueryParamsHandlerFuncs[T any]() []gin.HandlerFunc {
	var handlers []gin.HandlerFunc
	tType := reflect.TypeOf((*T)(nil)).Elem()
	if tType.Kind() != reflect.Struct {
		panic("QueryParamsHandlerFunc expects type T to be a struct type")
	}

	for i := range make([]struct{}, tType.NumField()) {
		var fieldName string

		field := tType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			fieldName = jsonTag
		} else {
			continue
		}

		handlers = append(handlers, func(fieldName string) gin.HandlerFunc {
			return func(c *gin.Context) {
				db := c.MustGet("db").(*gorm.DB)
				if value, exists := c.GetQuery(fieldName); exists {
					db = db.Where(fieldName+" = ?", value)
				}

				c.Set("db", db)
				c.Next()
			}
		}(fieldName))
	}

	return handlers
}

func Get[T any](r *gin.Engine, path string, cfg *Config) {
	var handlers []gin.HandlerFunc
	if cfg.AddQueryParams{
		handlers = append(cfg.Handlers, QueryParamsHandlerFuncs[T]()...)
	} else {
		handlers = cfg.Handlers
	}

	r.GET(path, append(handlers, func(c *gin.Context) {
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

func All[T any](r *gin.Engine, path string, cfg *Config) {
	Get[T](r, path, cfg)
	Create[T](r, path, cfg)
	Update[T](r, path, cfg)
	Delete[T](r, path, cfg)
}
