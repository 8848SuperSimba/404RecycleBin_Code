package dao

import (
	"bookstore/model"
	"bookstore/utils"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// fmt.Println("测试bookdao中的方法")
	m.Run()
}

func TestUser(t *testing.T) {
	// fmt.Println("测试userdao中的函数")
	// t.Run("验证用户名或密码：", testLogin)
	// t.Run("验证用户名：", testRegist)
	// t.Run("保存用户：", testSave)
}

// TestSaveUser 测试用户注册功能
func TestSaveUser(t *testing.T) {
	// 测试正常注册
	t.Run("测试正常注册", func(t *testing.T) {
		// 清理可能存在的测试数据
		cleanupTestUser(t, "testuser1")

		// 测试正常注册
		err := SaveUser("testuser1", "password123", "testuser1@example.com")
		if err != nil {
			t.Errorf("正常注册失败: %v", err)
		}

		// 验证用户是否成功创建
		user, err := CheckUserName("testuser1")
		if err != nil {
			t.Errorf("查询用户失败: %v", err)
		}
		if user.ID == 0 {
			t.Error("用户未成功创建")
		}
		if user.Username != "testuser1" {
			t.Errorf("用户名不匹配，期望: testuser1, 实际: %s", user.Username)
		}
		if user.Email != "testuser1@example.com" {
			t.Errorf("邮箱不匹配，期望: testuser1@example.com, 实际: %s", user.Email)
		}

		// 清理测试数据
		cleanupTestUser(t, "testuser1")
	})

	// 测试用户名重复
	t.Run("测试用户名重复", func(t *testing.T) {
		// 先创建一个用户
		err := SaveUser("duplicateuser", "password123", "duplicateuser@example.com")
		if err != nil {
			t.Errorf("创建初始用户失败: %v", err)
		}

		// 尝试用相同用户名创建用户
		err = SaveUser("duplicateuser", "password456", "duplicateuser2@example.com")
		if err == nil {
			t.Error("应该返回用户名重复错误，但没有返回错误")
		}

		// 清理测试数据
		cleanupTestUser(t, "duplicateuser")
	})

	// 测试邮箱重复
	t.Run("测试邮箱重复", func(t *testing.T) {
		// 使用时间戳确保唯一性
		timestamp := time.Now().UnixNano()
		username1 := fmt.Sprintf("user1_%d", timestamp)
		username2 := fmt.Sprintf("user2_%d", timestamp)
		email := fmt.Sprintf("duplicateemail_%d@example.com", timestamp)

		// 先创建一个用户
		err := SaveUser(username1, "password123", email)
		if err != nil {
			t.Errorf("创建初始用户失败: %v", err)
		}

		// 尝试用相同邮箱创建用户
		err = SaveUser(username2, "password456", email)
		if err == nil {
			t.Error("应该返回邮箱重复错误，但没有返回错误")
		}

		// 清理测试数据
		cleanupTestUser(t, username1)
		cleanupTestUser(t, username2)
	})

	// 测试空值处理
	t.Run("测试空值处理", func(t *testing.T) {
		// 测试空用户名
		err := SaveUser("", "password123", "test@example.com")
		if err == nil {
			t.Error("空用户名应该返回错误")
		}

		// 测试空密码
		err = SaveUser("testuser", "", "test@example.com")
		if err == nil {
			t.Error("空密码应该返回错误")
		}

		// 测试空邮箱
		err = SaveUser("testuser", "password123", "")
		if err == nil {
			t.Error("空邮箱应该返回错误")
		}
	})
}

