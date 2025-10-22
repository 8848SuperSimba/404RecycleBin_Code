package dao

import (
	"bookstore/model"
	"bookstore/utils"
	"fmt"
	"testing"
	"time"
)

// TestOrderFlow 测试订单流程
func TestOrderFlow(t *testing.T) {
	// 准备测试数据
	testUserID := 777

	// 确保测试环境满足：测试用户存在、orders 表包含 create_time 列（有些环境缺失）
	ensureTestSetup(t, testUserID)

	testBooks := []*model.Book{
		{Title: "订单测试图书1", Author: "订单测试作者1", Price: 10.00, Sales: 0, Stock: 100, ImgPath: "/static/img/order1.jpg"},
		{Title: "订单测试图书2", Author: "订单测试作者2", Price: 20.00, Sales: 0, Stock: 100, ImgPath: "/static/img/order2.jpg"},
		{Title: "订单测试图书3", Author: "订单测试作者3", Price: 30.00, Sales: 0, Stock: 100, ImgPath: "/static/img/order3.jpg"},
	}

	var addedBookIDs []int
	for _, book := range testBooks {
		err := AddBook(book)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取刚添加的图书ID
		books, err := GetBooks()
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
		cleanupTestOrder(t, testUserID)
		cleanupTestUserByID(t, testUserID)
	}()

	// 测试创建订单
	t.Run("测试创建订单", func(t *testing.T) {
		// 创建购物车
		cartID := utils.CreateUUID()
		cart := &model.Cart{
			CartID: cartID,
			UserID: testUserID,
		}

		// 创建购物项
		cartItems := []*model.CartItem{
			{Book: &model.Book{ID: addedBookIDs[0], Price: 10.00}, Count: 2, CartID: cartID},
			{Book: &model.Book{ID: addedBookIDs[1], Price: 20.00}, Count: 1, CartID: cartID},
		}
		cart.CartItems = cartItems

		// 添加购物车到数据库
		err := AddCart(cart)
		if err != nil {
			t.Errorf("添加购物车失败: %v", err)
		}

		// 创建订单
		orderID := utils.CreateUUID()
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		order := &model.Order{
			OrderID:     orderID,
			CreateTime:  timeStr,
			TotalCount:  cart.GetTotalCount(),
			TotalAmount: cart.GetTotalAmount(),
			State:       0, // 未发货
			UserID:      int64(testUserID),
		}

		// 添加订单到数据库
		err = AddOrder(order)
		if err != nil {
			t.Errorf("添加订单失败: %v", err)
		}

		// 验证订单是否成功创建
		orders, err := GetOrders()
		if err != nil {
			t.Errorf("获取订单列表失败: %v", err)
		}

		var foundOrder *model.Order
		for _, o := range orders {
			if o.OrderID == orderID {
				foundOrder = o
				break
			}
		}

		if foundOrder == nil {
			t.Error("订单应该存在")
			return
		}

		if foundOrder.UserID != int64(testUserID) {
			t.Errorf("用户ID不匹配，期望: %d, 实际: %d", testUserID, foundOrder.UserID)
		}

		if foundOrder.State != 0 {
			t.Errorf("订单状态不匹配，期望: 0, 实际: %d", foundOrder.State)
		}

		// 验证订单总价计算
		expectedTotalAmount := (2 * 10.00) + (1 * 20.00) // 40.00
		if foundOrder.TotalAmount != expectedTotalAmount {
			t.Errorf("订单总金额不匹配，期望: %.2f, 实际: %.2f", expectedTotalAmount, foundOrder.TotalAmount)
		}

		expectedTotalCount := int64(2 + 1) // 3
		if foundOrder.TotalCount != expectedTotalCount {
			t.Errorf("订单总数量不匹配，期望: %d, 实际: %d", expectedTotalCount, foundOrder.TotalCount)
		}

		// 创建订单项
		for i, cartItem := range cartItems {
			orderItem := &model.OrderItem{
				Count:   cartItem.Count,
				Amount:  cartItem.GetAmount(),
				Title:   cartItem.Book.Title,
				Author:  cartItem.Book.Author,
				Price:   cartItem.Book.Price,
				ImgPath: cartItem.Book.ImgPath,
				OrderID: orderID,
			}

			err = AddOrderItem(orderItem)
			if err != nil {
				t.Errorf("添加订单项失败: %v", err)
			}

			// 验证订单项
			orderItems, err := GetOrderItemsByOrderID(orderID)
			if err != nil {
				t.Errorf("获取订单项失败: %v", err)
			}

			if len(orderItems) != i+1 {
				t.Errorf("订单项数量不匹配，期望: %d, 实际: %d", i+1, len(orderItems))
			}

			// 验证订单项信息
			lastOrderItem := orderItems[len(orderItems)-1]
			if lastOrderItem.Title != cartItem.Book.Title {
				t.Errorf("订单项标题不匹配，期望: %s, 实际: %s", cartItem.Book.Title, lastOrderItem.Title)
			}

			if lastOrderItem.Count != cartItem.Count {
				t.Errorf("订单项数量不匹配，期望: %d, 实际: %d", cartItem.Count, lastOrderItem.Count)
			}

			expectedAmount := float64(cartItem.Count) * cartItem.Book.Price
			if lastOrderItem.Amount != expectedAmount {
				t.Errorf("订单项金额不匹配，期望: %.2f, 实际: %.2f", expectedAmount, lastOrderItem.Amount)
			}
		}
	})

	// 测试订单状态更新
	t.Run("测试订单状态更新", func(t *testing.T) {
		// 创建订单
		orderID := utils.CreateUUID()
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		order := &model.Order{
			OrderID:     orderID,
			CreateTime:  timeStr,
			TotalCount:  1,
			TotalAmount: 10.00,
			State:       0, // 未发货
			UserID:      int64(testUserID),
		}

		err := AddOrder(order)
		if err != nil {
			t.Errorf("添加订单失败: %v", err)
		}

		// 测试发货（状态从0更新为1）
		err = UpdateOrderState(orderID, 1)
		if err != nil {
			t.Errorf("更新订单状态失败: %v", err)
		}

		// 验证订单状态
		orders, err := GetOrders()
		if err != nil {
			t.Errorf("获取订单列表失败: %v", err)
		}

		var foundOrder *model.Order
		for _, o := range orders {
			if o.OrderID == orderID {
				foundOrder = o
				break
			}
		}

		if foundOrder == nil {
			t.Error("订单应该存在")
			return
		}

		if foundOrder.State != 1 {
			t.Errorf("订单状态不匹配，期望: 1, 实际: %d", foundOrder.State)
		}

		// 测试收货（状态从1更新为2）
		err = UpdateOrderState(orderID, 2)
		if err != nil {
			t.Errorf("更新订单状态失败: %v", err)
		}

		// 验证订单状态
		orders, err = GetOrders()
		if err != nil {
			t.Errorf("获取订单列表失败: %v", err)
		}

		foundOrder = nil
		for _, o := range orders {
			if o.OrderID == orderID {
				foundOrder = o
				break
			}
		}

		if foundOrder == nil {
			t.Error("订单应该存在")
			return
		}

		if foundOrder.State != 2 {
			t.Errorf("订单状态不匹配，期望: 2, 实际: %d", foundOrder.State)
		}
	})

	// 测试库存扣减
	t.Run("测试库存扣减", func(t *testing.T) {
		// 获取测试图书的初始库存
		initialBook, err := GetBookByID(fmt.Sprintf("%d", addedBookIDs[0]))
		if err != nil {
			t.Errorf("获取图书失败: %v", err)
		}

		initialStock := initialBook.Stock
		initialSales := initialBook.Sales

		// 创建购物车
		cartID := utils.CreateUUID()
		cart := &model.Cart{
			CartID: cartID,
			UserID: testUserID,
		}

		// 创建购物项
		cartItem := &model.CartItem{
			Book:   &model.Book{ID: addedBookIDs[0], Price: 10.00},
			Count:  3,
			CartID: cartID,
		}
		cart.CartItems = []*model.CartItem{cartItem}

		// 添加购物车到数据库
		err = AddCart(cart)
		if err != nil {
			t.Errorf("添加购物车失败: %v", err)
		}

		// 创建订单
		orderID := utils.CreateUUID()
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		order := &model.Order{
			OrderID:     orderID,
			CreateTime:  timeStr,
			TotalCount:  cart.GetTotalCount(),
			TotalAmount: cart.GetTotalAmount(),
			State:       0,
			UserID:      int64(testUserID),
		}

		// 添加订单到数据库
		err = AddOrder(order)
		if err != nil {
			t.Errorf("添加订单失败: %v", err)
		}

		// 模拟库存扣减和销量更新
		// 从数据库中读取完整的图书信息再更新，避免使用只包含ID/Price的局部对象导致字段丢失
		dbBook, err := GetBookByID(fmt.Sprintf("%d", cartItem.Book.ID))
		if err != nil {
			t.Errorf("获取数据库中图书失败: %v", err)
		} else {
			dbBook.Sales = dbBook.Sales + int(cartItem.Count)
			dbBook.Stock = dbBook.Stock - int(cartItem.Count)

			// 更新图书信息
			err = UpdateBook(dbBook)
			if err != nil {
				t.Errorf("更新图书信息失败: %v", err)
			}
		}

		// 验证库存扣减
		updatedBook, err := GetBookByID(fmt.Sprintf("%d", addedBookIDs[0]))
		if err != nil {
			t.Errorf("获取更新后的图书失败: %v", err)
		}

		expectedStock := initialStock - int(cartItem.Count)
		if updatedBook.Stock != expectedStock {
			t.Errorf("库存扣减不匹配，期望: %d, 实际: %d", expectedStock, updatedBook.Stock)
		}

		expectedSales := initialSales + int(cartItem.Count)
		if updatedBook.Sales != expectedSales {
			t.Errorf("销量更新不匹配，期望: %d, 实际: %d", expectedSales, updatedBook.Sales)
		}
	})

	// 测试销量更新
	t.Run("测试销量更新", func(t *testing.T) {
		// 获取测试图书的初始销量
		initialBook, err := GetBookByID(fmt.Sprintf("%d", addedBookIDs[1]))
		if err != nil {
			t.Errorf("获取图书失败: %v", err)
		}

		initialSales := initialBook.Sales

		// 创建购物车
		cartID := utils.CreateUUID()
		cart := &model.Cart{
			CartID: cartID,
			UserID: testUserID,
		}

		// 创建购物项
		cartItem := &model.CartItem{
			Book:   &model.Book{ID: addedBookIDs[1], Price: 20.00},
			Count:  2,
			CartID: cartID,
		}
		cart.CartItems = []*model.CartItem{cartItem}

		// 添加购物车到数据库
		err = AddCart(cart)
		if err != nil {
			t.Errorf("添加购物车失败: %v", err)
		}

		// 创建订单
		orderID := utils.CreateUUID()
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		order := &model.Order{
			OrderID:     orderID,
			CreateTime:  timeStr,
			TotalCount:  cart.GetTotalCount(),
			TotalAmount: cart.GetTotalAmount(),
			State:       0,
			UserID:      int64(testUserID),
		}

		// 添加订单到数据库
		err = AddOrder(order)
		if err != nil {
			t.Errorf("添加订单失败: %v", err)
		}

		// 模拟销量更新
		book := cartItem.Book
		book.Sales = book.Sales + int(cartItem.Count)

		// 更新图书信息
		err = UpdateBook(book)
		if err != nil {
			t.Errorf("更新图书信息失败: %v", err)
		}

		// 验证销量更新
		updatedBook, err := GetBookByID(fmt.Sprintf("%d", addedBookIDs[1]))
		if err != nil {
			t.Errorf("获取更新后的图书失败: %v", err)
		}

		expectedSales := initialSales + int(cartItem.Count)
		if updatedBook.Sales != expectedSales {
			t.Errorf("销量更新不匹配，期望: %d, 实际: %d", expectedSales, updatedBook.Sales)
		}
	})

	// 测试订单查询
	t.Run("测试订单查询", func(t *testing.T) {
		// 创建多个订单
		orderIDs := []string{}
		for i := 0; i < 3; i++ {
			orderID := utils.CreateUUID()
			timeStr := time.Now().Format("2006-01-02 15:04:05")
			order := &model.Order{
				OrderID:     orderID,
				CreateTime:  timeStr,
				TotalCount:  int64(i + 1),
				TotalAmount: float64((i + 1) * 10),
				State:       int64(i % 3), // 0, 1, 2
				UserID:      int64(testUserID),
			}

			err := AddOrder(order)
			if err != nil {
				t.Errorf("添加订单失败: %v", err)
			}

			orderIDs = append(orderIDs, orderID)
		}

		// 测试获取所有订单
		allOrders, err := GetOrders()
		if err != nil {
			t.Errorf("获取所有订单失败: %v", err)
		}

		if len(allOrders) < 3 {
			t.Errorf("订单数量不匹配，期望: >=3, 实际: %d", len(allOrders))
		}

		// 验证订单信息
		for _, orderID := range orderIDs {
			var found bool
			for _, order := range allOrders {
				if order.OrderID == orderID {
					found = true
					if order.UserID != int64(testUserID) {
						t.Errorf("用户ID不匹配，期望: %d, 实际: %d", testUserID, order.UserID)
					}
					break
				}
			}
			if !found {
				t.Errorf("订单 %s 未找到", orderID)
			}
		}

		// 测试获取我的订单
		myOrders, err := GetMyOrders(testUserID)
		if err != nil {
			t.Errorf("获取我的订单失败: %v", err)
		}

		if len(myOrders) < 3 {
			t.Errorf("我的订单数量不匹配，期望: >=3, 实际: %d", len(myOrders))
		}

		// 验证我的订单
		for _, order := range myOrders {
			if order.UserID != int64(testUserID) {
				t.Errorf("我的订单用户ID不匹配，期望: %d, 实际: %d", testUserID, order.UserID)
			}
		}
	})
}

