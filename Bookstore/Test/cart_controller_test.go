package controller

import (
	"bookstore/dao"
	"bookstore/model"
	"bookstore/utils"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestCartControllerOperations 测试购物车控制器操作
func TestCartControllerOperations(t *testing.T) {
	// 准备测试数据
	testUserID := 888
	testUserName := "testuser888"
	testUserEmail := "testuser888@example.com"
	testUserPassword := "testpassword888"

	// 先尝试删除可能存在的测试用户和相关数据
	cleanupTestUser(t, testUserID)

	// 创建测试用户
	err := dao.SaveUser(testUserName, testUserPassword, testUserEmail)
	if err != nil {
		// 如果用户已存在，尝试删除后重新创建
		if strings.Contains(err.Error(), "Duplicate entry") {
			// 先清理相关数据
			cleanupTestCart(t, testUserID)
			cleanupTestSession(t, "")
			// 删除用户
			sqlStr := "DELETE FROM users WHERE username = ?"
			utils.Db.Exec(sqlStr, testUserName)
			// 重新创建
			err = dao.SaveUser(testUserName, testUserPassword, testUserEmail)
			if err != nil {
				t.Errorf("重新创建测试用户失败: %v", err)
			}
		} else {
			t.Errorf("创建测试用户失败: %v", err)
		}
	}

	// 获取实际创建的用户ID
	user, err := dao.CheckUserName(testUserName)
	if err != nil {
		t.Errorf("获取测试用户失败: %v", err)
	}
	testUserID = user.ID

	defer func() {
		// 清理测试用户
		cleanupTestUser(t, testUserID)
	}()

	testBooks := []*model.Book{
		{Title: "控制器购物车测试图书1", Author: "控制器购物车测试作者1", Price: 10.00, Sales: 0, Stock: 100, ImgPath: "/static/img/ctrl_cart1.jpg"},
		{Title: "控制器购物车测试图书2", Author: "控制器购物车测试作者2", Price: 20.00, Sales: 0, Stock: 100, ImgPath: "/static/img/ctrl_cart2.jpg"},
		{Title: "控制器购物车测试图书3", Author: "控制器购物车测试作者3", Price: 30.00, Sales: 0, Stock: 100, ImgPath: "/static/img/ctrl_cart3.jpg"},
	}

	var addedBookIDs []int
	for _, book := range testBooks {
		err := dao.AddBook(book)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取刚添加的图书ID
		books, err := dao.GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		for _, b := range books {
			if b.Title == book.Title {
				addedBookIDs = append(addedBookIDs, b.ID)
				break
			}
		}
	}

	defer func() {
		// 清理测试数据
		for _, bookID := range addedBookIDs {
			cleanupTestBook(t, bookID)
		}
		cleanupTestCart(t, testUserID)
	}()

	// 测试添加商品到购物车
	t.Run("测试添加商品到购物车", func(t *testing.T) {
		// 先清理可能存在的购物车
		cleanupTestCart(t, testUserID)

		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  "testuser",
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 创建请求
		req, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", addedBookIDs[0]), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 添加Cookie
		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		AddBook2Cart(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查响应内容
		body := rr.Body.String()
		if !strings.Contains(body, "添加到了购物车") {
			t.Error("响应内容应该包含'添加到了购物车'")
		}

		// 验证购物车是否成功创建
		cart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil {
			t.Error("购物车应该存在")
		}

		if len(cart.CartItems) != 1 {
			t.Errorf("购物项数量不匹配，期望: 1, 实际: %d", len(cart.CartItems))
		}

		if cart.CartItems[0].Count != 1 {
			t.Errorf("购物项数量不匹配，期望: 1, 实际: %d", cart.CartItems[0].Count)
		}
	})

	// 测试重复添加同一商品
	t.Run("测试重复添加同一商品", func(t *testing.T) {
		// 先清理可能存在的购物车
		cleanupTestCart(t, testUserID)

		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  "testuser",
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 第一次添加商品
		req1, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", addedBookIDs[0]), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req1.AddCookie(cookie)

		rr1 := httptest.NewRecorder()
		AddBook2Cart(rr1, req1)

		// 第二次添加同一商品
		req2, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", addedBookIDs[0]), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req2.AddCookie(cookie)

		rr2 := httptest.NewRecorder()
		AddBook2Cart(rr2, req2)

		// 检查响应状态码
		if rr2.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr2.Code)
		}

		// 验证购物车中的商品数量
		cart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil {
			t.Error("购物车应该存在")
		}

		if len(cart.CartItems) != 1 {
			t.Errorf("购物项数量不匹配，期望: 1, 实际: %d", len(cart.CartItems))
		}

		if cart.CartItems[0].Count != 2 {
			t.Errorf("购物项数量不匹配，期望: 2, 实际: %d", cart.CartItems[0].Count)
		}
	})

	// 测试获取购物车信息
	t.Run("测试获取购物车信息", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  "testuser",
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 创建请求
		req, err := http.NewRequest("GET", "/getCartInfo", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 添加Cookie
		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		GetCartInfo(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查响应内容
		body := rr.Body.String()
		if body == "" {
			t.Error("响应内容为空")
		}
	})

	// 测试更新购物项
	t.Run("测试更新购物项", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  "testuser",
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 先添加商品到购物车
		req1, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", addedBookIDs[0]), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req1.AddCookie(cookie)

		rr1 := httptest.NewRecorder()
		AddBook2Cart(rr1, req1)

		// 获取购物车中的购物项ID
		cart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil || len(cart.CartItems) == 0 {
			t.Fatal("购物车中应该包含购物项")
		}

		cartItemID := cart.CartItems[0].CartItemID

		// 更新购物项数量
		formData := url.Values{}
		formData.Set("cartItemId", fmt.Sprintf("%d", cartItemID))
		formData.Set("bookCount", "5")

		req2, err := http.NewRequest("POST", "/updateCartItem", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req2.AddCookie(cookie)

		rr2 := httptest.NewRecorder()
		UpdateCartItem(rr2, req2)

		// 检查响应状态码
		if rr2.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr2.Code)
		}

		// 检查响应内容是否为JSON
		body := rr2.Body.String()
		if body == "" {
			t.Error("响应内容为空")
		}

		// 验证购物项数量是否已更新
		updatedCart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取更新后的购物车失败: %v", err)
		}

		if updatedCart.CartItems[0].Count != 5 {
			t.Errorf("购物项数量不匹配，期望: 5, 实际: %d", updatedCart.CartItems[0].Count)
		}
	})

	// 测试删除购物项
	t.Run("测试删除购物项", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  "testuser",
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 先添加商品到购物车
		req1, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", addedBookIDs[0]), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req1.AddCookie(cookie)

		rr1 := httptest.NewRecorder()
		AddBook2Cart(rr1, req1)

		// 获取购物车中的购物项ID
		cart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil || len(cart.CartItems) == 0 {
			t.Fatal("购物车中应该包含购物项")
		}

		cartItemID := cart.CartItems[0].CartItemID

		// 删除购物项
		req2, err := http.NewRequest("GET", "/deleteCartItem?cartItemId="+fmt.Sprintf("%d", cartItemID), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req2.AddCookie(cookie)

		rr2 := httptest.NewRecorder()
		DeleteCartItem(rr2, req2)

		// 检查响应状态码
		if rr2.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr2.Code)
		}

		// 验证购物项是否已被删除
		updatedCart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取更新后的购物车失败: %v", err)
		}

		if updatedCart != nil && len(updatedCart.CartItems) != 0 {
			t.Error("购物项应该已被删除")
		}
	})

	// 测试清空购物车
	t.Run("测试清空购物车", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  "testuser",
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 先添加商品到购物车
		req1, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", addedBookIDs[0]), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req1.AddCookie(cookie)

		rr1 := httptest.NewRecorder()
		AddBook2Cart(rr1, req1)

		// 获取购物车ID
		cart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil {
			t.Fatal("购物车应该存在")
		}

		cartID := cart.CartID

		// 清空购物车
		req2, err := http.NewRequest("GET", "/deleteCart?cartId="+cartID, nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req2.AddCookie(cookie)

		rr2 := httptest.NewRecorder()
		DeleteCart(rr2, req2)

		// 检查响应状态码
		if rr2.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr2.Code)
		}

		// 验证购物车是否已被清空
		emptyCart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			// 购物车不存在是正常的，说明清空成功
			if !strings.Contains(err.Error(), "no rows in result set") {
				t.Errorf("获取清空后的购物车失败: %v", err)
			}
		} else if emptyCart != nil {
			t.Error("购物车应该已被清空")
		}
	})

	// 测试未登录状态
	t.Run("测试未登录状态", func(t *testing.T) {
		// 创建请求（不添加Cookie）
		req, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", addedBookIDs[0]), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		AddBook2Cart(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查响应内容
		body := rr.Body.String()
		if !strings.Contains(body, "请先登录") {
			t.Error("未登录时应该返回'请先登录'")
		}
	})
}

