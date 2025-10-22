package dao

import (
	"bookstore/model"
	"bookstore/utils"
	"fmt"
	"testing"
	"time"
)

// TestCartOperations 测试购物车操作
func TestCartOperations(t *testing.T) {
	// 准备测试数据
	testUserID := createTestUser(t) // 使用动态创建的测试用户ID
	// 确保测试用户存在，避免外键约束失败
	// ensureTestUser 已被替换为 createTestUser

	testBooks := []*model.Book{
		{Title: "购物车测试图书1", Author: "购物车测试作者1", Price: 10.00, Sales: 0, Stock: 100, ImgPath: "/static/img/cart1.jpg"},
		{Title: "购物车测试图书2", Author: "购物车测试作者2", Price: 20.00, Sales: 0, Stock: 100, ImgPath: "/static/img/cart2.jpg"},
		{Title: "购物车测试图书3", Author: "购物车测试作者3", Price: 30.00, Sales: 0, Stock: 100, ImgPath: "/static/img/cart3.jpg"},
	}

	var addedBookIDs []int
	for _, book := range testBooks {
		err := AddBook(book)
		if err != nil {
			t.Fatalf("添加测试图书失败: %v", err)
		}

		// 获取刚添加的图书ID
		books, err := GetBooks()
		if err != nil {
			t.Fatalf("获取图书列表失败: %v", err)
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
		// 清理测试用户
		cleanupTestUserByID(t, testUserID)
	}()

	// 测试添加商品到购物车
	t.Run("测试添加商品到购物车", func(t *testing.T) {
		// 创建购物车
		cartID := utils.CreateUUID()
		cart := &model.Cart{
			CartID: cartID,
			UserID: testUserID,
		}

		// 创建购物项
		cartItem := &model.CartItem{
			Book:   &model.Book{ID: addedBookIDs[0], Price: 10.00},
			Count:  2,
			CartID: cartID,
		}
		cart.CartItems = []*model.CartItem{cartItem}

		// 添加购物车到数据库
		err := AddCart(cart)
		if err != nil {
			t.Fatalf("添加购物车失败: %v", err)
		}

		// 验证购物车是否成功创建
		retrievedCart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Fatalf("获取购物车失败: %v", err)
		}

		if retrievedCart == nil {
			t.Fatal("购物车应该存在")
		}

		if retrievedCart.CartID != cartID {
			t.Errorf("购物车ID不匹配，期望: %s, 实际: %s", cartID, retrievedCart.CartID)
		}

		if retrievedCart.UserID != testUserID {
			t.Errorf("用户ID不匹配，期望: %d, 实际: %d", testUserID, retrievedCart.UserID)
		}

		if len(retrievedCart.CartItems) != 1 {
			t.Errorf("购物项数量不匹配，期望: 1, 实际: %d", len(retrievedCart.CartItems))
		}

		if retrievedCart.CartItems[0].Count != 2 {
			t.Errorf("购物项数量不匹配，期望: 2, 实际: %d", retrievedCart.CartItems[0].Count)
		}

		// 验证总价计算
		expectedTotalAmount := 2 * 10.00
		if retrievedCart.GetTotalAmount() != expectedTotalAmount {
			t.Errorf("总金额不匹配，期望: %.2f, 实际: %.2f", expectedTotalAmount, retrievedCart.GetTotalAmount())
		}

		if retrievedCart.GetTotalCount() != 2 {
			t.Errorf("总数量不匹配，期望: 2, 实际: %d", retrievedCart.GetTotalCount())
		}
	})

	// 测试重复添加同一商品
	t.Run("测试重复添加同一商品", func(t *testing.T) {
		// 获取现有购物车
		cart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil {
			t.Fatal("购物车不存在")
		}

		// 添加同一商品到购物车
		cartItem := &model.CartItem{
			Book:   &model.Book{ID: addedBookIDs[0], Price: 10.00},
			Count:  1,
			CartID: cart.CartID,
		}

		// 检查是否已存在该商品的购物项
		existingCartItem, err := GetCartItemByBookIDAndCartID(fmt.Sprintf("%d", addedBookIDs[0]), cart.CartID)
		if err != nil {
			t.Errorf("查询购物项失败: %v", err)
		}

		if existingCartItem != nil {
			// 更新现有购物项的数量
			existingCartItem.Count += 1
			err = UpdateBookCount(existingCartItem)
			if err != nil {
				t.Errorf("更新购物项失败: %v", err)
			}
		} else {
			// 创建新的购物项
			cart.CartItems = append(cart.CartItems, cartItem)
			err = AddCartItem(cartItem)
			if err != nil {
				t.Errorf("添加购物项失败: %v", err)
			}
		}

		// 更新购物车
		err = UpdateCart(cart)
		if err != nil {
			t.Errorf("更新购物车失败: %v", err)
		}

		// 验证购物车更新
		updatedCart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取更新后的购物车失败: %v", err)
		}

		if updatedCart.GetTotalCount() != 3 {
			t.Errorf("总数量不匹配，期望: 3, 实际: %d", updatedCart.GetTotalCount())
		}

		expectedTotalAmount := 3 * 10.00
		if updatedCart.GetTotalAmount() != expectedTotalAmount {
			t.Errorf("总金额不匹配，期望: %.2f, 实际: %.2f", expectedTotalAmount, updatedCart.GetTotalAmount())
		}
	})

	// 测试更新商品数量
	t.Run("测试更新商品数量", func(t *testing.T) {
		// 获取现有购物车
		cart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil {
			t.Fatal("购物车不存在")
		}

		// 更新第一个购物项的数量
		if len(cart.CartItems) > 0 {
			cart.CartItems[0].Count = 5
			err = UpdateBookCount(cart.CartItems[0])
			if err != nil {
				t.Errorf("更新购物项数量失败: %v", err)
			}

			// 更新购物车
			err = UpdateCart(cart)
			if err != nil {
				t.Errorf("更新购物车失败: %v", err)
			}

			// 验证更新结果
			updatedCart, err := GetCartByUserID(testUserID)
			if err != nil {
				t.Errorf("获取更新后的购物车失败: %v", err)
			}

			if updatedCart.CartItems[0].Count != 5 {
				t.Errorf("购物项数量不匹配，期望: 5, 实际: %d", updatedCart.CartItems[0].Count)
			}

			expectedTotalAmount := 5 * 10.00
			if updatedCart.GetTotalAmount() != expectedTotalAmount {
				t.Errorf("总金额不匹配，期望: %.2f, 实际: %.2f", expectedTotalAmount, updatedCart.GetTotalAmount())
			}
		}
	})

	// 测试删除购物项
	t.Run("测试删除购物项", func(t *testing.T) {
		// 获取现有购物车
		cart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil {
			t.Fatal("购物车不存在")
		}

		if len(cart.CartItems) == 0 {
			t.Fatal("购物车中没有购物项")
		}

		// 删除第一个购物项
		cartItemID := cart.CartItems[0].CartItemID
		err = DeleteCartItemByID(fmt.Sprintf("%d", cartItemID))
		if err != nil {
			t.Errorf("删除购物项失败: %v", err)
		}

		// 从购物车中移除该购物项
		cart.CartItems = cart.CartItems[1:]

		// 更新购物车
		err = UpdateCart(cart)
		if err != nil {
			t.Errorf("更新购物车失败: %v", err)
		}

		// 验证删除结果
		updatedCart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取更新后的购物车失败: %v", err)
		}

		if len(updatedCart.CartItems) != 0 {
			t.Errorf("购物项数量不匹配，期望: 0, 实际: %d", len(updatedCart.CartItems))
		}

		if updatedCart.GetTotalCount() != 0 {
			t.Errorf("总数量不匹配，期望: 0, 实际: %d", updatedCart.GetTotalCount())
		}

		if updatedCart.GetTotalAmount() != 0 {
			t.Errorf("总金额不匹配，期望: 0.00, 实际: %.2f", updatedCart.GetTotalAmount())
		}
	})

	// 测试清空购物车
	t.Run("测试清空购物车", func(t *testing.T) {
		// 先添加一些购物项
		cart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if cart == nil {
			t.Fatal("购物车不存在")
		}

		// 添加多个购物项
		for i, bookID := range addedBookIDs {
			cartItem := &model.CartItem{
				Book:   &model.Book{ID: bookID, Price: float64((i + 1) * 10)},
				Count:  int64(i + 1),
				CartID: cart.CartID,
			}
			cart.CartItems = append(cart.CartItems, cartItem)
			err = AddCartItem(cartItem)
			if err != nil {
				t.Errorf("添加购物项失败: %v", err)
			}
		}

		// 更新购物车
		err = UpdateCart(cart)
		if err != nil {
			t.Errorf("更新购物车失败: %v", err)
		}

		// 验证购物车有内容
		updatedCart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if len(updatedCart.CartItems) == 0 {
			t.Error("购物车应该包含购物项")
		}

		// 清空购物车
		err = DeleteCartByCartID(cart.CartID)
		if err != nil {
			t.Errorf("清空购物车失败: %v", err)
		}

		// 验证购物车已被清空
		emptyCart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取清空后的购物车失败: %v", err)
		}

		if emptyCart != nil {
			t.Error("购物车应该已被清空")
		}
	})

	// 测试购物车总价计算
	t.Run("测试购物车总价计算", func(t *testing.T) {
		// 创建新的购物车
		cartID := utils.CreateUUID()
		cart := &model.Cart{
			CartID: cartID,
			UserID: testUserID,
		}

		// 添加多个不同价格的购物项
		cartItems := []*model.CartItem{
			{Book: &model.Book{ID: addedBookIDs[0], Price: 10.00}, Count: 2, CartID: cartID},
			{Book: &model.Book{ID: addedBookIDs[1], Price: 20.00}, Count: 1, CartID: cartID},
			{Book: &model.Book{ID: addedBookIDs[2], Price: 30.00}, Count: 3, CartID: cartID},
		}

		cart.CartItems = cartItems

		// 添加购物车到数据库
		err := AddCart(cart)
		if err != nil {
			t.Errorf("添加购物车失败: %v", err)
		}

		// 验证总价计算
		retrievedCart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		// 计算期望的总价
		expectedTotalAmount := (2 * 10.00) + (1 * 20.00) + (3 * 30.00) // 20 + 20 + 90 = 130
		expectedTotalCount := int64(2 + 1 + 3)                         // 6

		if retrievedCart.GetTotalAmount() != expectedTotalAmount {
			t.Errorf("总金额不匹配，期望: %.2f, 实际: %.2f", expectedTotalAmount, retrievedCart.GetTotalAmount())
		}

		if retrievedCart.GetTotalCount() != expectedTotalCount {
			t.Errorf("总数量不匹配，期望: %d, 实际: %d", expectedTotalCount, retrievedCart.GetTotalCount())
		}

		// 验证每个购物项的小计
		for i, cartItem := range retrievedCart.CartItems {
			expectedAmount := float64(cartItem.Count) * cartItem.Book.Price
			if cartItem.GetAmount() != expectedAmount {
				t.Errorf("购物项%d小计不匹配，期望: %.2f, 实际: %.2f", i, expectedAmount, cartItem.GetAmount())
			}
		}
	})
}

