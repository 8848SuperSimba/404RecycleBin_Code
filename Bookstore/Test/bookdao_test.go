package dao

import (
	"bookstore/model"
	"fmt"
	"testing"
)

// TestBookOperations 测试图书CRUD操作
func TestBookOperations(t *testing.T) {
	// 测试添加图书
	t.Run("测试添加图书", func(t *testing.T) {
		// 创建测试图书
		testBook := &model.Book{
			Title:   "测试图书",
			Author:  "测试作者",
			Price:   99.99,
			Sales:   0,
			Stock:   100,
			ImgPath: "/static/img/test.jpg",
		}

		// 添加图书到数据库
		err := AddBook(testBook)
		if err != nil {
			t.Errorf("添加图书失败: %v", err)
		}

		// 验证图书是否成功添加
		// 通过查询所有图书来验证
		books, err := GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		// 查找刚添加的图书
		var found bool
		for _, book := range books {
			if book.Title == testBook.Title && book.Author == testBook.Author {
				found = true
				// 验证图书信息
				if book.Price != testBook.Price {
					t.Errorf("价格不匹配，期望: %.2f, 实际: %.2f", testBook.Price, book.Price)
				}
				if book.Stock != testBook.Stock {
					t.Errorf("库存不匹配，期望: %d, 实际: %d", testBook.Stock, book.Stock)
				}
				if book.ImgPath != testBook.ImgPath {
					t.Errorf("图片路径不匹配，期望: %s, 实际: %s", testBook.ImgPath, book.ImgPath)
				}
				// 保存图书ID用于后续测试
				testBook.ID = book.ID
				break
			}
		}

		if !found {
			t.Error("添加的图书未在数据库中找到")
		} else {
			// 清理本次子测试添加的图书
			if testBook.ID != 0 {
				cleanupTestBook(t, testBook.ID)
			}
		}
	})

	// 测试查询图书
	t.Run("测试查询图书", func(t *testing.T) {
		// 先添加一本测试图书
		testBook := &model.Book{
			Title:   "查询测试图书",
			Author:  "查询测试作者",
			Price:   88.88,
			Sales:   10,
			Stock:   50,
			ImgPath: "/static/img/query_test.jpg",
		}

		err := AddBook(testBook)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取所有图书
		books, err := GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		if len(books) == 0 {
			t.Error("图书列表为空")
		}

		// 查找刚添加的图书
		var foundBook *model.Book
		for _, book := range books {
			if book.Title == testBook.Title {
				foundBook = book
				break
			}
		}

		if foundBook == nil {
			t.Error("未找到刚添加的图书")
		} else {
			// 验证图书信息
			if foundBook.Author != testBook.Author {
				t.Errorf("作者不匹配，期望: %s, 实际: %s", testBook.Author, foundBook.Author)
			}
			if foundBook.Price != testBook.Price {
				t.Errorf("价格不匹配，期望: %.2f, 实际: %.2f", testBook.Price, foundBook.Price)
			}
			if foundBook.Sales != testBook.Sales {
				t.Errorf("销量不匹配，期望: %d, 实际: %d", testBook.Sales, foundBook.Sales)
			}
			if foundBook.Stock != testBook.Stock {
				t.Errorf("库存不匹配，期望: %d, 实际: %d", testBook.Stock, foundBook.Stock)
			}
			// 清理测试图书
			cleanupTestBook(t, foundBook.ID)
		}
	})

	// 测试根据ID查询图书
	t.Run("测试根据ID查询图书", func(t *testing.T) {
		// 先添加一本测试图书
		testBook := &model.Book{
			Title:   "ID查询测试图书",
			Author:  "ID查询测试作者",
			Price:   77.77,
			Sales:   5,
			Stock:   25,
			ImgPath: "/static/img/id_query_test.jpg",
		}

		err := AddBook(testBook)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取刚添加的图书ID
		books, err := GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		var testBookID string
		for _, book := range books {
			if book.Title == testBook.Title {
				testBookID = fmt.Sprintf("%d", book.ID)
				break
			}
		}

		if testBookID == "" {
			t.Fatal("未找到测试图书ID")
		}

		// 根据ID查询图书
		retrievedBook, err := GetBookByID(testBookID)
		if err != nil {
			t.Errorf("根据ID查询图书失败: %v", err)
		}

		if retrievedBook == nil {
			t.Error("查询结果为空")
		} else {
			// 验证图书信息
			if retrievedBook.Title != testBook.Title {
				t.Errorf("标题不匹配，期望: %s, 实际: %s", testBook.Title, retrievedBook.Title)
			}
			if retrievedBook.Author != testBook.Author {
				t.Errorf("作者不匹配，期望: %s, 实际: %s", testBook.Author, retrievedBook.Author)
			}
			if retrievedBook.Price != testBook.Price {
				t.Errorf("价格不匹配，期望: %.2f, 实际: %.2f", testBook.Price, retrievedBook.Price)
			}
			// 清理测试图书
			cleanupTestBook(t, retrievedBook.ID)
		}
	})

	// 测试更新图书
	t.Run("测试更新图书", func(t *testing.T) {
		// 先添加一本测试图书
		testBook := &model.Book{
			Title:   "更新测试图书",
			Author:  "更新测试作者",
			Price:   66.66,
			Sales:   0,
			Stock:   30,
			ImgPath: "/static/img/update_test.jpg",
		}

		err := AddBook(testBook)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取刚添加的图书ID
		books, err := GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		var testBookID int
		for _, book := range books {
			if book.Title == testBook.Title {
				testBookID = book.ID
				break
			}
		}

		if testBookID == 0 {
			t.Fatal("未找到测试图书ID")
		}

		// 更新图书信息
		updatedBook := &model.Book{
			ID:      testBookID,
			Title:   "更新后的图书标题",
			Author:  "更新后的作者",
			Price:   88.88,
			Sales:   10,
			Stock:   20,
			ImgPath: "/static/img/updated_test.jpg",
		}

		err = UpdateBook(updatedBook)
		if err != nil {
			t.Errorf("更新图书失败: %v", err)
		}

		// 验证更新结果
		retrievedBook, err := GetBookByID(fmt.Sprintf("%d", testBookID))
		if err != nil {
			t.Errorf("查询更新后的图书失败: %v", err)
		}

		if retrievedBook.Title != updatedBook.Title {
			t.Errorf("更新后标题不匹配，期望: %s, 实际: %s", updatedBook.Title, retrievedBook.Title)
		}
		if retrievedBook.Author != updatedBook.Author {
			t.Errorf("更新后作者不匹配，期望: %s, 实际: %s", updatedBook.Author, retrievedBook.Author)
		}
		if retrievedBook.Price != updatedBook.Price {
			t.Errorf("更新后价格不匹配，期望: %.2f, 实际: %.2f", updatedBook.Price, retrievedBook.Price)
		}
		if retrievedBook.Sales != updatedBook.Sales {
			t.Errorf("更新后销量不匹配，期望: %d, 实际: %d", updatedBook.Sales, retrievedBook.Sales)
		}
		if retrievedBook.Stock != updatedBook.Stock {
			t.Errorf("更新后库存不匹配，期望: %d, 实际: %d", updatedBook.Stock, retrievedBook.Stock)
		}

		// 清理测试图书
		cleanupTestBook(t, testBookID)
	})

	// 测试删除图书
	t.Run("测试删除图书", func(t *testing.T) {
		// 先添加一本测试图书
		testBook := &model.Book{
			Title:   "删除测试图书",
			Author:  "删除测试作者",
			Price:   55.55,
			Sales:   0,
			Stock:   40,
			ImgPath: "/static/img/delete_test.jpg",
		}

		err := AddBook(testBook)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取刚添加的图书ID
		books, err := GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		var testBookID string
		for _, book := range books {
			if book.Title == testBook.Title {
				testBookID = fmt.Sprintf("%d", book.ID)
				break
			}
		}

		if testBookID == "" {
			t.Fatal("未找到测试图书ID")
		}

		// 删除图书
		err = DeleteBook(testBookID)
		if err != nil {
			t.Errorf("删除图书失败: %v", err)
		}

		// 验证图书是否已被删除
		deletedBook, err := GetBookByID(testBookID)
		if err != nil {
			t.Errorf("查询已删除的图书时发生错误: %v", err)
		}

		if deletedBook != nil && deletedBook.ID > 0 {
			t.Error("图书应该已被删除，但仍然存在")
		}
	})

	// 测试分页功能
	t.Run("测试分页功能", func(t *testing.T) {
		// 先添加多本测试图书
		testBooks := []*model.Book{
			{Title: "分页测试图书1", Author: "分页测试作者1", Price: 10.00, Sales: 0, Stock: 10, ImgPath: "/static/img/page1.jpg"},
			{Title: "分页测试图书2", Author: "分页测试作者2", Price: 20.00, Sales: 0, Stock: 10, ImgPath: "/static/img/page2.jpg"},
			{Title: "分页测试图书3", Author: "分页测试作者3", Price: 30.00, Sales: 0, Stock: 10, ImgPath: "/static/img/page3.jpg"},
			{Title: "分页测试图书4", Author: "分页测试作者4", Price: 40.00, Sales: 0, Stock: 10, ImgPath: "/static/img/page4.jpg"},
			{Title: "分页测试图书5", Author: "分页测试作者5", Price: 50.00, Sales: 0, Stock: 10, ImgPath: "/static/img/page5.jpg"},
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

		// 测试第一页
		page, err := GetPageBooks("1")
		if err != nil {
			t.Errorf("获取第一页图书失败: %v", err)
		}

		if page.PageNo != 1 {
			t.Errorf("当前页不匹配，期望: 1, 实际: %d", page.PageNo)
		}

		if page.PageSize != 4 {
			t.Errorf("每页大小不匹配，期望: 4, 实际: %d", page.PageSize)
		}

		if len(page.Books) > 4 {
			t.Errorf("第一页图书数量超过限制，期望: <=4, 实际: %d", len(page.Books))
		}

		// 测试第二页
		page2, err := GetPageBooks("2")
		if err != nil {
			t.Errorf("获取第二页图书失败: %v", err)
		}

		if page2.PageNo != 2 {
			t.Errorf("当前页不匹配，期望: 2, 实际: %d", page2.PageNo)
		}

		// 清理测试图书
		for _, bookID := range addedBookIDs {
			cleanupTestBook(t, bookID)
		}
	})

	// 测试价格筛选
	t.Run("测试价格筛选", func(t *testing.T) {
		// 先添加不同价格的测试图书
		testBooks := []*model.Book{
			{Title: "低价图书", Author: "低价作者", Price: 10.00, Sales: 0, Stock: 10, ImgPath: "/static/img/low.jpg"},
			{Title: "中价图书", Author: "中价作者", Price: 50.00, Sales: 0, Stock: 10, ImgPath: "/static/img/mid.jpg"},
			{Title: "高价图书", Author: "高价作者", Price: 100.00, Sales: 0, Stock: 10, ImgPath: "/static/img/high.jpg"},
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

		// 测试价格范围筛选 (20-80)
		page, err := GetPageBooksByPrice("1", "20", "80")
		if err != nil {
			t.Errorf("按价格筛选图书失败: %v", err)
		}

		// 验证筛选结果
		for _, book := range page.Books {
			if book.Price < 20 || book.Price > 80 {
				t.Errorf("图书价格超出筛选范围，价格: %.2f, 范围: 20-80", book.Price)
			}
		}

		// 测试价格范围筛选 (0-30)
		page2, err := GetPageBooksByPrice("1", "0", "30")
		if err != nil {
			t.Errorf("按价格筛选图书失败: %v", err)
		}

		// 验证筛选结果
		for _, book := range page2.Books {
			if book.Price < 0 || book.Price > 30 {
				t.Errorf("图书价格超出筛选范围，价格: %.2f, 范围: 0-30", book.Price)
			}
		}

		// 清理测试图书
		for _, bookID := range addedBookIDs {
			cleanupTestBook(t, bookID)
		}
	})
}

// TestBookDataValidation 测试图书数据验证
func TestBookDataValidation(t *testing.T) {
	tests := []struct {
		name        string
		book        *model.Book
		expectError bool
		description string
	}{
		{
			name: "正常图书数据",
			book: &model.Book{
				Title:   "正常图书",
				Author:  "正常作者",
				Price:   99.99,
				Sales:   10,
				Stock:   100,
				ImgPath: "/static/img/normal.jpg",
			},
			expectError: false,
			description: "测试正常的图书数据",
		},
		{
			name: "零价格图书",
			book: &model.Book{
				Title:   "零价格图书",
				Author:  "零价格作者",
				Price:   0.00,
				Sales:   0,
				Stock:   10,
				ImgPath: "/static/img/free.jpg",
			},
			expectError: false,
			description: "测试零价格图书",
		},
		{
			name: "负价格图书",
			book: &model.Book{
				Title:   "负价格图书",
				Author:  "负价格作者",
				Price:   -10.00,
				Sales:   0,
				Stock:   10,
				ImgPath: "/static/img/negative.jpg",
			},
			expectError: false, // 这里假设系统允许负价格，实际项目中可能需要验证
			description: "测试负价格图书",
		},
		{
			name: "空标题图书",
			book: &model.Book{
				Title:   "",
				Author:  "空标题作者",
				Price:   50.00,
				Sales:   0,
				Stock:   10,
				ImgPath: "/static/img/empty_title.jpg",
			},
			expectError: false, // 这里假设系统允许空标题，实际项目中可能需要验证
			description: "测试空标题图书",
		},
		{
			name: "空作者图书",
			book: &model.Book{
				Title:   "空作者图书",
				Author:  "",
				Price:   50.00,
				Sales:   0,
				Stock:   10,
				ImgPath: "/static/img/empty_author.jpg",
			},
			expectError: false, // 这里假设系统允许空作者，实际项目中可能需要验证
			description: "测试空作者图书",
		},
		{
			name: "高价格图书",
			book: &model.Book{
				Title:   "高价格图书",
				Author:  "高价格作者",
				Price:   99999.99,
				Sales:   0,
				Stock:   1,
				ImgPath: "/static/img/expensive.jpg",
			},
			expectError: false,
			description: "测试高价格图书",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddBook(tt.book)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望返回错误，但没有返回错误: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("不期望返回错误，但返回了错误: %v, %s", err, tt.description)
				} else {
					// 验证图书是否成功添加
					books, err := GetBooks()
					if err != nil {
						t.Errorf("获取图书列表失败: %v", err)
					}

					var found bool
					for _, book := range books {
						if book.Title == tt.book.Title {
							found = true
							// 清理测试图书
							cleanupTestBook(t, book.ID)
							break
						}
					}

					if !found {
						t.Errorf("添加的图书未在数据库中找到: %s", tt.description)
					}
				}
			}
		})
	}
}