// TestCheckUserNameAndPassword 测试用户登录验证功能
func TestCheckUserNameAndPassword(t *testing.T) {
	// 准备测试数据
	setupTestUser(t, "logintest", "password123", "logintest@example.com")
	defer cleanupTestUser(t, "logintest")

	// 测试正确用户名密码
	t.Run("测试正确用户名密码", func(t *testing.T) {
		user, err := CheckUserNameAndPassword("logintest", "password123")
		if err != nil {
			t.Errorf("查询用户失败: %v", err)
		}
		if user.ID == 0 {
			t.Error("应该找到用户，但用户ID为0")
		}
		if user.Username != "logintest" {
			t.Errorf("用户名不匹配，期望: logintest, 实际: %s", user.Username)
		}
		if user.Password != "password123" {
			t.Errorf("密码不匹配，期望: password123, 实际: %s", user.Password)
		}
	})

	// 测试错误密码
	t.Run("测试错误密码", func(t *testing.T) {
		user, _ := CheckUserNameAndPassword("logintest", "wrongpassword")
		// DAO may return empty user when not found; assert not found
		if user != nil && user.ID != 0 && user.Username != "" {
			t.Error("错误密码不应该找到用户，但找到了用户")
		}
	})

	// 测试不存在的用户名
	t.Run("测试不存在的用户名", func(t *testing.T) {
		user, _ := CheckUserNameAndPassword("nonexistentuser", "password123")
		if user != nil && user.ID != 0 && user.Username != "" {
			t.Error("不存在的用户名不应该找到用户，但找到了用户")
		}
	})

	// 测试空用户名
	t.Run("测试空用户名", func(t *testing.T) {
		user, _ := CheckUserNameAndPassword("", "password123")
		if user != nil && user.ID != 0 && user.Username != "" {
			t.Error("空用户名不应该找到用户，但找到了用户")
		}
	})

	// 测试空密码
	t.Run("测试空密码", func(t *testing.T) {
		user, _ := CheckUserNameAndPassword("logintest", "")
		if user != nil && user.ID != 0 && user.Username != "" {
			t.Error("空密码不应该找到用户，但找到了用户")
		}
	})
}

// setupTestUser 创建测试用户
func setupTestUser(t *testing.T, username, password, email string) {
	err := SaveUser(username, password, email)
	if err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
}