// TestCartDataValidation 测试购物车数据验证
func TestCartDataValidation(t *testing.T) {
	testUserID := createTestUser(t)
	testBookID := 1

	defer func() {
		cleanupTestCart(t, testUserID)
		cleanupTestUserByID(t, testUserID)
	}()

	tests := []struct {
		name        string
		cartItem    *model.CartItem
		expectError bool
		description string
	}{
		{
			name: "正常购物项",
			cartItem: &model.CartItem{
				Book:   &model.Book{ID: testBookID, Price: 10.00},
				Count:  1,
				CartID: "test_cart_id",
			},
			expectError: false,
			description: "测试正常的购物项",
		},
		{
			name: "零数量购物项",
			cartItem: &model.CartItem{
				Book:   &model.Book{ID: testBookID, Price: 10.00},
				Count:  0,
				CartID: "test_cart_id",
			},
			expectError: false, // 这里假设系统允许零数量，实际项目中可能需要验证
			description: "测试零数量购物项",
		},
		{
			name: "负数量购物项",
			cartItem: &model.CartItem{
				Book:   &model.Book{ID: testBookID, Price: 10.00},
				Count:  -1,
				CartID: "test_cart_id",
			},
			expectError: false, // 这里假设系统允许负数量，实际项目中可能需要验证
			description: "测试负数量购物项",
		},
		{
			name: "大数量购物项",
			cartItem: &model.CartItem{
				Book:   &model.Book{ID: testBookID, Price: 10.00},
				Count:  999999,
				CartID: "test_cart_id",
			},
			expectError: false,
			description: "测试大数量购物项",
		},
		{
			name: "零价格图书购物项",
			cartItem: &model.CartItem{
				Book:   &model.Book{ID: testBookID, Price: 0.00},
				Count:  1,
				CartID: "test_cart_id",
			},
			expectError: false,
			description: "测试零价格图书购物项",
		},
		{
			name: "负价格图书购物项",
			cartItem: &model.CartItem{
				Book:   &model.Book{ID: testBookID, Price: -10.00},
				Count:  1,
				CartID: "test_cart_id",
			},
			expectError: false, // 这里假设系统允许负价格，实际项目中可能需要验证
			description: "测试负价格图书购物项",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建购物车
			cartID := utils.CreateUUID()
			cart := &model.Cart{
				CartID: cartID,
				UserID: testUserID,
			}

			// 设置购物项的购物车ID
			tt.cartItem.CartID = cartID
			cart.CartItems = []*model.CartItem{tt.cartItem}

			// 添加购物车
			err := AddCart(cart)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望返回错误，但没有返回错误: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("不期望返回错误，但返回了错误: %v, %s", err, tt.description)
				} else {
					// 验证购物车是否成功创建
					retrievedCart, err := GetCartByUserID(testUserID)
					if err != nil {
						t.Errorf("获取购物车失败: %v", err)
					}

					if retrievedCart == nil {
						t.Errorf("购物车应该存在: %s", tt.description)
					}

					// 清理测试数据
					cleanupTestCart(t, testUserID)
				}
			}
		})
	}
}