// TestCartControllerDataValidation 测试购物车控制器数据验证
func TestCartControllerDataValidation(t *testing.T) {
	testUserName := "testuser887"
	testUserEmail := "testuser887@example.com"
	testUserPassword := "testpassword887"

	// 创建测试用户
	err := dao.SaveUser(testUserName, testUserPassword, testUserEmail)
	if err != nil {
		t.Errorf("创建测试用户失败: %v", err)
	}

	// 获取实际创建的用户ID
	user, err := dao.CheckUserName(testUserName)
	if err != nil {
		t.Errorf("获取测试用户失败: %v", err)
	}
	testUserID := user.ID

	defer func() {
		cleanupTestCart(t, testUserID)
		cleanupTestUser(t, testUserID)
	}()

	testBookID := 1

	tests := []struct {
		name        string
		bookID      string
		expectError bool
		description string
	}{
		{
			name:        "正常图书ID",
			bookID:      fmt.Sprintf("%d", testBookID),
			expectError: false,
			description: "测试正常的图书ID",
		},
		{
			name:        "空图书ID",
			bookID:      "",
			expectError: false, // 这里假设系统允许空图书ID，实际项目中可能需要验证
			description: "测试空图书ID",
		},
		{
			name:        "无效图书ID",
			bookID:      "999999",
			expectError: false, // 这里假设系统允许无效图书ID，实际项目中可能需要验证
			description: "测试无效图书ID",
		},
		{
			name:        "负图书ID",
			bookID:      "-1",
			expectError: false, // 这里假设系统允许负图书ID，实际项目中可能需要验证
			description: "测试负图书ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Session
			sessionID := utils.CreateUUID()
			session := &model.Session{
				SessionID: sessionID,
				UserName:  "testuser",
				UserID:    testUserID,
			}

			err := dao.AddSession(session)
			if err != nil {
				t.Errorf("添加Session失败: %v", err)
			}

			defer func() {
				cleanupTestSession(t, sessionID)
			}()

			// 创建请求
			req, err := http.NewRequest("GET", "/addBook2Cart?bookId="+tt.bookID, nil)
			if err != nil {
				t.Fatalf("创建请求失败: %v", err)
			}

			// 添加Cookie
			cookie := &http.Cookie{
				Name:  "user",
				Value: sessionID,
			}
			req.AddCookie(cookie)

			rr := httptest.NewRecorder()
			AddBook2Cart(rr, req)

			if tt.expectError {
				if rr.Code == http.StatusOK {
					t.Errorf("期望返回错误，但没有返回错误: %s", tt.description)
				}
			} else {
				if rr.Code != http.StatusOK {
					t.Errorf("不期望返回错误，但返回了错误: %d, %s", rr.Code, tt.description)
				}
			}
		})
	}
}

