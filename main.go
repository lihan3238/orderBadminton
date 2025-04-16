package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// 定义结构体来解析 API 返回的 JSON 数据
type TimeSlot struct {
	ID        int    `json:"id"`
	StrTime   string `json:"str_time"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type Resource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Data struct {
	Time     []TimeSlot           `json:"time"`
	Resource []Resource           `json:"resource"`
	Data     map[int]map[int]Sign `json:"data"`
}

type Sign struct {
	SignStatus int    `json:"sign_status"`
	Occupy     bool   `json:"occupy"`
	User       string `json:"user"`
}

// 定义配置结构体
type EmailConfig struct {
	From     string   `json:"from"`
	Password string   `json:"password"`
	To       []string `json:"to"`
	SMTPHost string   `json:"smtp_host"`
	SMTPPort string   `json:"smtp_port"`
}

// 声明全局配置变量
var emailConfig EmailConfig

// 加载配置
func loadEmailConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&emailConfig)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := loadEmailConfig("email_config.json")
	if err != nil {
		fmt.Println("加载邮件配置失败:", err)
		return
	}

	r := gin.Default()
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	r.GET("/api/status", func(c *gin.Context) {
		availableRooms := checkAvailability()
		if len(availableRooms) > 0 {
			sendEmailNotification(availableRooms)
		}
		c.JSON(http.StatusOK, gin.H{
			"available": len(availableRooms) > 0,
			"rooms":     availableRooms,
		})
	})

	r.Run(":8080")
}

// checkAvailability 访问新的 API 并解析是否有空闲场地
func checkAvailability() []string {
	url := "https://workflow.cuc.edu.cn/reservation/api/resource/large-screen?id=1293"

	// 发送 GET 请求获取 JSON 数据
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("请求失败:", err)
		return nil
	}
	defer resp.Body.Close()

	// 解析 JSON 数据
	var apiResponse struct {
		E string `json:"e"`
		M string `json:"m"`
		D Data   `json:"d"`
	}

	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return nil
	}

	var availableRooms []string

	// 遍历每个场地及其预约情况
	for _, resource := range apiResponse.D.Resource {
		// 遍历每个时间段
		for _, timeSlot := range apiResponse.D.Time {
			roomData := apiResponse.D.Data[resource.ID]
			slotData := roomData[timeSlot.ID]
			// 如果该时间段没有被占用，标记该场地为空闲
			if !slotData.Occupy {
				availableRooms = append(availableRooms, fmt.Sprintf("%s - %s", resource.Name, timeSlot.StrTime))
			}
		}
	}

	return availableRooms
}

// sendEmailNotification 发送邮件通知
func sendEmailNotification(availableRooms []string) {
	subject := "空闲场地提醒"
	body := fmt.Sprintf("以下场地是空闲的：\n\n%s", strings.Join(availableRooms, "\n"))
	message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))

	auth := smtp.PlainAuth("", emailConfig.From, emailConfig.Password, emailConfig.SMTPHost)

	err := smtp.SendMail(
		emailConfig.SMTPHost+":"+emailConfig.SMTPPort,
		auth,
		emailConfig.From,
		emailConfig.To,
		message,
	)
	if err != nil {
		fmt.Println("邮件发送失败:", err)
		return
	}
	fmt.Println("邮件发送成功")
}
