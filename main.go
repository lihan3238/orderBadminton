package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
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

type DetailedData struct {
	Time []TimeSlot `json:"time"`
	Day  []string   `json:"day"`
	Data map[string]map[string]struct {
		Sign_status int    `json:"sign_status"`
		Occupy      bool   `json:"occupy"`
		User        string `json:"user"`
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

var lastAvailableSummary string

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
		today_date := time.Now().Format("2006-01-02")
		tomorrow_date := time.Now().Add(24 * time.Hour).Format("2006-01-02")
		today, tomorrow := checkAvailability(today_date), checkAvailability(tomorrow_date)

		// 构建当前这次的摘要（用换行连接确保顺序一致）
		currentSummary := strings.Join(append(today, tomorrow...), "\n")

		// 如果内容不同才发邮件
		if currentSummary != "" && currentSummary != lastAvailableSummary {
			sendEmailNotification(today, tomorrow)
			lastAvailableSummary = currentSummary
		}

		c.JSON(http.StatusOK, gin.H{
			"today_available":    today,
			"tomorrow_available": tomorrow,
		})
	})

	r.Run(":8080")
}

func checkAvailability(date string) []string {
	//tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	date_name := date
	if date == time.Now().Format("2006-01-02") {
		date_name = "今天"
	} else if date == time.Now().Add(24*time.Hour).Format("2006-01-02") {
		date_name = "明天"
	}
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

		now := time.Now()

		dayData, exists := result.D.Data[date]
		if !exists {
			continue
		}
		for _, slot := range result.D.Time {

			// 从 "15:00-16:00" 中提取起始时间 "15:00"
			timeParts := strings.Split(slot.StrTime, "-")
			if len(timeParts) < 1 {
				continue
			}
			startTimeStr := strings.TrimSpace(timeParts[0])
			fullTimeStr := fmt.Sprintf("%s %s", date, startTimeStr)

			startTime, err := time.ParseInLocation("2006-01-02 15:04", fullTimeStr, time.Local)
			if err != nil {
				fmt.Println("时间解析失败:", fullTimeStr)
				continue
			}
			if now.After(startTime) {
				// 当前时间晚于场次开始时间，跳过
				//fmt.Println("跳过已开始的场次:", fullTimeStr)
				continue
			}

			slotID := fmt.Sprintf("%d", slot.ID)
			if item, ok := dayData[slotID]; ok && !item.Occupy {
				free = append(free, fmt.Sprintf("【%s】场地ID %d %s", date_name, id-1293, slot.StrTime))
			}
		}
	}
	return free
}

func sendEmailNotification(today, tomorrow []string) {
	// 构造昵称和 From 头（RFC2047 编码）
	subject := "羽毛球场地空闲提醒"
	nickname := "CUCBadminton 小助手"
	encodedNickname := mime.BEncoding.Encode("UTF-8", nickname)
	fromHeader := fmt.Sprintf("%s <%s>", encodedNickname, emailConfig.From)

	// 构造邮件正文并 Base64 编码
	var bodyBuilder strings.Builder
	if len(today) > 0 {
		bodyBuilder.WriteString("今天空闲场地:\n" + strings.Join(today, "\n") + "\n\n")
	}
	if len(tomorrow) > 0 {
		bodyBuilder.WriteString("明天空闲场地:\n" + strings.Join(tomorrow, "\n"))
	}
	bodyEncoded := base64.StdEncoding.EncodeToString([]byte(bodyBuilder.String()))

	// 完整邮件消息（含 MIME 头）
	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: =?UTF-8?B?%s?=\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"Content-Transfer-Encoding: base64\r\n\r\n"+
			"%s",
		fromHeader,
		strings.Join(emailConfig.To, ", "),
		base64.StdEncoding.EncodeToString([]byte(subject)),
		bodyEncoded,
	))

	// --- 隐式 TLS 连接 (SMTPS) ---
	addr := emailConfig.SMTPHost + ":" + emailConfig.SMTPPort
	// 1. 建立 TLS 连接
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,                 // 如需校验证书，请设为 false，并确保 System CA 可用
		ServerName:         emailConfig.SMTPHost, // 用于 SNI
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig) // 隐式 TLS 握手&#8203;:contentReference[oaicite:6]{index=6}
	if err != nil {
		fmt.Println("TLS 连接失败:", err)
		return
	}
	defer conn.Close()

	// 2. 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, emailConfig.SMTPHost)
	if err != nil {
		fmt.Println("SMTP 客户端创建失败:", err)
		return
	}
	defer client.Quit()

	// 3. 认证
	auth := smtp.PlainAuth("", emailConfig.From, emailConfig.Password, emailConfig.SMTPHost) // port independent :contentReference[oaicite:7]{index=7}
	if err = client.Auth(auth); err != nil {
		fmt.Println("认证失败:", err)
		return
	}

	// 4. 发件人 & 收件人
	if err = client.Mail(emailConfig.From); err != nil {
		fmt.Println("发件人设置失败:", err)
		return
	}
	for _, rcpt := range emailConfig.To {
		if err = client.Rcpt(rcpt); err != nil {
			fmt.Println("收件人设置失败:", err)
			return
		}
	}

	// 5. 写入消息体
	wc, err := client.Data()
	if err != nil {
		fmt.Println("进入 Data 模式失败:", err)
		return
	}
	if _, err = wc.Write(msg); err != nil {
		fmt.Println("消息发送失败:", err)
	}
	wc.Close()
	fmt.Println("邮件发送成功，时间：", time.Now().Format("15:04:05"))

}