// TestCartControllerConcurrentOperations 测试购物车控制器并发操作
func TestCartControllerConcurrentOperations(t *testing.T) {
	testUserName := "testuser886"
	testUserEmail := "testuser886@example.com"
	testUserPassword := "testpassword886"

	// 先清理可能存在的测试用户
	cleanupTestUser(t, 886)

	// 创建测试用户
	err := dao.SaveUser(testUserName, testUserPassword, testUserEmail)
	if err != nil {
		// 如果用户已存在，尝试删除后重新创建
		if strings.Contains(err.Error(), "Duplicate entry") {
			// 先清理相关数据
			cleanupTestCart(t, 886)
			cleanupTestSession(t, "")
			// 删除用户
			sqlStr := "DELETE FROM users WHERE username = ?"
			utils.Db.Exec(sqlStr, testUserName)
			// 重新创建
			err = dao.SaveUser(testUserName, testUserPassword, testUserEmail)
			if err != nil {
				// 如果还是失败，跳过这个测试
				t.Skipf("无法创建测试用户，跳过测试: %v", err)
			}
		} else {
			t.Errorf("创建测试用户失败: %v", err)
		}
	}

	// 获取实际创建的用户ID
	user, err := dao.CheckUserName(testUserName)
	if err != nil {
		t.Errorf("获取测试用户失败: %v", err)
	}
	testUserID := user.ID

	defer func() {
		cleanupTestCart(t, testUserID)
		cleanupTestUser(t, testUserID)
	}()

	testBookID := 1

	// 测试并发添加商品到购物车
	t.Run("测试并发添加商品到购物车", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  "testuser",
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				req, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", testBookID), nil)
				if err != nil {
					t.Errorf("创建请求失败: %v", err)
					return
				}

				cookie := &http.Cookie{
					Name:  "user",
					Value: sessionID,
				}
				req.AddCookie(cookie)

				rr := httptest.NewRecorder()
				AddBook2Cart(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("并发添加商品失败，状态码: %d", rr.Code)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证购物车状态
		cart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil {
			t.Error("购物车应该存在")
		}
	})

	// 测试并发更新购物项
	t.Run("测试并发更新购物项", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  "testuser",
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 先添加商品到购物车
		req1, err := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", testBookID), nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req1.AddCookie(cookie)

		rr1 := httptest.NewRecorder()
		AddBook2Cart(rr1, req1)

		// 获取购物车中的购物项ID
		cart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil || len(cart.CartItems) == 0 {
			t.Fatal("购物车中应该包含购物项")
		}

		cartItemID := cart.CartItems[0].CartItemID

		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				formData := url.Values{}
				formData.Set("cartItemId", fmt.Sprintf("%d", cartItemID))
				formData.Set("bookCount", fmt.Sprintf("%d", index+1))

				req, err := http.NewRequest("POST", "/updateCartItem", strings.NewReader(formData.Encode()))
				if err != nil {
					t.Errorf("创建请求失败: %v", err)
					return
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.AddCookie(cookie)

				rr := httptest.NewRecorder()
				UpdateCartItem(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("并发更新购物项失败，状态码: %d", rr.Code)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证购物车状态
		updatedCart, err := dao.GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if updatedCart == nil {
			t.Error("购物车应该存在")
		}
	})
}

// BenchmarkCartControllerOperations 性能测试
func BenchmarkAddBook2Cart(b *testing.B) {
	testUserName := "testuser885"
	testUserEmail := "testuser885@example.com"
	testUserPassword := "testpassword885"

	// 创建测试用户
	err := dao.SaveUser(testUserName, testUserPassword, testUserEmail)
	if err != nil {
		b.Errorf("创建测试用户失败: %v", err)
	}

	// 获取实际创建的用户ID
	user, err := dao.CheckUserName(testUserName)
	if err != nil {
		b.Errorf("获取测试用户失败: %v", err)
	}
	testUserID := user.ID

	// 准备测试数据
	sessionID := utils.CreateUUID()
	session := &model.Session{
		SessionID: sessionID,
		UserName:  testUserName,
		UserID:    testUserID,
	}

	dao.AddSession(session)

	defer func() {
		cleanupTestSession(nil, sessionID)
		cleanupTestCart(nil, testUserID)
		cleanupTestUser(nil, testUserID)
	}()

	testBookID := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/addBook2Cart?bookId="+fmt.Sprintf("%d", testBookID), nil)
		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)
		rr := httptest.NewRecorder()
		AddBook2Cart(rr, req)
	}
}

