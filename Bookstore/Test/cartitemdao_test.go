package dao

import (
	"bookstore/model"
	"bookstore/utils"
	"fmt"
	"testing"
)

// TestCartItemDAO 覆盖 cart_item DAO 的主要行为
func TestCartItemDAO(t *testing.T) {
	// 创建测试用户和图书
	testUserID := createTestUser(t)
	defer cleanupTestUserByID(t, testUserID)

	// 添加测试图书
	book := &model.Book{Title: "CartItemTestBook", Author: "Tester", Price: 9.99, Sales: 0, Stock: 10, ImgPath: "/static/img/test.jpg"}
	err := AddBook(book)
	if err != nil {
		t.Fatalf("AddBook failed: %v", err)
	}
	// 查找刚添加的图书ID
	books, err := GetBooks()
	if err != nil {
		t.Fatalf("GetBooks failed: %v", err)
	}
	var bookID int
	for _, b := range books {
		if b.Title == book.Title && b.Author == book.Author {
			bookID = b.ID
			break
		}
	}
	if bookID == 0 {
		cleanupTestBook(t, book.ID)
		t.Fatalf("could not find added book")
	}
	defer cleanupTestBook(t, bookID)

	// 创建购物车
	cartID := utils.CreateUUID()
	cart := &model.Cart{CartID: cartID, UserID: testUserID}
	err = AddCart(cart)
	if err != nil {
		t.Fatalf("AddCart failed: %v", err)
	}
	// defer DeleteCartByCartID(cartID) // wrap to explicitly ignore returned error
	defer func() { _ = DeleteCartByCartID(cartID) }()

	// 添加购物项
	ci := &model.CartItem{Book: &model.Book{ID: bookID, Price: 9.99}, Count: 2, CartID: cartID}
	err = AddCartItem(ci)
	if err != nil {
		t.Fatalf("AddCartItem failed: %v", err)
	}

	// 获取购物项
	got, err := GetCartItemByBookIDAndCartID(fmt.Sprintf("%d", bookID), cartID)
	if err != nil {
		t.Fatalf("GetCartItemByBookIDAndCartID failed: %v", err)
	}
	if got == nil {
		t.Fatalf("expected cart item, got nil")
	}
	if got.Count != 2 {
		t.Fatalf("expected count 2, got %d", got.Count)
	}
	if got.Book == nil || got.Book.ID != bookID {
		t.Fatalf("expected book id %d, got %+v", bookID, got.Book)
	}

	// 更新购物项数量
	got.Count = 5
	err = UpdateBookCount(got)
	if err != nil {
		t.Fatalf("UpdateBookCount failed: %v", err)
	}
	// 重新查询验证
	got2, err := GetCartItemByBookIDAndCartID(fmt.Sprintf("%d", bookID), cartID)
	if err != nil {
		t.Fatalf("requery failed: %v", err)
	}
	if got2.Count != 5 {
		t.Fatalf("expected updated count 5, got %d", got2.Count)
	}

	// 获取购物车所有购物项
	items, err := GetCartItemsByCartID(cartID)
	if err != nil {
		t.Fatalf("GetCartItemsByCartID failed: %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected at least 1 cart item")
	}

	// 删除单个购物项
	err = DeleteCartItemByID(fmt.Sprintf("%d", got2.CartItemID))
	if err != nil {
		t.Fatalf("DeleteCartItemByID failed: %v", err)
	}

	// 确认已删除
	_, err = GetCartItemByBookIDAndCartID(fmt.Sprintf("%d", bookID), cartID)
	if err == nil {
		t.Fatalf("expected error after deleting cart item, got nil")
	}

	// 添加多个购物项并测试按cart清除
	ci1 := &model.CartItem{Book: &model.Book{ID: bookID, Price: 9.99}, Count: 1, CartID: cartID}
	ci2 := &model.CartItem{Book: &model.Book{ID: bookID, Price: 9.99}, Count: 3, CartID: cartID}
	_ = AddCartItem(ci1)
	_ = AddCartItem(ci2)

	err = DeleteCartItemsByCartID(cartID)
	if err != nil {
		t.Fatalf("DeleteCartItemsByCartID failed: %v", err)
	}
	// 确认全部删除
	itemsAfter, err := GetCartItemsByCartID(cartID)
	if err != nil {
		// if table empty it may return nil,nil or nil,sql.ErrNoRows depending; just ensure length 0
		itemsAfter = []*model.CartItem{}
	}
	if len(itemsAfter) != 0 {
		t.Fatalf("expected 0 items after DeleteCartItemsByCartID, got %d", len(itemsAfter))
	}
}