// TestCartConcurrentOperations 测试购物车并发操作
func TestCartConcurrentOperations(t *testing.T) {
	testUserID := createTestUser(t)
	testBookID := 1

	defer func() {
		cleanupTestCart(t, testUserID)
		cleanupTestUserByID(t, testUserID)
	}()

	// 测试并发添加购物项
	t.Run("测试并发添加购物项", func(t *testing.T) {
		// 创建购物车
		cartID := utils.CreateUUID()
		cart := &model.Cart{
			CartID: cartID,
			UserID: testUserID,
		}

		err := AddCart(cart)
		if err != nil {
			t.Errorf("创建购物车失败: %v", err)
		}

		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				cartItem := &model.CartItem{
					Book:   &model.Book{ID: testBookID, Price: float64(index * 10)},
					Count:  int64(index + 1),
					CartID: cartID,
				}

				err := AddCartItem(cartItem)
				if err != nil {
					t.Errorf("并发添加购物项失败: %v", err)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证购物车状态
		retrievedCart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if retrievedCart == nil {
			t.Error("购物车应该存在")
		}

		// 清理测试数据
		cleanupTestCart(t, testUserID)
	})

	// 测试并发更新购物车
	t.Run("测试并发更新购物车", func(t *testing.T) {
		// 创建购物车
		cartID := utils.CreateUUID()
		cart := &model.Cart{
			CartID: cartID,
			UserID: testUserID,
		}

		// 添加初始购物项
		cartItem := &model.CartItem{
			Book:   &model.Book{ID: testBookID, Price: 10.00},
			Count:  1,
			CartID: cartID,
		}
		cart.CartItems = []*model.CartItem{cartItem}

		err := AddCart(cart)
		if err != nil {
			t.Errorf("创建购物车失败: %v", err)
		}

		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				// 获取购物车
				cart, err := GetCartByUserID(testUserID)
				if err != nil {
					t.Errorf("获取购物车失败: %v", err)
					return
				}

				if cart != nil && len(cart.CartItems) > 0 {
					// 更新购物项数量
					cart.CartItems[0].Count += 1
					err = UpdateBookCount(cart.CartItems[0])
					if err != nil {
						t.Errorf("更新购物项失败: %v", err)
						return
					}

					// 更新购物车
					err = UpdateCart(cart)
					if err != nil {
						t.Errorf("更新购物车失败: %v", err)
					}
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证购物车状态
		retrievedCart, err := GetCartByUserID(testUserID)
		if err != nil {
			t.Errorf("获取购物车失败: %v", err)
		}

		if retrievedCart == nil {
			t.Error("购物车应该存在")
		}

		// 清理测试数据
		cleanupTestCart(t, testUserID)
	})
}

// BenchmarkCartOperations 性能测试
func BenchmarkAddCart(b *testing.B) {
	testUserID := createTestUserNoT()
	testBookID := 1

	defer func() {
		cleanupTestCart(nil, testUserID)
		cleanupTestUserByID(nil, testUserID)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cartID := utils.CreateUUID()
		cart := &model.Cart{
			CartID: cartID,
			UserID: testUserID,
		}

		cartItem := &model.CartItem{
			Book:   &model.Book{ID: testBookID, Price: 10.00},
			Count:  1,
			CartID: cartID,
		}
		cart.CartItems = []*model.CartItem{cartItem}

		AddCart(cart)
	}
}

func BenchmarkGetCartByUserID(b *testing.B) {
	testUserID := createTestUserNoT()
	testBookID := 1

	// 准备测试数据
	cartID := utils.CreateUUID()
	cart := &model.Cart{
		CartID: cartID,
		UserID: testUserID,
	}

	cartItem := &model.CartItem{
		Book:   &model.Book{ID: testBookID, Price: 10.00},
		Count:  1,
		CartID: cartID,
	}
	cart.CartItems = []*model.CartItem{cartItem}

	AddCart(cart)

	defer func() {
		cleanupTestCart(nil, testUserID)
		cleanupTestUserByID(nil, testUserID)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCartByUserID(testUserID)
	}
}

// cleanupTestCart 清理测试购物车
func cleanupTestCart(t *testing.T, userID int) {
	// 获取用户的购物车
	cart, err := GetCartByUserID(userID)
	if err != nil {
		if t != nil {
			t.Logf("获取购物车失败: %v", err)
		}
		return
	}

	if cart != nil {
		// 删除购物车
		err = DeleteCartByCartID(cart.CartID)
		if err != nil {
			if t != nil {
				t.Logf("删除购物车失败: %v", err)
			}
		}
	}
}

// createTestUser 创建一个临时测试用户并返回其ID
func createTestUser(t *testing.T) int {
	username := fmt.Sprintf("test_user_%d", time.Now().UnixNano())
	email := fmt.Sprintf("%s@example.com", username)
	res, err := utils.Db.Exec("INSERT INTO users (username, password, email) VALUES (?, ?, ?)", username, "password", email)
	if err != nil {
		t.Fatalf("createTestUser 插入用户失败: %v", err)
	}
	id64, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("createTestUser 获取插入ID失败: %v", err)
	}
	return int(id64)
}

// createTestUserNoT 为基准测试创建用户（不使用 *testing.T）
func createTestUserNoT() int {
	username := fmt.Sprintf("bench_user_%d", time.Now().UnixNano())
	email := fmt.Sprintf("%s@example.com", username)
	res, err := utils.Db.Exec("INSERT INTO users (username, password, email) VALUES (?, ?, ?)", username, "password", email)
	if err != nil {
		// 如果插入失败，回退到默认存在的用户 id=1
		return 1
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 1
	}
	return int(id64)
}