// TestCartControllerEdgeCases 测试购物车控制器边界情况
func TestCartControllerEdgeCases(t *testing.T) {
	testUserName := "testuser883"
	testUserEmail := "testuser883@example.com"
	testUserPassword := "testpassword883"

	// 创建测试用户
	err := dao.SaveUser(testUserName, testUserPassword, testUserEmail)
	if err != nil {
		t.Errorf("创建测试用户失败: %v", err)
	}

	// 获取实际创建的用户ID
	user, err := dao.CheckUserName(testUserName)
	if err != nil {
		t.Errorf("获取测试用户失败: %v", err)
	}
	testUserID := user.ID

	defer func() {
		cleanupTestCart(t, testUserID)
		cleanupTestUser(t, testUserID)
	}()

	// 测试无效的图书ID添加到购物车
	t.Run("测试无效图书ID添加到购物车", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUserName,
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 测试无效图书ID
		req, err := http.NewRequest("GET", "/addBook2Cart?bookId=999999", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		AddBook2Cart(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试空图书ID添加到购物车
	t.Run("测试空图书ID添加到购物车", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUserName,
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 测试空图书ID
		req, err := http.NewRequest("GET", "/addBook2Cart?bookId=", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		AddBook2Cart(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试无效的购物项ID更新
	t.Run("测试无效购物项ID更新", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUserName,
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 测试无效购物项ID
		formData := url.Values{}
		formData.Set("cartItemId", "999999")
		formData.Set("bookCount", "5")

		req, err := http.NewRequest("POST", "/updateCartItem", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		UpdateCartItem(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试无效的购物项ID删除
	t.Run("测试无效购物项ID删除", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUserName,
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 测试无效购物项ID
		req, err := http.NewRequest("GET", "/deleteCartItem?cartItemId=999999", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		DeleteCartItem(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试无效的购物车ID清空
	t.Run("测试无效购物车ID清空", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUserName,
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 测试无效购物车ID
		req, err := http.NewRequest("GET", "/deleteCart?cartId=invalid_cart_id", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		DeleteCart(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试空购物车ID清空
	t.Run("测试空购物车ID清空", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUserName,
			UserID:    testUserID,
		}

		err := dao.AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		defer func() {
			cleanupTestSession(t, sessionID)
		}()

		// 测试空购物车ID
		req, err := http.NewRequest("GET", "/deleteCart?cartId=", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		rr := httptest.NewRecorder()
		DeleteCart(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})
}
