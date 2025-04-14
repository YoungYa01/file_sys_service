package middleware

import (
	"bytes"
	"encoding/json"
	"gin_back/app/models"
	"github.com/gin-gonic/gin"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mssola/user_agent"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func LogMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间和基础信息
		start := time.Now()
		logEntry := models.Log{
			CreatedAt: start,
			Method:    c.Request.Method,
			ApiUrl:    c.Request.URL.Path,
			Ip:        c.ClientIP(),
		}

		// 读取请求参数
		var paramsBytes []byte
		if c.Request.Method != http.MethodGet {
			paramsBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(paramsBytes))
			logEntry.Params = string(paramsBytes)
		}

		// 处理请求前获取用户信息
		if claims, exists := c.Get("claims"); exists {
			if customClaims, ok := claims.(*models.CustomClaims); ok {
				logEntry.UserId = customClaims.UserID
				logEntry.UserName = customClaims.UserName // 需要确保claims包含用户名
				log.Println("用户信息:", customClaims)
				log.Println("用户ID:", customClaims.UserID)
				log.Println("用户名:", customClaims.UserName)
			}
			log.Println("logEntry用户ID:", logEntry.UserId)
			log.Println("logEntry用户名:", logEntry.UserName)
		}
		marshalIndent, _ := json.MarshalIndent(logEntry, "", "  ")
		log.Println("logEntry is", string(marshalIndent))

		// 解析User-Agent
		if uaString := c.Request.UserAgent(); uaString != "" {
			ua := user_agent.New(uaString)
			logEntry.Browser, _ = ua.Browser()
			logEntry.Os = ua.OS()
		}

		// 处理请求
		c.Next()

		// 补充请求后信息
		logEntry.UpdatedAt = time.Now()

		if c.Request.Method != http.MethodGet {
			// 异步处理地理位置和存储
			go func(entry models.Log) {
				// 获取地理位置
				entry.Province, entry.City = getLocation(entry.Ip)
				// 存储到数据库
				if err := db.Create(&entry).Error; err != nil {
					log.Printf("操作日志存储失败: %v", err)
				}
			}(logEntry)
		}

	}
}

var ipSearcher *xdb.Searcher

// 增加初始化错误处理
func InitIPDatabase() {
	dbPath := filepath.Join("data", "ip2region.xdb")

	// 检查文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("IP数据库文件不存在: %s", dbPath)
	}

	cBuff, err := xdb.LoadContentFromFile(dbPath)
	if err != nil {
		log.Fatalf("加载IP数据库失败: %v", err)
	}

	searcher, err := xdb.NewWithBuffer(cBuff)
	if err != nil {
		log.Fatalf("创建IP搜索器失败: %v", err)
	}

	ipSearcher = searcher
	log.Println("IP数据库初始化成功")
}

// 获取地理位置（使用本地IP库版本）
func getLocation(ip string) (province, city string) {
	// 处理本地回环地址
	if ip == "::1" || ip == "127.0.0.1" {
		return "本地", "内网"
	}

	if ipSearcher == nil || ip == "" {
		return "", ""
	}

	// 执行查询
	regionStr, err := ipSearcher.SearchByStr(ip)
	if err != nil {
		return "", ""
	}

	/*
		解析格式（示例）：
		国家|区域|省份|城市|ISP
		中国|0|浙江省|杭州市|电信
	*/
	parts := strings.Split(regionStr, "|")
	if len(parts) < 5 {
		return "", ""
	}

	log.Println(parts)

	// 省份城市处理
	province = parts[2]
	city = parts[3]

	// 转换行政区划代码（需要自行维护映射表）
	return province, city
}
