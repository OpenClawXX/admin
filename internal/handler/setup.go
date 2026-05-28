package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"crypto/rand"
	"encoding/hex"

	"ops-platform/internal/pkg/auth"
	"ops-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type SetupHandler struct {
	initFilePath string
}

func NewSetupHandler() *SetupHandler {
	return &SetupHandler{initFilePath: ".initialized"}
}

func (h *SetupHandler) Status(c *gin.Context) {
	data, err := os.ReadFile(h.initFilePath)
	installed := err == nil && strings.TrimSpace(string(data)) == "ok"
	c.JSON(200, gin.H{"installed": installed})
}

// Probe 连接数据库并列出所有数据库及其状态
func (h *SetupHandler) Probe(c *gin.Context) {
	var req struct {
		DBHost     string `json:"db_host" binding:"required"`
		DBPort     string `json:"db_port"`
		DBUser     string `json:"db_user" binding:"required"`
		DBPassword string `json:"db_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写数据库连接信息")
		return
	}
	if req.DBPort == "" {
		req.DBPort = "5432"
	}

	dsn := "host=" + req.DBHost + " port=" + req.DBPort + " user=" + req.DBUser + " password=" + req.DBPassword + " dbname=postgres sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Printf("[安装向导] 数据库连接失败: %v", err)
		response.BadRequest(c, "连接失败，请检查主机/端口/用户名/密码")
		return
	}
	defer db.Close()

	// List databases (exclude template/system databases)
	var databases []map[string]interface{}
	rows, err := db.Query(`SELECT d.datname,
		pg_catalog.pg_get_userbyid(d.datdba) as owner,
		pg_catalog.pg_encoding_to_char(d.encoding) as encoding
		FROM pg_catalog.pg_database d
		WHERE d.datname NOT IN ('template0', 'template1', 'postgres')
		ORDER BY d.datname`)
	if err != nil {
		response.InternalError(c, "查询数据库列表失败")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, owner, encoding string
		if err := rows.Scan(&name, &owner, &encoding); err != nil {
			continue
		}

		// Check table count for each database
		var tableCount int
		childDSN := "host=" + req.DBHost + " port=" + req.DBPort + " user=" + req.DBUser + " password=" + req.DBPassword + " dbname=" + name + " sslmode=disable"
		childDB, err := sqlx.Connect("postgres", childDSN)
		if err == nil {
			childDB.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'").Scan(&tableCount)
			childDB.Close()
		}

		databases = append(databases, gin.H{
			"name":        name,
			"owner":       owner,
			"encoding":    encoding,
			"table_count": tableCount,
			"has_data":    tableCount > 0,
		})
	}

	if databases == nil {
		databases = []map[string]interface{}{}
	}

	response.Success(c, gin.H{
		"databases": databases,
		"connected": true,
	})
}

func (h *SetupHandler) Execute(c *gin.Context) {
	// Check if already initialized - read file content to verify
	if data, err := os.ReadFile(h.initFilePath); err == nil && strings.TrimSpace(string(data)) == "ok" {
		response.BadRequest(c, "安装锁定文件已存在，请删除 .initialized 文件后再次尝试")
		return
	}

	var req struct {
		// Database
		DBHost     string `json:"db_host"`
		DBPort     string `json:"db_port"`
		DBUser     string `json:"db_user"`
		DBPassword string `json:"db_password"`
		DBName     string `json:"db_name"`
		DBMode     string `json:"db_mode"` // "fresh" or "import"
		// Admin
		AdminUsername string `json:"admin_username"`
		AdminPassword string `json:"admin_password"`
		AdminRealName string `json:"admin_realname"`
		// Branding
		CompanyName  string `json:"company_name"`
		PlatformName string `json:"platform_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	// Validate required fields
	if req.DBHost == "" || req.DBUser == "" || req.DBName == "" {
		response.BadRequest(c, "请填写数据库配置")
		return
	}
	if req.AdminUsername == "" || req.AdminPassword == "" {
		response.BadRequest(c, "请填写管理员账号密码")
		return
	}
	if len(req.AdminPassword) < 6 {
		response.BadRequest(c, "密码至少需要6个字符")
		return
	}

	// Validate database name (alphanumeric and underscores only)
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(req.DBName) {
		response.BadRequest(c, "数据库名称只能包含字母、数字和下划线")
		return
	}

	// Step 1: Connect to default postgres database to create target DB if needed
	adminDSN := "host=" + req.DBHost + " port=" + req.DBPort + " user=" + req.DBUser + " password=" + req.DBPassword + " dbname=postgres sslmode=disable"
	log.Printf("[安装向导] 连接数据库: host=%s port=%s user=%s dbname=%s", req.DBHost, req.DBPort, req.DBUser, req.DBName)
	adminDB, err := sqlx.Connect("postgres", adminDSN)
	if err != nil {
		log.Printf("[安装向导] 数据库连接失败: %v", err)
		response.BadRequest(c, "数据库连接失败，请检查主机/端口/用户名/密码")
		return
	}

	// Check if target database exists, create if not
	var exists bool
	adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", req.DBName).Scan(&exists)
	if !exists {
		log.Printf("[安装向导] 创建数据库: %s", req.DBName)
		_, err = adminDB.Exec("CREATE DATABASE " + req.DBName)
		if err != nil {
			log.Printf("[安装向导] 创建数据库失败: %v", err)
			adminDB.Close()
			response.BadRequest(c, "创建数据库失败，请检查数据库配置")
			return
		}
	}
	adminDB.Close()

	// Step 2: Connect to the target database
	dsn := "host=" + req.DBHost + " port=" + req.DBPort + " user=" + req.DBUser + " password=" + req.DBPassword + " dbname=" + req.DBName + " sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Printf("[安装向导] 连接目标数据库失败: %v", err)
		response.BadRequest(c, "连接目标数据库失败，请检查配置")
		return
	}
	defer db.Close()

	// Step 3: Determine mode - use user's choice or auto-detect
	var dbMode string
	var tableCount int
	if req.DBMode == "fresh" {
		// User chose fresh install - always run migrations
		dbMode = "new"
		log.Printf("[安装向导] 用户选择全新安装")
		if err := runMigrations(db); err != nil {
			log.Printf("[安装向导] 数据库迁移失败: %v", err)
			response.InternalError(c, "数据库迁移失败")
			return
		}
	} else if req.DBMode == "import" {
		// User chose import - skip migrations, verify tables
		db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'").Scan(&tableCount)
		dbMode = "existing"
		log.Printf("[安装向导] 用户选择导入模式（%d 张表）", tableCount)

		// Verify critical tables exist
		requiredTables := []string{"users", "tickets", "projects", "teams"}
		for _, t := range requiredTables {
			var cnt int
			db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=$1", t).Scan(&cnt)
			if cnt == 0 {
				log.Printf("[安装向导] 导入模式缺少必要表: %s，将执行迁移补齐", t)
				if err := runMigrations(db); err != nil {
					response.InternalError(c, "数据库缺少必要表且补齐失败: "+t)
					return
				}
				break
			}
		}
	} else {
		// Auto-detect mode (fallback)
		err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'").Scan(&tableCount)
		if err != nil {
			response.InternalError(c, "检查数据库状态失败")
			return
		}
		if tableCount > 0 {
			dbMode = "existing"
			log.Printf("[安装向导] 自动检测: 已有数据库（%d 张表）", tableCount)
		} else {
			dbMode = "new"
			log.Printf("[安装向导] 自动检测: 空数据库")
			if err := runMigrations(db); err != nil {
				response.InternalError(c, "数据库迁移失败")
				return
			}
		}
	}

	// Create admin user
	hashedPassword, err := auth.HashPassword(req.AdminPassword)
	if err != nil {
		response.InternalError(c, "密码加密失败")
		return
	}

	if req.AdminRealName == "" {
		req.AdminRealName = "系统管理员"
	}

	_, err = db.Exec(`INSERT INTO users (username, password, real_name, role, status) VALUES ($1, $2, $3, 'admin', 1)
		ON CONFLICT (username) DO UPDATE SET password = $2, real_name = $3, role = 'admin', status = 1, updated_at = NOW()`,
		req.AdminUsername, hashedPassword, req.AdminRealName)
	if err != nil {
		log.Printf("[安装向导] 创建管理员账号失败: %v", err)
		response.InternalError(c, "创建管理员账号失败")
		return
	}

	// Save branding config
	platformName := req.PlatformName
	if platformName == "" {
		platformName = "运维管理平台"
	}
	companyName := req.CompanyName
	if companyName == "" {
		companyName = "成都商惠计算机系统有限公司"
	}

	brandingJSON, _ := json.Marshal(map[string]string{
		"platform_name": platformName,
		"company_name":  companyName,
	})
	if _, err := db.Exec(`INSERT INTO system_configs (key, value) VALUES ('branding', $1) ON CONFLICT (key) DO UPDATE SET value = $1`, brandingJSON); err != nil {
		log.Printf("[安装向导] 保存品牌配置失败: %v", err)
	}

	// Save .env file
	jwtSecretBytes := make([]byte, 32)
	if _, err := rand.Read(jwtSecretBytes); err != nil {
		response.InternalError(c, "生成密钥失败")
		return
	}
	jwtSecret := hex.EncodeToString(jwtSecretBytes)

	envContent := "SERVER_PORT=1365\nGIN_MODE=release\n" +
		"DB_HOST=" + req.DBHost + "\n" +
		"DB_PORT=" + req.DBPort + "\n" +
		"DB_USER=" + req.DBUser + "\n" +
		"DB_PASSWORD=" + req.DBPassword + "\n" +
		"DB_NAME=" + req.DBName + "\n" +
		"DB_SSLMODE=disable\n" +
		"JWT_SECRET=" + jwtSecret + "\n" +
		"JWT_EXPIRE_HOUR=24\n"
	if err := os.WriteFile(".env", []byte(envContent), 0600); err != nil {
		log.Printf("[安装向导] 保存 .env 失败: %v", err)
		response.InternalError(c, "保存配置失败")
		return
	}

	// Mark as initialized
	if err := os.WriteFile(h.initFilePath, []byte("ok"), 0644); err != nil {
		log.Printf("[安装向导] 创建初始化标记失败: %v", err)
		response.InternalError(c, "初始化标记失败")
		return
	}

	dbMessage := "新数据库已创建并初始化"
	if dbMode == "existing" {
		dbMessage = fmt.Sprintf("已接入现有数据库（%d 张表），管理员账号已更新", tableCount)
	}

	response.Success(c, gin.H{
		"message":       "安装完成",
		"db_mode":       dbMode,
		"db_message":    dbMessage,
		"admin_user":    req.AdminUsername,
		"platform_name": platformName,
		"company_name":  companyName,
	})

	// Auto-restart after response is sent
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("安装完成，正在重启服务器...")
		os.Exit(0)
	}()
}

