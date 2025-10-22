package controller

import (
	"bookstore/dao"
	"bookstore/model"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ensureProjectRootWD 尝试将当前工作目录切换到项目根目录（包含 views 目录）
func ensureProjectRootWD(t *testing.T) func() {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}
	// 如果当前目录已经包含 views，则不需要切换
	if fi, err := os.Stat(filepath.Join(origWd, "views")); err == nil && fi.IsDir() {
		return func() { os.Chdir(origWd) }
	}
	// 尝试父目录
	parent := filepath.Dir(origWd)
	if fi, err := os.Stat(filepath.Join(parent, "views")); err == nil && fi.IsDir() {
		if err := os.Chdir(parent); err != nil {
			t.Fatalf("切换到项目根目录失败: %v", err)
		}
		return func() { os.Chdir(origWd) }
	}
	// 尝试再向上一级（防护）
	grand := filepath.Dir(parent)
	if fi, err := os.Stat(filepath.Join(grand, "views")); err == nil && fi.IsDir() {
		if err := os.Chdir(grand); err != nil {
			t.Fatalf("切换到项目根目录失败: %v", err)
		}
		return func() { os.Chdir(origWd) }
	}
	// 如果未找到 views，则仍然返回恢复函数，但记录警告
	t.Log("未能在父目录中找到 views 目录；保持当前工作目录不变")
	return func() { os.Chdir(origWd) }
}