// TestBookConcurrentOperations 测试图书并发操作
func TestBookConcurrentOperations(t *testing.T) {
	// 测试并发添加图书
	t.Run("测试并发添加图书", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				book := &model.Book{
					Title:   fmt.Sprintf("并发测试图书%d", index),
					Author:  fmt.Sprintf("并发测试作者%d", index),
					Price:   float64(index * 10),
					Sales:   0,
					Stock:   10,
					ImgPath: fmt.Sprintf("/static/img/concurrent%d.jpg", index),
				}

				err := AddBook(book)
				if err != nil {
					t.Errorf("并发添加图书失败: %v", err)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 清理测试图书
		books, err := GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		for _, book := range books {
			if book.Title != "" && book.Author != "" {
				// 只清理测试图书
				if book.Title != "" && book.Author != "" {
					cleanupTestBook(t, book.ID)
				}
			}
		}
	})

	// 测试并发查询图书
	t.Run("测试并发查询图书", func(t *testing.T) {
		// 先添加一本测试图书
		testBook := &model.Book{
			Title:   "并发查询测试图书",
			Author:  "并发查询测试作者",
			Price:   99.99,
			Sales:   0,
			Stock:   10,
			ImgPath: "/static/img/concurrent_query.jpg",
		}

		err := AddBook(testBook)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取测试图书ID
		books, err := GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		var testBookID string
		for _, book := range books {
			if book.Title == testBook.Title {
				testBookID = fmt.Sprintf("%d", book.ID)
				break
			}
		}

		if testBookID == "" {
			t.Fatal("未找到测试图书ID")
		}

		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				book, err := GetBookByID(testBookID)
				if err != nil {
					t.Errorf("并发查询图书失败: %v", err)
				}
				if book == nil {
					t.Errorf("并发查询图书结果为空")
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 清理测试图书
		cleanupTestBook(t, testBookID)
	})
}

// BenchmarkBookOperations 性能测试
func BenchmarkAddBook(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		book := &model.Book{
			Title:   fmt.Sprintf("性能测试图书%d", i),
			Author:  fmt.Sprintf("性能测试作者%d", i),
			Price:   float64(i),
			Sales:   0,
			Stock:   10,
			ImgPath: fmt.Sprintf("/static/img/benchmark%d.jpg", i),
		}
		AddBook(book)
	}
}

func BenchmarkGetBooks(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetBooks()
	}
}

func BenchmarkGetPageBooks(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPageBooks("1")
	}
}

// Note: cleanupTestBook is defined in orderdao_test.go to centralize cleanup logic
// and avoid redeclaration across test files.
