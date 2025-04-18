# CUC 羽毛球场地监测（捡漏）

## 功能

- 定期监测羽毛球场地的使用情况
- 今明两日空闲场地自动邮件通知

## 使用

### 1-从代码部署

1. 下载代码

```bash
git clone github.com/lihan3238/orderBadminton.git
cd orderBadminton
```

2. 安装依赖

```bash
go get github.com/gin-gonic/gin
```

3. 配置邮箱

在目录下创建 `email_config.json` 文件并进行配置：

```json
{
    "from": "xxx@xxx.com",          // 发件人邮箱
    "password": "qwedqwwdqw",       // 发件人邮箱授权码（SMTP）
    "to": [                         // 需要接收提醒的收件人邮箱
        "xxx@xxx.com",
        "mmm@mmm.com"
    ],
    "smtp_host": "smtp.xxx.com",    // SMTP服务器地址
    "smtp_port": "25"               // SMTP服务器端口（非SSL）
}
```

4. 运行

```bash
go run main.go
```

- 默认每 30s 监测一次场地使用情况，如需修改，请在 `/static/script.js` 中修改 `setInterval` 的参数:

```js
setInterval(updateStatus, 15000);//将 15000 修改为你想要的时间间隔（单位：ms）
```

### 2-从 EXE 应用部署

1. 下载 EXE 应用

从 Github 下载[orderBadminton](https://github.com/lihan3238/orderBadminton/releases/latest)

2. 目录结构
    
```bash
project/
├── orderBadminton.exe
├── email_config.json
└── static/
    ├── index.html
    ├── style.css
    └── script.js
```

3. 配置邮箱

在目录下创建 `email_config.json` 文件并进行配置：

```json
{
    "from": "xxx@xxx.com",          // 发件人邮箱
    "password": "qwedqwwdqw",       // 发件人邮箱授权码（SMTP）
    "to": [                         // 需要接收提醒的收件人邮箱
        "xxx@xxx.com",
        "mmm@mmm.com"
    ],
    "smtp_host": "smtp.xxx.com",    // SMTP服务器地址
    "smtp_port": "25"               // SMTP服务器端口（非SSL）
}
```

4. 运行

```bash
./orderBadminton.exe
```

- 默认每 30s 监测一次场地使用情况，如需修改，请在 `/static/script.js` 中修改 `setInterval` 的参数:

```js
setInterval(updateStatus, 30000);//将 30000 修改为你想要的时间间隔（单位：ms）
```

## 注意

- [羽毛球场地](https://workflow.cuc.edu.cn/reservation/fe/site/appointmentscreen?id=1293)为无需登录的当日羽毛球场地核销情况页面。

- [羽毛球场地](https://workflow.cuc.edu.cn/reservation/fe/site/appointmentscreen?id=1293)的数据来自 [API：(https://workflow.cuc.edu.cn/reservation/api/resource/large-screen?id=1293](https://workflow.cuc.edu.cn/reservation/api/resource/large-screen?id=1293)，其中 `id=1293` 为室内羽毛球场地的 id。

- [API：(https://workflow.cuc.edu.cn/reservation/api/resource/large-screen?id=n](https://workflow.cuc.edu.cn/reservation/api/resource/large-screen?id=1294)，其中 `id=1294-1303` 时，数据为一到十号场地 day-2 到 day+4 的七日预约情况。

- 163 邮箱 SMTP 服务器地址为 `smtp.163.com`，非 SSL 端口为 `25`，SSL 端口为 `465`。
- QQ 邮箱 SMTP 服务器地址为 `smtp.qq.com`，SSL 端口为 `465`。

## 声明

本项目完全出于学习目的，禁止用于任何商业用途。
本项目不对任何人或组织负责，使用本项目的代码和数据所产生的后果由使用者自行承担。