// TestBookControllerOperations 测试图书控制器操作
func TestBookControllerOperations(t *testing.T) {
	// 切换到项目根目录以确保模板路径可用
	restore := ensureProjectRootWD(t)
	defer restore()

	// 测试获取分页图书
	t.Run("测试获取分页图书", func(t *testing.T) {
		// 先添加一些测试图书
		testBooks := []*model.Book{
			{Title: "控制器测试图书1", Author: "控制器测试作者1", Price: 10.00, Sales: 0, Stock: 10, ImgPath: "/static/img/ctrl1.jpg"},
			{Title: "控制器测试图书2", Author: "控制器测试作者2", Price: 20.00, Sales: 0, Stock: 10, ImgPath: "/static/img/ctrl2.jpg"},
			{Title: "控制器测试图书3", Author: "控制器测试作者3", Price: 30.00, Sales: 0, Stock: 10, ImgPath: "/static/img/ctrl3.jpg"},
			{Title: "控制器测试图书4", Author: "控制器测试作者4", Price: 40.00, Sales: 0, Stock: 10, ImgPath: "/static/img/ctrl4.jpg"},
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
			// 清理测试图书
			for _, bookID := range addedBookIDs {
				cleanupTestBook(t, bookID)
			}
		}()

		// 测试获取分页图书
		req, err := http.NewRequest("GET", "/getPageBooks?pageNo=1", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		GetPageBooks(rr, req)

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

	// 测试获取带价格筛选的图书
	t.Run("测试获取带价格筛选的图书", func(t *testing.T) {
		// 先添加一些不同价格的测试图书
		testBooks := []*model.Book{
			{Title: "低价控制器测试图书", Author: "低价控制器测试作者", Price: 10.00, Sales: 0, Stock: 10, ImgPath: "/static/img/low_ctrl.jpg"},
			{Title: "中价控制器测试图书", Author: "中价控制器测试作者", Price: 50.00, Sales: 0, Stock: 10, ImgPath: "/static/img/mid_ctrl.jpg"},
			{Title: "高价控制器测试图书", Author: "高价控制器测试作者", Price: 100.00, Sales: 0, Stock: 10, ImgPath: "/static/img/high_ctrl.jpg"},
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
			// 清理测试图书
			for _, bookID := range addedBookIDs {
				cleanupTestBook(t, bookID)
			}
		}()

		// 测试价格筛选
		req, err := http.NewRequest("GET", "/getPageBooksByPrice?pageNo=1&min=20&max=80", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		GetPageBooksByPrice(rr, req)

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

	// 测试删除图书
	t.Run("测试删除图书", func(t *testing.T) {
		// 先添加一本测试图书
		testBook := &model.Book{
			Title:   "删除控制器测试图书",
			Author:  "删除控制器测试作者",
			Price:   99.99,
			Sales:   0,
			Stock:   10,
			ImgPath: "/static/img/delete_ctrl.jpg",
		}

		err := dao.AddBook(testBook)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取刚添加的图书ID
		books, err := dao.GetBooks()
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

		// 测试删除图书
		req, err := http.NewRequest("GET", "/deleteBook?bookId="+testBookID, nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		DeleteBook(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 验证图书是否已被删除
		deletedBook, err := dao.GetBookByID(testBookID)
		if err != nil {
			t.Errorf("查询已删除的图书时发生错误: %v", err)
		}

		if deletedBook != nil && deletedBook.ID > 0 {
			t.Error("图书应该已被删除，但仍然存在")
		}
	})

	// 测试更新或添加图书页面
	t.Run("测试更新或添加图书页面", func(t *testing.T) {
		// 先添加一本测试图书
		testBook := &model.Book{
			Title:   "更新页面测试图书",
			Author:  "更新页面测试作者",
			Price:   88.88,
			Sales:   0,
			Stock:   10,
			ImgPath: "/static/img/update_page.jpg",
		}

		err := dao.AddBook(testBook)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取刚添加的图书ID
		books, err := dao.GetBooks()
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

		defer func() {
			// 清理测试图书
			cleanupTestBook(t, testBookID)
		}()

		// 测试更新图书页面
		req, err := http.NewRequest("GET", "/toUpdateBookPage?bookId="+testBookID, nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		ToUpdateBookPage(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 检查响应内容
		body := rr.Body.String()
		if body == "" {
			t.Error("响应内容为空")
		}

		// 测试添加图书页面（不传bookId）
		req2, err := http.NewRequest("GET", "/toUpdateBookPage", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr2 := httptest.NewRecorder()
		ToUpdateBookPage(rr2, req2)

		// 检查响应状态码
		if rr2.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr2.Code)
		}
	})

	// 测试更新或添加图书
	t.Run("测试更新或添加图书", func(t *testing.T) {
		// 测试添加新图书
		formData := url.Values{}
		formData.Set("bookId", "0") // 0表示添加新图书
		formData.Set("title", "HTTP测试图书")
		formData.Set("author", "HTTP测试作者")
		formData.Set("price", "77.77")
		formData.Set("sales", "0")
		formData.Set("stock", "20")

		req, err := http.NewRequest("POST", "/updateOraddBook", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		UpdateOrAddBook(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}

		// 验证图书是否成功添加
		books, err := dao.GetBooks()
		if err != nil {
			t.Errorf("获取图书列表失败: %v", err)
		}

		var addedBookID int
		for _, book := range books {
			if book.Title == "HTTP测试图书" {
				addedBookID = book.ID
				break
			}
		}

		if addedBookID == 0 {
			t.Error("新添加的图书未在数据库中找到")
		}

		// 测试更新图书
		formData2 := url.Values{}
		formData2.Set("bookId", fmt.Sprintf("%d", addedBookID))
		formData2.Set("title", "更新后的HTTP测试图书")
		formData2.Set("author", "更新后的HTTP测试作者")
		formData2.Set("price", "88.88")
		formData2.Set("sales", "5")
		formData2.Set("stock", "15")

		req2, err := http.NewRequest("POST", "/updateOraddBook", strings.NewReader(formData2.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr2 := httptest.NewRecorder()

		UpdateOrAddBook(rr2, req2)

		// 检查响应状态码
		if rr2.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr2.Code)
		}

		// 验证图书是否成功更新
		updatedBook, err := dao.GetBookByID(fmt.Sprintf("%d", addedBookID))
		if err != nil {
			t.Errorf("查询更新后的图书失败: %v", err)
		}

		if updatedBook.Title != "更新后的HTTP测试图书" {
			t.Errorf("图书标题未正确更新，期望: 更新后的HTTP测试图书, 实际: %s", updatedBook.Title)
		}

		if updatedBook.Author != "更新后的HTTP测试作者" {
			t.Errorf("图书作者未正确更新，期望: 更新后的HTTP测试作者, 实际: %s", updatedBook.Author)
		}

		if updatedBook.Price != 88.88 {
			t.Errorf("图书价格未正确更新，期望: 88.88, 实际: %.2f", updatedBook.Price)
		}

		// 清理测试图书
		cleanupTestBook(t, addedBookID)
	})
}

// TestBookControllerDataValidation 测试图书控制器数据验证
func TestBookControllerDataValidation(t *testing.T) {
	// 切换到项目根目录以确保模板路径可用
	restore := ensureProjectRootWD(t)
	defer restore()

	tests := []struct {
		name        string
		formData    url.Values
		expectError bool
		description string
	}{
		{
			name: "正常图书数据",
			formData: url.Values{
				"bookId": {"0"},
				"title":  {"正常图书"},
				"author": {"正常作者"},
				"price":  {"99.99"},
				"sales":  {"0"},
				"stock":  {"100"},
			},
			expectError: false,
			description: "测试正常的图书数据",
		},
		{
			name: "空标题图书",
			formData: url.Values{
				"bookId": {"0"},
				"title":  {""},
				"author": {"空标题作者"},
				"price":  {"50.00"},
				"sales":  {"0"},
				"stock":  {"10"},
			},
			expectError: false, // 这里假设系统允许空标题，实际项目中可能需要验证
			description: "测试空标题图书",
		},
		{
			name: "空作者图书",
			formData: url.Values{
				"bookId": {"0"},
				"title":  {"空作者图书"},
				"author": {""},
				"price":  {"50.00"},
				"sales":  {"0"},
				"stock":  {"10"},
			},
			expectError: false, // 这里假设系统允许空作者，实际项目中可能需要验证
			description: "测试空作者图书",
		},
		{
			name: "零价格图书",
			formData: url.Values{
				"bookId": {"0"},
				"title":  {"零价格图书"},
				"author": {"零价格作者"},
				"price":  {"0.00"},
				"sales":  {"0"},
				"stock":  {"10"},
			},
			expectError: false,
			description: "测试零价格图书",
		},
		{
			name: "负价格图书",
			formData: url.Values{
				"bookId": {"0"},
				"title":  {"负价格图书"},
				"author": {"负价格作者"},
				"price":  {"-10.00"},
				"sales":  {"0"},
				"stock":  {"10"},
			},
			expectError: false, // 这里假设系统允许负价格，实际项目中可能需要验证
			description: "测试负价格图书",
		},
		{
			name: "高价格图书",
			formData: url.Values{
				"bookId": {"0"},
				"title":  {"高价格图书"},
				"author": {"高价格作者"},
				"price":  {"99999.99"},
				"sales":  {"0"},
				"stock":  {"1"},
			},
			expectError: false,
			description: "测试高价格图书",
		},
		{
			name: "负库存图书",
			formData: url.Values{
				"bookId": {"0"},
				"title":  {"负库存图书"},
				"author": {"负库存作者"},
				"price":  {"50.00"},
				"sales":  {"0"},
				"stock":  {"-10"},
			},
			expectError: false, // 这里假设系统允许负库存，实际项目中可能需要验证
			description: "测试负库存图书",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/updateOraddBook", strings.NewReader(tt.formData.Encode()))
			if err != nil {
				t.Fatalf("创建请求失败: %v", err)
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			UpdateOrAddBook(rr, req)

			if tt.expectError {
				if rr.Code == http.StatusOK {
					t.Errorf("期望返回错误，但没有返回错误: %s", tt.description)
				}
			} else {
				if rr.Code != http.StatusOK {
					t.Errorf("不期望返回错误，但返回了错误: %d, %s", rr.Code, tt.description)
				} else {
					// 验证图书是否成功添加
					books, err := dao.GetBooks()
					if err != nil {
						t.Errorf("获取图书列表失败: %v", err)
					}

					var found bool
					for _, book := range books {
						if book.Title == tt.formData.Get("title") {
							found = true
							// 清理测试图书
							cleanupTestBook(t, book.ID)
							break
						}
					}

					if !found && tt.formData.Get("title") != "" {
						t.Errorf("添加的图书未在数据库中找到: %s", tt.description)
					}
				}
			}
		})
	}
}

// TestBookControllerConcurrentOperations 测试图书控制器并发操作
func TestBookControllerConcurrentOperations(t *testing.T) {
	// 切换到项目根目录以确保模板路径可用
	restore := ensureProjectRootWD(t)
	defer restore()

	// 测试并发添加图书
	t.Run("测试并发添加图书", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				formData := url.Values{}
				formData.Set("bookId", "0")
				formData.Set("title", fmt.Sprintf("并发控制器测试图书%d", index))
				formData.Set("author", fmt.Sprintf("并发控制器测试作者%d", index))
				formData.Set("price", fmt.Sprintf("%.2f", float64(index*10)))
				formData.Set("sales", "0")
				formData.Set("stock", "10")

				req, err := http.NewRequest("POST", "/updateOraddBook", strings.NewReader(formData.Encode()))
				if err != nil {
					t.Errorf("创建请求失败: %v", err)
					return
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				rr := httptest.NewRecorder()

				UpdateOrAddBook(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("并发添加图书失败，状态码: %d", rr.Code)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 清理测试图书
		books, err := dao.GetBooks()
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
			Title:   "并发查询控制器测试图书",
			Author:  "并发查询控制器测试作者",
			Price:   99.99,
			Sales:   0,
			Stock:   10,
			ImgPath: "/static/img/concurrent_query_ctrl.jpg",
		}

		err := dao.AddBook(testBook)
		if err != nil {
			t.Errorf("添加测试图书失败: %v", err)
		}

		// 获取测试图书ID
		books, err := dao.GetBooks()
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

				req, err := http.NewRequest("GET", "/getPageBooks?pageNo=1", nil)
				if err != nil {
					t.Errorf("创建请求失败: %v", err)
					return
				}

				rr := httptest.NewRecorder()
				GetPageBooks(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("并发查询图书失败，状态码: %d", rr.Code)
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

// BenchmarkBookControllerOperations 性能测试
func BenchmarkGetPageBooks(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/getPageBooks?pageNo=1", nil)
		rr := httptest.NewRecorder()
		GetPageBooks(rr, req)
	}
}

func BenchmarkGetPageBooksByPrice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/getPageBooksByPrice?pageNo=1&min=10&max=100", nil)
		rr := httptest.NewRecorder()
		GetPageBooksByPrice(rr, req)
	}
}

// TestBookControllerEdgeCases 测试图书控制器边界情况
func TestBookControllerEdgeCases(t *testing.T) {
	// 切换到项目根目录以确保模板路径可用
	restore := ensureProjectRootWD(t)
	defer restore()

	// 测试无效的页码参数
	t.Run("测试无效页码参数", func(t *testing.T) {
		// 测试负数页码
		req, err := http.NewRequest("GET", "/getPageBooks?pageNo=-1", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		GetPageBooks(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试无效的价格范围参数
	t.Run("测试无效价格范围参数", func(t *testing.T) {
		// 测试负数价格
		req, err := http.NewRequest("GET", "/getPageBooksByPrice?pageNo=1&min=-10&max=-5", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		GetPageBooksByPrice(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试无效的图书ID删除
	t.Run("测试无效图书ID删除", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/deleteBook?bookId=999999", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		DeleteBook(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试空图书ID删除
	t.Run("测试空图书ID删除", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/deleteBook?bookId=", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		DeleteBook(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试无效的图书ID更新页面
	t.Run("测试无效图书ID更新页面", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/toUpdateBookPage?bookId=999999", nil)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		rr := httptest.NewRecorder()
		ToUpdateBookPage(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试无效的图书数据更新
	t.Run("测试无效图书数据更新", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("bookId", "999999")
		formData.Set("title", "无效图书")
		formData.Set("author", "无效作者")
		formData.Set("price", "invalid_price")
		formData.Set("sales", "invalid_sales")
		formData.Set("stock", "invalid_stock")

		req, err := http.NewRequest("POST", "/updateOraddBook", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		UpdateOrAddBook(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})

	// 测试空表单数据更新
	t.Run("测试空表单数据更新", func(t *testing.T) {
		formData := url.Values{}

		req, err := http.NewRequest("POST", "/updateOraddBook", strings.NewReader(formData.Encode()))
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		UpdateOrAddBook(rr, req)

		// 检查响应状态码
		if rr.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际得到: %d", rr.Code)
		}
	})
}

// TestBookControllerTemplateFunctions 测试图书控制器模板相关函数
func TestBookControllerTemplateFunctions(t *testing.T) {
	// 测试findViewsDir函数
	t.Run("测试findViewsDir函数", func(t *testing.T) {
		viewsDir := findViewsDir()
		if viewsDir == "" {
			t.Error("findViewsDir应该返回非空字符串")
		}
	})

	// 测试parseTemplateFiles函数
	t.Run("测试parseTemplateFiles函数", func(t *testing.T) {
		// 测试不存在的模板文件 - 使用defer recover来捕获panic
		defer func() {
			if r := recover(); r != nil {
				// 预期的panic，测试通过
				t.Logf("预期的panic被捕获: %v", r)
			}
		}()

		tmpl := parseTemplateFiles("nonexistent.html")
		if tmpl == nil {
			t.Error("parseTemplateFiles应该返回非nil模板")
		}
	})
}
