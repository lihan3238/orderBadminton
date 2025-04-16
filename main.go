package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// API结构体
type TimeSlot struct {
	ID      int    `json:"id"`
	StrTime string `json:"str_time"`
}
type Resource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type LargeScreenData struct {
	Time     []TimeSlot `json:"time"`
	Resource []Resource `json:"resource"`
	Data     map[int]map[int]struct {
		Occupy bool   `json:"occupy"`
		User   string `json:"user"`
	} `json:"data"`
}
type DetailedData struct {
	Time []TimeSlot `json:"time"`
	Day  []string   `json:"day"`
	Data map[string]map[string]struct {
		Status   int     `json:"status"`
		Username *string `json:"username"`
		Text     string  `json:"text"`
	} `json:"data"`
}
type EmailConfig struct {
	From     string   `json:"from"`
	Password string   `json:"password"`
	To       []string `json:"to"`
	SMTPHost string   `json:"smtp_host"`
	SMTPPort string   `json:"smtp_port"`
}

var emailConfig EmailConfig

func loadEmailConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(&emailConfig)
}

func main() {
	if err := loadEmailConfig("email_config.json"); err != nil {
		fmt.Println("加载邮件配置失败:", err)
		return
	}

	r := gin.Default()
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})
	r.GET("/api/status", func(c *gin.Context) {
		today, tomorrow := checkTodayAvailability(), checkTomorrowAvailability()
		if len(today)+len(tomorrow) > 0 {
			sendEmailNotification(today, tomorrow)
		}
		c.JSON(http.StatusOK, gin.H{
			"today_available":    today,
			"tomorrow_available": tomorrow,
		})
	})
	r.Run(":8080")
}

func checkTodayAvailability() []string {
	resp, err := http.Get("https://workflow.cuc.edu.cn/reservation/api/resource/large-screen?id=1293")
	if err != nil {
		fmt.Println("请求失败:", err)
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		D LargeScreenData `json:"d"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("JSON解析失败:", err)
		return nil
	}

	var free []string
	for _, res := range result.D.Resource {
		for _, slot := range result.D.Time {
			if info, ok := result.D.Data[res.ID][slot.ID]; ok && !info.Occupy {
				free = append(free, fmt.Sprintf("【今天】%s %s", res.Name, slot.StrTime))
			}
		}
	}
	return free
}

func checkTomorrowAvailability() []string {
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	var free []string
	for id := 1294; id <= 1303; id++ {
		url := fmt.Sprintf("https://workflow.cuc.edu.cn/reservation/api/resource/large-screen?id=%d", id)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("请求失败:", err)
			continue
		}
		var result struct {
			D DetailedData `json:"d"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Println("JSON解析失败:", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		dayData, exists := result.D.Data[tomorrow]
		if !exists {
			continue
		}
		for _, slot := range result.D.Time {
			slotID := fmt.Sprintf("%d", slot.ID)
			if item, ok := dayData[slotID]; ok && item.Username == nil {
				free = append(free, fmt.Sprintf("【明天】场地ID %d %s", id-1293, slot.StrTime))
			}
		}
	}
	return free
}

func sendEmailNotification(today, tomorrow []string) {
	subject := "羽毛球场地空闲提醒"
	var body strings.Builder
	if len(today) > 0 {
		body.WriteString("今天空闲场地:\n" + strings.Join(today, "\n") + "\n\n")
	}
	if len(tomorrow) > 0 {
		body.WriteString("明天空闲场地:\n" + strings.Join(tomorrow, "\n"))
	}
	message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body.String()))

	auth := smtp.PlainAuth("", emailConfig.From, emailConfig.Password, emailConfig.SMTPHost)
	err := smtp.SendMail(emailConfig.SMTPHost+":"+emailConfig.SMTPPort, auth, emailConfig.From, emailConfig.To, message)
	if err != nil {
		fmt.Println("邮件发送失败:", err)
	} else {
		fmt.Println("邮件发送成功")
	}
}
