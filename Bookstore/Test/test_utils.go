package controller

import (
	"bookstore/dao"
	"bookstore/utils"
	"fmt"
	"strings"
	"testing"
)

// cleanupTestBook 清理测试图书
func cleanupTestBook(t *testing.T, bookID interface{}) {
	var idStr string
	switch v := bookID.(type) {
	case int:
		idStr = fmt.Sprintf("%d", v)
	case string:
		idStr = v
	default:
		if t != nil {
			t.Logf("不支持的bookID类型: %T", bookID)
		}
		return
	}

	sqlStr := "DELETE FROM books WHERE id = ?"
	_, err := utils.Db.Exec(sqlStr, idStr)
	if err != nil {
		if t != nil {
			t.Logf("清理测试图书失败: %v", err)
		}
	}
}

// cleanupTestSession 清理测试Session
func cleanupTestSession(t *testing.T, sessionID string) {
	if sessionID == "" {
		// 清理所有测试Session
		sqlStr := "DELETE FROM sessions WHERE username LIKE '%test%'"
		_, err := utils.Db.Exec(sqlStr)
		if err != nil {
			if t != nil {
				t.Logf("清理测试Session失败: %v", err)
			}
		}
	} else {
		sqlStr := "DELETE FROM sessions WHERE session_id = ?"
		_, err := utils.Db.Exec(sqlStr, sessionID)
		if err != nil {
			if t != nil {
				t.Logf("清理测试Session失败: %v", err)
			}
		}
	}
}

// cleanupTestUser 清理测试用户 (支持用户名和用户ID)
func cleanupTestUser(t *testing.T, userIdentifier interface{}) {
	var userID int
	var username string

	switch v := userIdentifier.(type) {
	case int:
		userID = v
	case string:
		username = v
		// 通过用户名获取用户ID
		user, err := dao.CheckUserName(username)
		if err == nil && user != nil {
			userID = user.ID
		}
	default:
		if t != nil {
			t.Logf("不支持的用户标识符类型: %T", userIdentifier)
		}
		return
	}

	if userID > 0 {
		// 先清理购物车相关数据
		cleanupTestCart(t, userID)

		// 清理Session
		sqlStr := "DELETE FROM sessions WHERE user_id = ?"
		utils.Db.Exec(sqlStr, userID)

		// 清理订单相关数据
		sqlStr = "DELETE FROM order_items WHERE order_id IN (SELECT id FROM orders WHERE user_id = ?)"
		utils.Db.Exec(sqlStr, userID)
		sqlStr = "DELETE FROM orders WHERE user_id = ?"
		utils.Db.Exec(sqlStr, userID)

		// 删除用户
		sqlStr = "DELETE FROM users WHERE id = ?"
		_, err := utils.Db.Exec(sqlStr, userID)
		if err != nil {
			if t != nil {
				t.Logf("清理测试用户失败: %v", err)
			}
		}
	} else if username != "" {
		// 如果无法获取用户ID，直接通过用户名删除
		sqlStr := "DELETE FROM users WHERE username = ?"
		_, err := utils.Db.Exec(sqlStr, username)
		if err != nil {
			if t != nil {
				t.Logf("清理测试用户失败: %v", err)
			}
		}
	}
}

// cleanupTestCart 清理测试购物车
func cleanupTestCart(t *testing.T, userID int) {
	// 获取用户的购物车
	cart, err := dao.GetCartByUserID(userID)
	if err != nil {
		if t != nil {
			t.Logf("获取购物车失败: %v", err)
		}
		return
	}

	if cart != nil {
		// 先删除购物项
		for _, item := range cart.CartItems {
			sqlStr := "DELETE FROM cart_items WHERE id = ?"
			utils.Db.Exec(sqlStr, item.CartItemID)
		}
		// 删除购物车
		err = dao.DeleteCartByCartID(cart.CartID)
		if err != nil {
			if t != nil {
				t.Logf("删除购物车失败: %v", err)
			}
		}
	}
}

// cleanupTestOrder 清理测试订单
func cleanupTestOrder(t *testing.T, userID int) {
	// 获取用户的订单
	orders, err := dao.GetMyOrders(userID)
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

// setupTestUser 创建测试用户
func setupTestUser(t *testing.T, username, password, email string) {
	// 先清理可能存在的用户
	cleanupTestUser(t, username)

	err := dao.SaveUser(username, password, email)
	if err != nil {
		// 如果用户已存在，尝试删除后重新创建
		if strings.Contains(err.Error(), "Duplicate entry") {
			// 先清理相关数据
			cleanupTestSession(t, "")
			// 删除用户
			sqlStr := "DELETE FROM users WHERE username = ?"
			utils.Db.Exec(sqlStr, username)
			// 重新创建
			err = dao.SaveUser(username, password, email)
			if err != nil {
				t.Fatalf("重新创建测试用户失败: %v", err)
			}
		} else {
			t.Fatalf("创建测试用户失败: %v", err)
		}
	}
}
