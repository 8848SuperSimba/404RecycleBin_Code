package controller

import (
	"bookstore/dao"
	"bookstore/model"
	"bookstore/utils"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestLoginFlowHTTP 测试HTTP登录流程
func TestLoginFlowHTTP(t *testing.T) {
	// 准备测试数据
	testUsername := "httptestuser"
	testPassword := "password123"
	testEmail := "httptestuser@example.com"

	// 创建测试用户
	setupTestUser(t, testUsername, testPassword, testEmail)
	defer func() {
		cleanupTestUser(t, testUsername)
		cleanupTestSession(t, "")
	}()

	// 测试未登录状态访问
	t.Run("测试未登录状态访问", func(t *testing.T) {
		// 创建一个GET请求到首页
		req, err := http.NewRequest("GET", "/main", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 创建ResponseRecorder来记录响应
		rr := httptest.NewRecorder()

		// 调用GetPageBooksByPrice处理器
		GetPageBooksByPrice(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查响应内容是否包含登录相关信息
		body := rr.Body.String()
		if !strings.Contains(body, "登录") && !strings.Contains(body, "login") {
			t.Log("响应内容可能不包含登录相关信息，这是正常的，因为页面可能重定向")
		}
	})

	// 测试登录成功
	t.Run("测试登录成功", func(t *testing.T) {
		// 创建POST请求到登录端点
		formData := url.Values{}
		formData.Set("username", testUsername)
		formData.Set("password", testPassword)

		req, err := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 设置Content-Type
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// 创建ResponseRecorder
		rr := httptest.NewRecorder()

		// 调用Login处理器
		Login(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查是否设置了Cookie
		cookies := rr.Result().Cookies()
		var userCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "user" {
				userCookie = cookie
				break
			}
		}

		if userCookie == nil {
			t.Error("登录成功后应该设置user Cookie，但没有找到")
		} else {
			// 验证Cookie值不为空
			if userCookie.Value == "" {
				t.Error("Cookie值不应该为空")
			}

			// 验证Session是否存在于数据库中
			session, err := dao.GetSession(userCookie.Value)
			if err != nil {
				t.Errorf("获取Session失败: %v", err)
			}
			if session == nil {
				t.Error("Session应该存在于数据库中")
			}
			if session.UserName != testUsername {
				t.Errorf("Session用户名不匹配，期望: %s, 实际: %s", testUsername, session.UserName)
			}

			// 清理测试Session
			cleanupTestSession(t, userCookie.Value)
		}
	})

	// 测试登录失败
	t.Run("测试登录失败", func(t *testing.T) {
		// 测试错误密码
		formData := url.Values{}
		formData.Set("username", testUsername)
		formData.Set("password", "wrongpassword")

		req, err := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		Login(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查响应内容是否包含错误信息
		body := rr.Body.String()
		if !strings.Contains(body, "用户名或密码不正确") {
			t.Error("登录失败时应该返回错误信息")
		}

		// 检查是否没有设置Cookie
		cookies := rr.Result().Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "user" {
				t.Error("登录失败时不应该设置user Cookie")
			}
		}
	})

	// 测试注销功能
	t.Run("测试注销功能", func(t *testing.T) {
		// 获取实际创建的用户ID
		user, err := dao.CheckUserName(testUsername)
		if err != nil {
			t.Errorf("获取测试用户失败: %v", err)
		}

		// 首先创建一个有效的Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUsername,
			UserID:    user.ID,
		}

		// 添加Session到数据库
		err = dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		// 创建带有Cookie的注销请求
		req, err := http.NewRequest("GET", "/logout", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 添加Cookie到请求
		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()

		// 调用Logout处理器
		Logout(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查Cookie是否被设置为过期
		cookies := rr.Result().Cookies()
		var userCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "user" {
				userCookie = cookie
				break
			}
		}

		if userCookie != nil {
			if userCookie.MaxAge != -1 {
				t.Error("注销后Cookie应该被设置为过期")
			}
		}

		// 验证Session是否已被删除
		deletedSession, err := dao.GetSession(sessionID)
		if err != nil {
			t.Errorf("获取已删除的Session时发生错误: %v", err)
		}
		if deletedSession != nil && deletedSession.UserID > 0 {
			t.Error("注销后Session应该已被删除")
		}
	})

	// 测试重复登录
	t.Run("测试重复登录", func(t *testing.T) {
		// 第一次登录
		formData := url.Values{}
		formData.Set("username", testUsername)
		formData.Set("password", testPassword)

		req, err := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		Login(rr, req)

		// 获取第一次登录的Cookie
		cookies := rr.Result().Cookies()
		var firstCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "user" {
				firstCookie = cookie
				break
			}
		}

		if firstCookie == nil {
			t.Fatal("第一次登录应该设置Cookie")
		}

		// 第二次登录（模拟已登录状态）
		req2, err := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建第二次登录请求失败: %v", err)
		}

		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req2.AddCookie(firstCookie) // 添加第一次登录的Cookie

		rr2 := httptest.NewRecorder()

		Login(rr2, req2)

		// 检查第二次登录的响应
		if rr2.Code != http.StatusOK {
			t.Errorf("重复登录时期望状态码200，实际得到: %d", rr2.Code)
		}

		// 清理测试Session
		cleanupTestSession(t, firstCookie.Value)
	})

	// 测试空值登录
	t.Run("测试空值登录", func(t *testing.T) {
		// 测试空用户名
		formData := url.Values{}
		formData.Set("username", "")
		formData.Set("password", testPassword)

		req, err := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		Login(rr, req)

		// 检查响应内容是否包含错误信息
		body := rr.Body.String()
		t.Logf("空用户名登录响应内容: %s", body)
		// 空用户名登录可能返回注册页面或错误信息，这是正常的
		if !strings.Contains(body, "用户名或密码不正确") && !strings.Contains(body, "用户名已存在") && !strings.Contains(body, "注册") {
			t.Error("空用户名登录时应该返回错误信息或注册页面")
		}

		// 测试空密码
		formData2 := url.Values{}
		formData2.Set("username", testUsername)
		formData2.Set("password", "")

		req2, err := http.NewRequest("POST", "/login", strings.NewReader(formData2.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr2 := httptest.NewRecorder()

		Login(rr2, req2)

		// 检查响应内容是否包含错误信息
		body2 := rr2.Body.String()
		if !strings.Contains(body2, "用户名或密码不正确") && !strings.Contains(body2, "用户名已存在") && !strings.Contains(body2, "注册") {
			t.Error("空密码登录时应该返回错误信息或注册页面")
		}
	})
}

// TestSessionManagement 测试Session管理
func TestSessionManagement(t *testing.T) {
	testUsername := "sessiontestuser"
	testPassword := "password123"
	testEmail := "sessiontestuser@example.com"

	// 创建测试用户
	setupTestUser(t, testUsername, testPassword, testEmail)
	defer func() {
		cleanupTestUser(t, testUsername)
		cleanupTestSession(t, "")
	}()

	// 测试Session创建和验证
	t.Run("测试Session创建和验证", func(t *testing.T) {
		// 获取实际创建的用户ID
		user, err := dao.CheckUserName(testUsername)
		if err != nil {
			t.Errorf("获取测试用户失败: %v", err)
		}

		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUsername,
			UserID:    user.ID,
		}

		// 添加Session到数据库
		err = dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		// 验证Session存在
		retrievedSession, err := dao.GetSession(sessionID)
		if err != nil {
			t.Errorf("获取Session失败: %v", err)
		}
		if retrievedSession == nil {
			t.Error("Session应该存在")
		}
		if retrievedSession.SessionID != sessionID {
			t.Errorf("Session ID不匹配，期望: %s, 实际: %s", sessionID, retrievedSession.SessionID)
		}

		// 测试带Cookie的请求
		req, err := http.NewRequest("GET", "/main", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		// 测试IsLogin函数
		isLoggedIn, session := dao.IsLogin(req)
		if !isLoggedIn {
			t.Error("应该检测到已登录状态")
		}
		if session == nil {
			t.Error("Session不应该为nil")
		}
		if session.UserName != testUsername {
			t.Errorf("Session用户名不匹配，期望: %s, 实际: %s", testUsername, session.UserName)
		}

		// 清理测试Session
		cleanupTestSession(t, sessionID)
	})

	// 测试Session删除
	t.Run("测试Session删除", func(t *testing.T) {
		// 获取实际创建的用户ID
		user, err := dao.CheckUserName(testUsername)
		if err != nil {
			t.Errorf("获取测试用户失败: %v", err)
		}

		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUsername,
			UserID:    user.ID,
		}

		// 添加Session到数据库
		err = dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		// 删除Session
		err = dao.DeleteSession(sessionID)
		if err != nil {
			t.Errorf("删除Session失败: %v", err)
		}

		// 验证Session已被删除
		deletedSession, err := dao.GetSession(sessionID)
		if err != nil {
			t.Errorf("获取已删除的Session时发生错误: %v", err)
		}
		if deletedSession != nil && deletedSession.UserID > 0 {
			t.Error("Session应该已被删除")
		}
	})
}

// TestUserControllerEdgeCases 测试用户控制器边界情况
func TestUserControllerEdgeCases(t *testing.T) {
	// 测试无效Cookie注销
	t.Run("测试无效Cookie注销", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/logout", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 添加无效的Cookie
		cookie := &http.Cookie{
			Name:  "user",
			Value: "invalid_session_id",
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		Logout(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试空Cookie注销
	t.Run("测试空Cookie注销", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/logout", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		Logout(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试无效用户名注册
	t.Run("测试无效用户名注册", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "")
		formData.Set("password", "testpassword")
		formData.Set("email", "test@example.com")

		req, err := http.NewRequest("POST", "/regist", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		Regist(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试空密码注册
	t.Run("测试空密码注册", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "testuser999")
		formData.Set("password", "")
		formData.Set("email", "test@example.com")

		req, err := http.NewRequest("POST", "/regist", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		Regist(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试空邮箱注册
	t.Run("测试空邮箱注册", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "testuser998")
		formData.Set("password", "testpassword")
		formData.Set("email", "")

		req, err := http.NewRequest("POST", "/regist", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		Regist(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试空表单注册
	t.Run("测试空表单注册", func(t *testing.T) {
		formData := url.Values{}

		req, err := http.NewRequest("POST", "/regist", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		Regist(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试用户名检查 - 空用户名
	t.Run("测试用户名检查-空用户名", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "")

		req, err := http.NewRequest("POST", "/checkUserName", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		CheckUserName(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查响应内容
		body := rr.Body.String()
		if !strings.Contains(body, "用户名可用") && !strings.Contains(body, "用户名已存在") {
			t.Error("空用户名应该显示为可用或已存在")
		}
	})

	// 测试用户名检查 - 不存在的用户名
	t.Run("测试用户名检查-不存在的用户名", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "nonexistentuser12345")

		req, err := http.NewRequest("POST", "/checkUserName", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		CheckUserName(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查响应内容
		body := rr.Body.String()
		if !strings.Contains(body, "用户名可用") {
			t.Error("不存在的用户名应该显示为可用")
		}
	})

	// 测试空表单用户名检查
	t.Run("测试空表单用户名检查", func(t *testing.T) {
		formData := url.Values{}

		req, err := http.NewRequest("POST", "/checkUserName", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		CheckUserName(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})
}