func cleanupTestBook(t *testing.T, bookID interface{}) {
	var idStr string
	switch v := bookID.(type) {
	case int:
		idStr = fmt.Sprintf("%d", v)
	case string:
		idStr = v
	default:
		t.Logf("不支持的bookID类型: %T", bookID)
		return
	}

	// 先删除与该图书关联的购物项（cart_items）和订单项（order_items），避免外键约束
	_, _ = utils.Db.Exec("DELETE FROM cart_items WHERE book_id = ?", idStr)
	_, _ = utils.Db.Exec("DELETE FROM order_items WHERE title IN (SELECT title FROM books WHERE id = ?)", idStr)

	sqlStr := "DELETE FROM books WHERE id = ?"
	_, err := utils.Db.Exec(sqlStr, idStr)
	if err != nil {
		t.Logf("清理测试图书失败: %v", err)
	}
}

// TestOrderDataValidation 测试订单数据验证
func TestOrderDataValidation(t *testing.T) {
	testUserID := 776
	ensureTestSetup(t, testUserID)

	defer func() {
		cleanupTestOrder(t, testUserID)
		cleanupTestUserByID(t, testUserID)
	}()

	tests := []struct {
		name        string
		order       *model.Order
		expectError bool
		description string
	}{
		{
			name: "正常订单",
			order: &model.Order{
				OrderID:     utils.CreateUUID(),
				CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
				TotalCount:  1,
				TotalAmount: 10.00,
				State:       0,
				UserID:      int64(testUserID),
			},
			expectError: false,
			description: "测试正常的订单",
		},
		{
			name: "零金额订单",
			order: &model.Order{
				OrderID:     utils.CreateUUID(),
				CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
				TotalCount:  1,
				TotalAmount: 0.00,
				State:       0,
				UserID:      int64(testUserID),
			},
			expectError: false,
			description: "测试零金额订单",
		},
		{
			name: "负金额订单",
			order: &model.Order{
				OrderID:     utils.CreateUUID(),
				CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
				TotalCount:  1,
				TotalAmount: -10.00,
				State:       0,
				UserID:      int64(testUserID),
			},
			expectError: false, // 这里假设系统允许负金额，实际项目中可能需要验证
			description: "测试负金额订单",
		},
		{
			name: "零数量订单",
			order: &model.Order{
				OrderID:     utils.CreateUUID(),
				CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
				TotalCount:  0,
				TotalAmount: 10.00,
				State:       0,
				UserID:      int64(testUserID),
			},
			expectError: false, // 这里假设系统允许零数量，实际项目中可能需要验证
			description: "测试零数量订单",
		},
		{
			name: "高金额订单",
			order: &model.Order{
				OrderID:     utils.CreateUUID(),
				CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
				TotalCount:  1,
				TotalAmount: 999999.99,
				State:       0,
				UserID:      int64(testUserID),
			},
			expectError: false,
			description: "测试高金额订单",
		},
		{
			name: "已完成订单",
			order: &model.Order{
				OrderID:     utils.CreateUUID(),
				CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
				TotalCount:  1,
				TotalAmount: 10.00,
				State:       2, // 已完成
				UserID:      int64(testUserID),
			},
			expectError: false,
			description: "测试已完成订单",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddOrder(tt.order)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望返回错误，但没有返回错误: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("不期望返回错误，但返回了错误: %v, %s", err, tt.description)
				} else {
					// 验证订单是否成功创建
					orders, err := GetOrders()
					if err != nil {
						t.Errorf("获取订单列表失败: %v", err)
					}

					var found bool
					for _, order := range orders {
						if order.OrderID == tt.order.OrderID {
							found = true
							break
						}
					}

					if !found {
						t.Errorf("添加的订单未在数据库中找到: %s", tt.description)
					}
				}
			}
		})
	}
}