func runMigrations(db *sqlx.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY, username VARCHAR(50) NOT NULL UNIQUE, password VARCHAR(255) NOT NULL,
			real_name VARCHAR(50) DEFAULT '', email VARCHAR(100) DEFAULT '', phone VARCHAR(20) DEFAULT '',
			role VARCHAR(20) NOT NULL DEFAULT 'engineer', team_id BIGINT, skills JSONB DEFAULT '[]',
			status INT DEFAULT 1, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS teams (
			id BIGSERIAL PRIMARY KEY, name VARCHAR(100) NOT NULL,
			supervisor_id BIGINT, description TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id BIGSERIAL PRIMARY KEY, code VARCHAR(20) DEFAULT '', name VARCHAR(200) NOT NULL,
			description TEXT DEFAULT '', type VARCHAR(20) DEFAULT 'daily', priority VARCHAR(10) DEFAULT 'medium',
			status VARCHAR(20) DEFAULT 'active', requester VARCHAR(200) DEFAULT '',
			manager_id BIGINT, budget NUMERIC(12,2), remark TEXT DEFAULT '',
			start_date TIMESTAMP WITH TIME ZONE, end_date TIMESTAMP WITH TIME ZONE,
			actual_end_date TIMESTAMP WITH TIME ZONE, rectification TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS assets (
			id BIGSERIAL PRIMARY KEY, name VARCHAR(200) NOT NULL, type VARCHAR(50) NOT NULL,
			ip VARCHAR(50) DEFAULT '', status VARCHAR(20) DEFAULT 'active',
			location VARCHAR(200) DEFAULT '', serial_number VARCHAR(100) DEFAULT '',
			brand VARCHAR(100) DEFAULT '', model VARCHAR(100) DEFAULT '',
			purchase_date DATE, warranty_date DATE, responsible_id BIGINT,
			description TEXT DEFAULT '', metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS tickets (
			id BIGSERIAL PRIMARY KEY, title VARCHAR(500) NOT NULL, description TEXT DEFAULT '',
			type VARCHAR(20) NOT NULL, priority VARCHAR(10) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'created',
			creator_id BIGINT, assignee_id BIGINT, project_id BIGINT, asset_id BIGINT,
			alert_id BIGINT, sla_deadline TIMESTAMP WITH TIME ZONE,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_logs (
			id BIGSERIAL PRIMARY KEY, ticket_id BIGINT NOT NULL,
			action VARCHAR(50) NOT NULL, operator_id BIGINT NOT NULL,
			content TEXT DEFAULT '', created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_completions (
			id BIGSERIAL PRIMARY KEY, ticket_id BIGINT NOT NULL UNIQUE,
			solution TEXT NOT NULL DEFAULT '', root_cause TEXT DEFAULT '',
			result VARCHAR(20) DEFAULT 'resolved', impact TEXT DEFAULT '',
			remaining TEXT DEFAULT '', suggestion TEXT DEFAULT '',
			handover TEXT DEFAULT '', created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_files (
			id BIGSERIAL PRIMARY KEY, ticket_id BIGINT NOT NULL,
			filename VARCHAR(500) NOT NULL, filepath VARCHAR(500) NOT NULL,
			filesize BIGINT DEFAULT 0, filetype VARCHAR(50) DEFAULT '',
			uploader_id BIGINT NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS scores (
			id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL,
			ticket_id BIGINT, period VARCHAR(20) NOT NULL,
			dimension VARCHAR(50) NOT NULL, score NUMERIC(5,2) DEFAULT 0,
			remark TEXT DEFAULT '', created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS kb_categories (
			id BIGSERIAL PRIMARY KEY, name VARCHAR(200) NOT NULL,
			parent_id BIGINT DEFAULT 0, sort_order INT DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS kb_articles (
			id BIGSERIAL PRIMARY KEY, title VARCHAR(500) NOT NULL,
			content TEXT DEFAULT '', content_html TEXT DEFAULT '',
			category_id BIGINT DEFAULT 0, status VARCHAR(20) DEFAULT 'draft',
			author_id BIGINT NOT NULL, tags JSONB DEFAULT '[]',
			view_count INT DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS kb_files (
			id BIGSERIAL PRIMARY KEY, article_id BIGINT NOT NULL,
			filename VARCHAR(500) NOT NULL, filepath VARCHAR(500) NOT NULL,
			filesize BIGINT DEFAULT 0, filetype VARCHAR(50) DEFAULT '',
			uploader_id BIGINT NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS monitor_alerts (
			id BIGSERIAL PRIMARY KEY, source VARCHAR(50) NOT NULL,
			alert_name VARCHAR(200) NOT NULL, level VARCHAR(20) NOT NULL,
			ticket_id BIGINT, raw_payload JSONB DEFAULT '{}',
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL,
			title VARCHAR(200) NOT NULL, content TEXT DEFAULT '',
			type VARCHAR(20) DEFAULT 'info', is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS operation_logs (
			id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL,
			action VARCHAR(50) NOT NULL, resource VARCHAR(50) NOT NULL,
			resource_id BIGINT DEFAULT 0, detail TEXT DEFAULT '',
			ip VARCHAR(50) DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS project_members (
			id BIGSERIAL PRIMARY KEY, project_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(project_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS project_rectifications (
			id BIGSERIAL PRIMARY KEY, project_id BIGINT NOT NULL,
			type VARCHAR(20) NOT NULL DEFAULT 'rectification',
			content TEXT NOT NULL DEFAULT '', operator_id BIGINT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS system_configs (
			key VARCHAR(100) PRIMARY KEY,
			value JSONB NOT NULL DEFAULT '{}',
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		// Default configs
		`INSERT INTO system_configs (key, value) VALUES
			('sla_rules', '{"p0":{"response_minutes":15,"resolve_hours":4},"p1":{"response_minutes":30,"resolve_hours":8},"p2":{"response_minutes":60,"resolve_hours":24},"p3":{"response_minutes":120,"resolve_hours":72}}'::jsonb),
			('score_weights', '{"response_timeliness":30,"processing_efficiency":25,"sla_compliance":20,"ticket_quality":15,"knowledge_contribution":10}'::jsonb),
			('notification_settings', '{"ticket_assigned":true,"status_changed":true,"sla_warning":true,"daily_summary":false}'::jsonb)
			ON CONFLICT (key) DO NOTHING`,
		// Default categories
		`INSERT INTO kb_categories (name, sort_order) VALUES
			('网络运维', 1), ('服务器管理', 2), ('数据库运维', 3), ('故障排查', 4), ('安全管理', 5),
			('操作手册', 6), ('最佳实践', 7), ('常见问题', 8)
			ON CONFLICT DO NOTHING`,
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_assignee ON tickets(assignee_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_creator ON tickets(creator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_project ON tickets(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ticket_logs_ticket ON ticket_logs(ticket_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ticket_files_ticket ON ticket_files(ticket_id)`,
		`CREATE INDEX IF NOT EXISTS idx_scores_user_period ON scores(user_id, period)`,
		`CREATE INDEX IF NOT EXISTS idx_kb_articles_category ON kb_articles(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_kb_files_article ON kb_files(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, is_read)`,
		`CREATE INDEX IF NOT EXISTS idx_project_members_project ON project_members(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_rect_project ON project_rectifications(project_id)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("迁移失败: %w\nSQL: %s", err, m[:min(100, len(m))])
		}
	}
	return nil
}