// TestLoginFlow 测试登录流程
func TestLoginFlow(t *testing.T) {
	// 准备测试数据
	testUsername := "logintestuser"
	testPassword := "password123"
	testEmail := "logintestuser@example.com"

	// 创建测试用户
	setupTestUser(t, testUsername, testPassword, testEmail)
	defer func() {
		cleanupTestUser(t, testUsername)
		cleanupTestSession(t, testUsername)
	}()

	// 测试未登录状态访问
	t.Run("测试未登录状态访问", func(t *testing.T) {
		// 创建一个模拟的HTTP请求，不包含Cookie
		req, err := http.NewRequest("GET", "/main", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 测试IsLogin函数
		isLoggedIn, session := IsLogin(req)
		if isLoggedIn {
			t.Error("未登录状态应该返回false，但返回了true")
		}
		if session != nil {
			t.Error("未登录状态session应该为nil，但不为nil")
		}
	})

	// 测试登录成功
	t.Run("测试登录成功", func(t *testing.T) {
		// 测试用户名密码验证
		user, err := CheckUserNameAndPassword(testUsername, testPassword)
		if err != nil {
			t.Errorf("验证用户失败: %v", err)
		}
		if user.ID == 0 {
			t.Error("应该找到用户，但用户ID为0")
		}
		if user.Username != testUsername {
			t.Errorf("用户名不匹配，期望: %s, 实际: %s", testUsername, user.Username)
		}

		// 测试Session创建
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  user.Username,
			UserID:    user.ID,
		}

		// 添加Session到数据库
		err = AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		// 验证Session是否成功创建
		retrievedSession, err := GetSession(sessionID)
		if err != nil {
			t.Errorf("获取Session失败: %v", err)
		}
		if retrievedSession.SessionID != sessionID {
			t.Errorf("Session ID不匹配，期望: %s, 实际: %s", sessionID, retrievedSession.SessionID)
		}
		if retrievedSession.UserName != testUsername {
			t.Errorf("用户名不匹配，期望: %s, 实际: %s", testUsername, retrievedSession.UserName)
		}

		// 测试带Cookie的登录状态检查
		req, err := http.NewRequest("GET", "/main", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 添加Cookie到请求
		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		// 测试IsLogin函数
		isLoggedIn, session := IsLogin(req)
		if !isLoggedIn {
			t.Error("已登录状态应该返回true，但返回了false")
		}
		if session == nil {
			t.Error("已登录状态session不应该为nil")
		}
		if session.UserName != testUsername {
			t.Errorf("Session用户名不匹配，期望: %s, 实际: %s", testUsername, session.UserName)
		}

		// 清理测试Session
		cleanupTestSession(t, sessionID)
	})

	// 测试登录失败
	t.Run("测试登录失败", func(t *testing.T) {
		// 测试错误密码
		user, _ := CheckUserNameAndPassword(testUsername, "wrongpassword")
		// DAO returns empty user when not found; assert no user found
		if user.ID != 0 {
			t.Error("错误密码不应该找到用户，但找到了用户")
		}

		// 测试不存在的用户名
		user, _ = CheckUserNameAndPassword("nonexistentuser", testPassword)
		if user.ID != 0 {
			t.Error("不存在的用户名不应该找到用户，但找到了用户")
		}

		// 测试空用户名
		user, _ = CheckUserNameAndPassword("", testPassword)
		if user.ID != 0 {
			t.Error("空用户名不应该找到用户，但找到了用户")
		}

		// 测试空密码
		user, _ = CheckUserNameAndPassword(testUsername, "")
		if user.ID != 0 {
			t.Error("空密码不应该找到用户，但找到了用户")
		}
	})

	// 测试注销功能
	t.Run("测试注销功能", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUsername,
			UserID:    1, // 假设用户ID为1
		}

		// 添加Session到数据库
		err := AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		// 验证Session存在
		retrievedSession, err := GetSession(sessionID)
		if err != nil {
			t.Errorf("获取Session失败: %v", err)
		}
		if retrievedSession == nil {
			t.Error("Session应该存在，但为nil")
		}

		// 测试注销功能 - 删除Session
		err = DeleteSession(sessionID)
		if err != nil {
			t.Errorf("删除Session失败: %v", err)
		}

		// 验证Session已被删除
		deletedSession, err := GetSession(sessionID)
		if err != nil {
			t.Errorf("获取已删除的Session时发生错误: %v", err)
		}
		if deletedSession != nil && deletedSession.UserID > 0 {
			t.Error("Session应该已被删除，但仍然存在")
		}

		// 测试注销后的登录状态检查
		req, err := http.NewRequest("GET", "/main", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 添加已删除的Cookie到请求
		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		// 测试IsLogin函数
		isLoggedIn, session := IsLogin(req)
		if isLoggedIn {
			t.Error("注销后应该返回false，但返回了true")
		}
	})

	// 测试Session过期处理
	t.Run("测试Session过期处理", func(t *testing.T) {
		// 创建Session
		sessionID := utils.CreateUUID()
		session := &model.Session{
			SessionID: sessionID,
			UserName:  testUsername,
			UserID:    1,
		}

		// 添加Session到数据库
		err := AddSession(session)
		if err != nil {
			t.Errorf("添加Session失败: %v", err)
		}

		// 立即删除Session（模拟过期）
		err = DeleteSession(sessionID)
		if err != nil {
			t.Errorf("删除Session失败: %v", err)
		}

		// 测试带过期Cookie的请求
		req, err := http.NewRequest("GET", "/main", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		// 添加过期的Cookie
		cookie := &http.Cookie{
			Name:  "user",
			Value: sessionID,
		}
		req.AddCookie(cookie)

		// 测试IsLogin函数
		isLoggedIn, session := IsLogin(req)
		if isLoggedIn {
			t.Error("过期Session应该返回false，但返回了true")
		}
		if session != nil {
			t.Error("过期Session应该返回nil，但不为nil")
		}
	})

	// 测试并发登录
	t.Run("测试并发登录", func(t *testing.T) {
		// 这里可以添加并发测试逻辑
		// 使用goroutine同时进行登录操作
		// 验证系统在并发情况下的稳定性
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				// 测试用户名密码验证
				user, err := CheckUserNameAndPassword(testUsername, testPassword)
				if err != nil {
					t.Errorf("并发测试中验证用户失败: %v", err)
					return
				}
				if user.ID == 0 {
					t.Errorf("并发测试中应该找到用户，但用户ID为0")
					return
				}

				// 创建Session
				sessionID := utils.CreateUUID()
				session := &model.Session{
					SessionID: sessionID,
					UserName:  user.Username,
					UserID:    user.ID,
				}

				// 添加Session
				err = AddSession(session)
				if err != nil {
					t.Errorf("并发测试中添加Session失败: %v", err)
					return
				}

				// 立即清理Session
				cleanupTestSession(t, sessionID)
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// cleanupTestUser 清理测试用户
func cleanupTestUser(t *testing.T, username string) {
	// 先获取用户ID
	var userID int
	sqlStr := "SELECT id FROM users WHERE username = ?"
	err := utils.Db.QueryRow(sqlStr, username).Scan(&userID)
	if err != nil {
		t.Logf("获取用户ID失败: %v", err)
		return
	}

	// 删除相关的购物车项
	sqlStr = "DELETE FROM cart_items WHERE cart_id IN (SELECT cart_id FROM carts WHERE user_id = ?)"
	_, err = utils.Db.Exec(sqlStr, userID)
	if err != nil {
		t.Logf("删除购物车项失败: %v", err)
	}

	// 删除相关的购物车
	sqlStr = "DELETE FROM carts WHERE user_id = ?"
	_, err = utils.Db.Exec(sqlStr, userID)
	if err != nil {
		t.Logf("删除购物车失败: %v", err)
	}

	// 删除相关的订单项
	sqlStr = "DELETE FROM order_items WHERE order_id IN (SELECT id FROM orders WHERE user_id = ?)"
	_, err = utils.Db.Exec(sqlStr, userID)
	if err != nil {
		t.Logf("删除订单项失败: %v", err)
	}

	// 删除相关的订单
	sqlStr = "DELETE FROM orders WHERE user_id = ?"
	_, err = utils.Db.Exec(sqlStr, userID)
	if err != nil {
		t.Logf("删除订单失败: %v", err)
	}

	// 删除相关的Session
	sqlStr = "DELETE FROM sessions WHERE user_id = ?"
	_, err = utils.Db.Exec(sqlStr, userID)
	if err != nil {
		t.Logf("删除Session失败: %v", err)
	}

	// 最后删除用户
	sqlStr = "DELETE FROM users WHERE username = ?"
	_, err = utils.Db.Exec(sqlStr, username)
	if err != nil {
		t.Logf("清理测试用户失败: %v", err)
	}
}

// cleanupTestSession 清理测试Session
func cleanupTestSession(t *testing.T, sessionID string) {
	sqlStr := "DELETE FROM sessions WHERE session_id = ?"
	_, err := utils.Db.Exec(sqlStr, sessionID)
	if err != nil {
		t.Logf("清理测试Session失败: %v", err)
	}
}

func testLogin(t *testing.T) {
	user, _ := CheckUserNameAndPassword("admin", "123456")
	fmt.Println("获取用户信息是：", user)
}
func testRegist(t *testing.T) {
	user, _ := CheckUserName("admin")
	fmt.Println("获取用户信息是：", user)
}
func testSave(t *testing.T) {
	SaveUser("admin3", "123456", "admin@atguigu.com")
}

func TestBook(t *testing.T) {
	// fmt.Println("测试bookdao中的相关函数")
	// t.Run("测试获取所有图书", testGetBooks)
	// t.Run("测试添加图书", testAddBook)
	// t.Run("测试删除图书", testDeleteBook)
	// t.Run("测试获取一本图书", testGetBook)
	// t.Run("测试更新图书", testUpdateBook)
	// t.Run("测试获取带分页的图书", testGetPageBooks)
	// t.Run("测试获取带分页和价格范围的图书", testGetPageBooksByPrice)
}

func testGetBooks(t *testing.T) {
	books, _ := GetBooks()
	//遍历得到每一本图书
	for k, v := range books {
		fmt.Printf("第%v本图书的信息是：%v\n", k+1, v)
	}
}
func testAddBook(t *testing.T) {
	book := &model.Book{
		Title:   "三国演义",
		Author:  "罗贯中",
		Price:   88.88,
		Sales:   100,
		Stock:   100,
		ImgPath: "/static/img/default.jpg",
	}
	//调用添加图书的函数
	AddBook(book)
}
func testDeleteBook(t *testing.T) {
	//调用删除图书的函数
	DeleteBook("34")
}
func testGetBook(t *testing.T) {
	//调用获取图书的函数
	book, _ := GetBookByID("32")
	fmt.Println("获取的图书信息是：", book)
}
func testUpdateBook(t *testing.T) {
	book := &model.Book{
		ID:      32,
		Title:   "3个女人与105个男人的故事",
		Author:  "罗贯中",
		Price:   66.66,
		Sales:   10000,
		Stock:   1,
		ImgPath: "/static/img/default.jpg",
	}
	//调用更新图书的函数
	UpdateBook(book)
}

func testGetPageBooks(t *testing.T) {
	page, _ := GetPageBooks("9")
	fmt.Println("当前页是：", page.PageNo)
	fmt.Println("总页数是：", page.TotalPageNo)
	fmt.Println("总记录数是：", page.TotalRecord)
	fmt.Println("当前页中的图书有：")
	for _, v := range page.Books {
		fmt.Println("图书的信息是：", v)
	}
}
func testGetPageBooksByPrice(t *testing.T) {
	page, _ := GetPageBooksByPrice("3", "10", "30")
	fmt.Println("当前页是：", page.PageNo)
	fmt.Println("总页数是：", page.TotalPageNo)
	fmt.Println("总记录数是：", page.TotalRecord)
	fmt.Println("当前页中的图书有：")
	for _, v := range page.Books {
		fmt.Println("图书的信息是：", v)
	}
}

func TestSession(t *testing.T) {
	// fmt.Println("测试Session相关函数")
	// t.Run("测试添加Session", testAddSession)
	// t.Run("测试删除Session", testDeleteSession)
	// t.Run("测试获取Session", testGetSession)
}

func testAddSession(t *testing.T) {
	sess := &model.Session{
		SessionID: "13838381438",
		UserName:  "马蓉",
		UserID:    5,
	}
	AddSession(sess)
}

func testDeleteSession(t *testing.T) {
	DeleteSession("13838381438")
}
func testGetSession(t *testing.T) {
	sess, _ := GetSession("c65d2a76-9447-44cc-5fe8-c183e1414076")
	fmt.Println("Session的信息是：", sess)
}

func TestCart(t *testing.T) {
	// fmt.Println("测试购物车的相关函数")
	// t.Run("测试添加购物车", testAddCart)
	// t.Run("测试根据图书的id获取对应的购物项", testGetCartItemByBookID)
	// t.Run("测试根据购物车的id获取所有的购物项", testGetCartItemsByCartID)
	// t.Run("测试根据用户的id获取对应的购物车", testGetCartByUserID)
	// t.Run("测试根据图书的id和购物车的id以及输入的图书的数量更新购物项", testUpdateBookCount)
	// t.Run("测试购物车的id删除购物项和购物车", testDeleteCartByCartID)
	// t.Run("测试删除购物项", testDeleteCartItemByID)
}

func testAddCart(t *testing.T) {
	//设置要买的第一本书
	book := &model.Book{
		ID:    1,
		Price: 27.20,
	}
	//设置要买的第二本书
	book2 := &model.Book{
		ID:    2,
		Price: 23.00,
	}
	//创建一个购物项切片
	var cartItems []*model.CartItem
	//创建两个购物项
	cartItem := &model.CartItem{
		Book:   book,
		Count:  10,
		CartID: "66668888",
	}
	cartItems = append(cartItems, cartItem)
	cartItem2 := &model.CartItem{
		Book:   book2,
		Count:  10,
		CartID: "66668888",
	}
	cartItems = append(cartItems, cartItem2)
	//创建购物车
	cart := &model.Cart{
		CartID:    "66668888",
		CartItems: cartItems,
		UserID:    1,
	}
	//将购物车插入到数据库中
	AddCart(cart)
}

func testGetCartItemByBookID(t *testing.T) {
	cartItem, _ := GetCartItemByBookIDAndCartID("1", "66668888")
	fmt.Println("图书id=1的购物项的信息是：", cartItem)
}
func testGetCartItemsByCartID(t *testing.T) {
	cartItems, _ := GetCartItemsByCartID("66668888")
	for k, v := range cartItems {
		fmt.Printf("第%v个购物项是：%v\n", k+1, v)
	}
}

func testGetCartByUserID(t *testing.T) {
	cart, _ := GetCartByUserID(3)
	fmt.Println("id为2的用户的购物车信息是：", cart)
}

func testUpdateBookCount(t *testing.T) {
	// UpdateBookCount(100, 1, "66668888")
}
func testDeleteCartByCartID(t *testing.T) {
	DeleteCartByCartID("80bb8008-8383-47d0-4694-5eae94f39ffd")
}

func testDeleteCartItemByID(t *testing.T) {
	DeleteCartItemByID("21")
}

func TestOrder(t *testing.T) {
	fmt.Println("测试订单相关函数")
	// t.Run("测试添加订单和订单项", testAddOrder)
	// t.Run("测试获取所有的订单", testGetOrders)
	// t.Run("测试获取所有的订单项", testGetOrderItems)
	// t.Run("测试获取我的订单", testGetMyOrders)
	t.Run("测试发货和收货", testUpdateOrderState)

}

func testAddOrder(t *testing.T) {
	//生成订单号
	orderID := "88888888"
	//创建订单
	order := &model.Order{
		OrderID:     orderID,
		CreateTime:  time.Now().String(),
		TotalCount:  2,
		TotalAmount: 400,
		State:       0,
		UserID:      1,
	}
	//创建订单项
	orderItem := &model.OrderItem{
		Count:   1,
		Amount:  300,
		Title:   "三国演义",
		Author:  "罗贯中",
		Price:   300,
		ImgPath: "/static/img/default.jpg",
		OrderID: orderID,
	}
	orderItem2 := &model.OrderItem{
		Count:   1,
		Amount:  100,
		Title:   "西游记",
		Author:  "吴承恩",
		Price:   100,
		ImgPath: "/static/img/default.jpg",
		OrderID: orderID,
	}
	//保存订单
	AddOrder(order)
	//保存订单项
	AddOrderItem(orderItem)
	AddOrderItem(orderItem2)
}
func testGetOrders(t *testing.T) {
	orders, _ := GetOrders()
	for _, v := range orders {
		fmt.Println("订单信息是：", v)
	}
}
func testGetOrderItems(t *testing.T) {
	orderItems, _ := GetOrderItemsByOrderID("9a738546-d240-4c1a-7b3d-a3837100977f")
	for _, v := range orderItems {
		fmt.Println("订单项的信息是：", v)
	}
}

func testGetMyOrders(t *testing.T) {
	orders, _ := GetMyOrders(2)
	for _, v := range orders {
		fmt.Println("我的订单有：", v)
	}
}
func testUpdateOrderState(t *testing.T) {
	UpdateOrderState("5823f37e-f4e6-4a39-7567-2a7c0fd8e638", 1)
}