// TestOrderConcurrentOperations 测试订单并发操作
func TestOrderConcurrentOperations(t *testing.T) {
	testUserID := 775
	ensureTestSetup(t, testUserID)

	defer func() {
		cleanupTestOrder(t, testUserID)
		cleanupTestUserByID(t, testUserID)
	}()

	// 测试并发创建订单
	t.Run("测试并发创建订单", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				orderID := utils.CreateUUID()
				timeStr := time.Now().Format("2006-01-02 15:04:05")
				order := &model.Order{
					OrderID:     orderID,
					CreateTime:  timeStr,
					TotalCount:  int64(index + 1),
					TotalAmount: float64((index + 1) * 10),
					State:       0,
					UserID:      int64(testUserID),
				}

				err := AddOrder(order)
				if err != nil {
					t.Errorf("并发创建订单失败: %v", err)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证订单创建
		orders, err := GetMyOrders(testUserID)
		if err != nil {
			t.Errorf("获取我的订单失败: %v", err)
		}

		if len(orders) < 10 {
			t.Errorf("订单数量不匹配，期望: >=10, 实际: %d", len(orders))
		}
	})

	// 测试并发更新订单状态
	t.Run("测试并发更新订单状态", func(t *testing.T) {
		// 先创建一个订单
		orderID := utils.CreateUUID()
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		order := &model.Order{
			OrderID:     orderID,
			CreateTime:  timeStr,
			TotalCount:  1,
			TotalAmount: 10.00,
			State:       0,
			UserID:      int64(testUserID),
		}

		err := AddOrder(order)
		if err != nil {
			t.Errorf("创建订单失败: %v", err)
		}

		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				// 更新订单状态
				state := int64(index % 3) // 0, 1, 2
				err := UpdateOrderState(orderID, state)
				if err != nil {
					t.Errorf("并发更新订单状态失败: %v", err)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证订单状态
		orders, err := GetOrders()
		if err != nil {
			t.Errorf("获取订单列表失败: %v", err)
		}

		var foundOrder *model.Order
		for _, o := range orders {
			if o.OrderID == orderID {
				foundOrder = o
				break
			}
		}

		if foundOrder == nil {
			t.Error("订单应该存在")
			return
		}
	})
}

// BenchmarkAddOrder 性能测试
func BenchmarkAddOrder(b *testing.B) {
	testUserID := 774
	ensureTestSetup(nil, testUserID)

	defer func() {
		cleanupTestOrder(nil, testUserID)
		cleanupTestUserByID(nil, testUserID)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orderID := utils.CreateUUID()
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		order := &model.Order{
			OrderID:     orderID,
			CreateTime:  timeStr,
			TotalCount:  int64(i + 1),
			TotalAmount: float64((i + 1) * 10),
			State:       0,
			UserID:      int64(testUserID),
		}
		AddOrder(order)
	}
}

func BenchmarkGetMyOrders(b *testing.B) {
	testUserID := 773
	ensureTestSetup(nil, testUserID)

	// 准备测试数据
	orderID := utils.CreateUUID()
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	order := &model.Order{
		OrderID:     orderID,
		CreateTime:  timeStr,
		TotalCount:  1,
		TotalAmount: 10.00,
		State:       0,
		UserID:      int64(testUserID),
	}

	AddOrder(order)

	defer func() {
		cleanupTestOrder(nil, testUserID)
		cleanupTestUserByID(nil, testUserID)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetMyOrders(testUserID)
	}
}

// cleanup and helpers
// cleanupTestOrder 清理测试订单
func cleanupTestOrder(t *testing.T, userID int) {
	// 获取用户的订单
	orders, err := GetMyOrders(userID)
	if err != nil {
		if t != nil {
			t.Logf("获取订单失败: %v", err)
		}
		return
	}

	// 删除所有订单
	for _, order := range orders {
		// 删除订单项
		sqlStr := "DELETE FROM order_items WHERE order_id = ?"
		_, err := utils.Db.Exec(sqlStr, order.OrderID)
		if err != nil {
			if t != nil {
				t.Logf("删除订单项失败: %v", err)
			}
		}

		// 删除订单
		sqlStr = "DELETE FROM orders WHERE id = ?"
		_, err = utils.Db.Exec(sqlStr, order.OrderID)
		if err != nil {
			if t != nil {
				t.Logf("删除订单失败: %v", err)
			}
		}
	}
}

// ensureTestSetup 确保测试用户存在并修复缺失的 create_time 列（t 可以为 nil）
func ensureTestSetup(t *testing.T, userID int) {
	// 确保用户存在
	err := ensureTestUser(t, userID)
	if err != nil {
		if t != nil {
			t.Fatalf("ensureTestUser 失败: %v", err)
		}
		return
	}

	// 确保 orders.create_time 列存在（某些测试 DB 可能缺失）
	var cnt int
	row := utils.Db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_schema = DATABASE() AND table_name = 'orders' AND column_name = 'create_time'")
	err = row.Scan(&cnt)
	if err != nil {
		if t != nil {
			t.Logf("检查 create_time 列时出错: %v", err)
		}
		return
	}
	if cnt == 0 {
		// 列不存在，尝试添加（兼容大多数 MySQL 版本）
		_, err := utils.Db.Exec("ALTER TABLE orders ADD COLUMN create_time DATETIME NULL DEFAULT NULL")
		if err != nil {
			if t != nil {
				t.Logf("尝试添加 create_time 列失败: %v", err)
			}
			// 不致命：如果不能添加，后续 Insert 会报错，测试会按原样失败以提示手动修复
		}
	}
}

// ensureTestUser 创建测试用户（尽量兼容不同 users 表结构），返回 error 而不直接 Fatal，以便在 benchmark 中使用 nil t
func ensureTestUser(t *testing.T, userID int) error {
	// 尝试常见列集合
	sqls := []string{
		"INSERT IGNORE INTO users (id, username, password, email) VALUES (?, ?, ?, ?)",
		"INSERT IGNORE INTO users (id, username, password) VALUES (?, ?, ?)",
		"INSERT IGNORE INTO users (id, username) VALUES (?, ?)",
	}
	var lastErr error
	for _, s := range sqls {
		var err error
		switch s {
		case sqls[0]:
			_, err = utils.Db.Exec(s, userID, fmt.Sprintf("test_user_%d", userID), "password", fmt.Sprintf("test_%d@example.com", userID))
		case sqls[1]:
			_, err = utils.Db.Exec(s, userID, fmt.Sprintf("test_user_%d", userID), "password")
		case sqls[2]:
			_, err = utils.Db.Exec(s, userID, fmt.Sprintf("test_user_%d", userID))
		}
		if err == nil {
			return nil
		}
		lastErr = err
		if t != nil {
			t.Logf("ensureTestUser 尝试失败(%s): %v", s, err)
		}
	}
	// 如果所有尝试都失败，返回最后一个错误
	return fmt.Errorf("无法插入测试用户(尝试多种列组合)，请在测试 DB 中手动创建用户 id=%d: 最后错误: %v", userID, lastErr)
}

// cleanupTestUserByID 删除测试用户，t 可以为 nil
func cleanupTestUserByID(t *testing.T, userID int) {
	// 删除 session
	_, _ = utils.Db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	// 删除购物项（先找到 carts）
	rows, err := utils.Db.Query("SELECT id FROM carts WHERE user_id = ?", userID)
	if err == nil {
		var cartID string
		for rows.Next() {
			rows.Scan(&cartID)
			_, _ = utils.Db.Exec("DELETE FROM cart_items WHERE cart_id = ?", cartID)
		}
		rows.Close()
	}
	// 删除购物车
	_, _ = utils.Db.Exec("DELETE FROM carts WHERE user_id = ?", userID)
	// 删除该用户的订单项和订单
	orderRows, err := utils.Db.Query("SELECT id FROM orders WHERE user_id = ?", userID)
	if err == nil {
		var orderID string
		for orderRows.Next() {
			orderRows.Scan(&orderID)
			_, _ = utils.Db.Exec("DELETE FROM order_items WHERE order_id = ?", orderID)
		}
		orderRows.Close()
	}
	_, _ = utils.Db.Exec("DELETE FROM orders WHERE user_id = ?", userID)

	// 最后删除用户
	_, err = utils.Db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		if t != nil {
			t.Logf("cleanupTestUserByID 删除用户失败: %v", err)
		}
	}
}
