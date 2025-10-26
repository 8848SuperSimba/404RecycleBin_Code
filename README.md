# 404书城 - 网上书店管理系统

## 项目概述

404书城是一个基于Go语言开发的Web书店管理系统，提供了完整的图书在线购买功能。该项目采用前后端结合的方式，使用Go语言作为后端服务，MySQL作为数据库，HTML+CSS+JavaScript构建前端界面，实现了用户注册登录、图书浏览购买、购物车管理、订单处理等核心电商功能。

## 技术架构

### 后端技术栈
- **编程语言**: Go 1.25.2
- **Web框架**: 标准库 net/http
- **数据库**: MySQL 5.7+
- **数据库驱动**: github.com/go-sql-driver/mysql v1.9.3
- **模板引擎**: Go标准库 html/template
- **会话管理**: Cookie + 数据库Session

### 前端技术栈
- **HTML**: 用于页面结构
- **CSS**: 自定义样式表（style.css）
- **JavaScript**: jQuery 1.7.2 用于交互处理和Ajax请求
- **模板引擎**: Go Template（{{ }}语法）
- **响应式交互**: Ajax异步请求，无刷新更新页面

### 系统架构
项目采用经典的**MVC三层架构**模式：
```
Model (模型层)   - 数据模型定义，业务逻辑
View (视图层)    - HTML模板，前端展示
Controller (控制层) - HTTP请求处理，业务协调
DAO (数据访问层)   - 数据库操作抽象
```

## 项目结构

```
Bookstore/
├── main.go                 # 应用入口，路由配置
├── controller/             # 控制器层
│   ├── userhandler.go     # 用户相关功能（登录、注册、注销）
│   ├── bookhandler.go     # 图书相关功能（查询、分页、增删改）
│   ├── carthandler.go     # 购物车功能
│   └── orderhandler.go    # 订单管理功能
├── model/                 # 数据模型层
│   ├── user.go           # 用户模型
│   ├── book.go           # 图书模型
│   ├── cart.go           # 购物车模型
│   ├── cartItem.go       # 购物项模型
│   ├── order.go          # 订单模型
│   ├── orderItem.go      # 订单项模型
│   ├── session.go        # 会话模型
│   ├── page.go           # 分页模型
│   └── json.go           # Ajax响应数据结构
├── dao/                   # 数据访问层
│   ├── userdao.go        # 用户数据库操作
│   ├── userdao_test.go   # 用户数据访问测试（早期测试文件）
│   ├── bookdao.go        # 图书数据库操作
│   ├── cartdao.go        # 购物车数据库操作
│   ├── cartItemdao.go    # 购物项数据库操作
│   ├── orderdao.go       # 订单数据库操作
│   ├── orderItemdao.go   # 订单项数据库操作
│   └── sessiondao.go     # 会话数据库操作
├── utils/                 # 工具类
│   ├── db.go             # 数据库连接
│   └── uuid.go           # UUID生成工具
├── views/                 # 前端视图
│   ├── index.html        # 首页（图书展示）
│   ├── static/           # 静态资源
│   │   ├── css/style.css        # 样式文件
│   │   ├── img/          # 图片资源（logo.gif, Go.jpg, python.jpg, mysql.jpg, js.jpg, default.jpg）
│   │   └── script/jquery-1.7.2.js  # jQuery库
│   └── pages/            # 功能页面
│       ├── user/         # 用户页面（登录、注册、登录成功、注册成功）
│       ├── cart/         # 购物车页面（购物车、结账）
│       ├── manager/      # 管理员页面（后台管理、图书管理、图书编辑）
│       └── order/        # 订单页面（我的订单、订单管理、订单详情）
├── Test/                 # 单元测试
└── sql.sql               # 数据库初始化脚本
```

## 数据库设计

### 数据库表结构

#### 1. 用户表 (users)
```sql
CREATE TABLE users(
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,  -- 用户名（唯一）
    password VARCHAR(100) NOT NULL,         -- 密码
    email VARCHAR(100) NOT NULL UNIQUE     -- 邮箱（唯一）
);
```

#### 2. 图书表 (books)
```sql
CREATE TABLE books(
    id INT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,     -- 书名
    author VARCHAR(100) NOT NULL,    -- 作者
    price DOUBLE(11,2) NOT NULL,     -- 价格
    sales INT NOT NULL,              -- 销量
    stock INT NOT NULL,              -- 库存
    img_path VARCHAR(100)            -- 图片路径
);
```

#### 3. 会话表 (sessions)
```sql
CREATE TABLE sessions(
    session_id VARCHAR(100) PRIMARY KEY, -- 会话ID（UUID）
    username VARCHAR(100) NOT NULL,      -- 用户名
    user_id INT NOT NULL,                -- 用户ID（外键）
    FOREIGN KEY(user_id) REFERENCES users(id)
);
```

#### 4. 购物车表 (carts)
```sql
CREATE TABLE carts(
    id VARCHAR(100) PRIMARY KEY,          -- 购物车ID（UUID）
    total_count INT NOT NULL,             -- 商品总数
    total_amount DOUBLE(11,2) NOT NULL,   -- 总金额
    user_id INT NOT NULL,                 -- 用户ID（外键）
    FOREIGN KEY(user_id) REFERENCES users(id)
);
```

#### 5. 购物项表 (cart_items)
```sql
CREATE TABLE cart_items(
    id INT PRIMARY KEY AUTO_INCREMENT,
    count INT NOT NULL,                   -- 商品数量
    amount DOUBLE(11,2) NOT NULL,         -- 小计金额
    book_id INT NOT NULL,                 -- 图书ID（外键）
    cart_id VARCHAR(100) NOT NULL,        -- 购物车ID（外键）
    FOREIGN KEY(book_id) REFERENCES books(id),
    FOREIGN KEY(cart_id) REFERENCES carts(id)
);
```

#### 6. 订单表 (orders)
```sql
CREATE TABLE orders(
    id VARCHAR(100) PRIMARY KEY,          -- 订单号（UUID）
    create_time DATETIME NOT NULL,        -- 创建时间
    total_count INT NOT NULL,              -- 商品总数
    total_amount DOUBLE(11,2) NOT NULL,    -- 总金额
    state INT NOT NULL,                   -- 订单状态（0未发货，1已发货，2已完成）
    user_id INT,                          -- 用户ID（外键）
    FOREIGN KEY(user_id) REFERENCES users(id)
);
```

#### 7. 订单项表 (order_items)
```sql
CREATE TABLE order_items(
    id INT PRIMARY KEY AUTO_INCREMENT,
    count INT NOT NULL,                   -- 商品数量
    amount DOUBLE(11,2) NOT NULL,         -- 小计金额
    title VARCHAR(100) NOT NULL,          -- 书名
    author VARCHAR(100) NOT NULL,         -- 作者
    price DOUBLE(11,2) NOT NULL,          -- 价格
    img_path VARCHAR(100) NOT NULL,       -- 图片路径
    order_id VARCHAR(100) NOT NULL,       -- 订单ID（外键）
    FOREIGN KEY(order_id) REFERENCES orders(id)
);
```

## 核心功能

### 1. 用户管理模块

#### 用户注册 (Regist)
- **路径**: `/regist`
- **方法**: POST
- **功能**: 
  - 验证用户名唯一性（Ajax实时验证）
  - 验证邮箱唯一性
  - 用户信息持久化
  - 注册成功跳转到登录页面

#### 用户登录 (Login)
- **路径**: `/login`
- **方法**: POST
- **功能**:
  - 用户名密码验证
  - 生成UUID作为Session ID
  - 创建Cookie关联Session
  - 登录成功后显示欢迎信息

#### 用户注销 (Logout)
- **路径**: `/logout`
- **功能**:
  - 删除数据库中的Session记录
  - 使Cookie失效
  - 重定向到首页

#### 用户名验证 (CheckUserName)
- **路径**: `/checkUserName`
- **方法**: POST (Ajax)
- **功能**: 
  - 异步验证用户名是否可用
  - 实时反馈提示信息（"用户名已存在！"或"用户名可用！"）
  - 使用Ajax技术无需刷新页面

### 2. 图书管理模块

#### 图书浏览 - 首页 (GetPageBooksByPrice)
- **路径**: `/main`
- **功能**:
  - 支持分页显示图书（每页4条）
  - 支持按价格范围筛选
  - 显示图书详细信息（标题、作者、价格、销量、库存）
  - 显示库存状态，缺货时提示"小二拼命补货中..."
  - 登录用户可添加到购物车

#### 图书管理 - 后台 (GetPageBooks)
- **路径**: `/getPageBooks`
- **功能**:
  - 分页显示所有图书
  - 支持添加、修改、删除图书
  - 图书信息包括：标题、作者、价格、销量、库存

#### 添加/修改图书 (UpdateOrAddBook)
- **路径**: `/updateOraddBook`
- **方法**: POST
- **功能**:
  - 根据bookId判断是添加还是修改
  - 更新图书信息（标题、作者、价格、销量、库存）
  - 新增或更新后刷新图书列表

#### 删除图书 (DeleteBook)
- **路径**: `/deleteBook`
- **功能**: 根据bookId删除指定图书

#### 跳转到编辑页面 (ToUpdateBookPage)
- **路径**: `/toUpdateBookPage?bookId=xxx`
- **功能**: 跳转到图书编辑页面，支持新增和编辑两种模式

### 3. 购物车模块

#### 添加商品到购物车 (AddBook2Cart)
- **路径**: `/addBook2Cart`
- **方法**: POST (Ajax)
- **功能**:
  - 验证用户登录状态
  - 判断用户是否已有购物车，若无则创建
  - 检查购物车中是否已有该商品
  - 已有则数量+1，无则创建新购物项
  - 更新购物车总数量和总金额
  - Ajax返回添加成功提示

#### 查看购物车 (GetCartInfo)
- **路径**: `/getCartInfo`
- **功能**:
  - 显示购物车中所有商品
  - 显示每件商品的数量、单价、小计
  - 支持修改商品数量（Ajax异步更新）
  - 支持删除购物项
  - 显示购物车总数量和总金额
  - 提供结账、清空购物车、继续购物等功能

#### 删除购物项 (DeleteCartItem)
- **路径**: `/deleteCartItem?cartItemId=xxx`
- **功能**:
  - 从购物车中删除指定商品
  - 更新购物车总数量和总金额
  - 自动刷新购物车

#### 更新购物项数量 (UpdateCartItem)
- **路径**: `/updateCartItem`
- **方法**: POST (Ajax)
- **功能**:
  - 修改商品数量时异步更新
  - 实时更新小计金额、总数量和总金额
  - 无需刷新页面即可看到更改结果

#### 清空购物车 (DeleteCart)
- **路径**: `/deleteCart?cartId=xxx`
- **功能**: 清空购物车中所有商品

### 4. 订单管理模块

#### 结账 (Checkout)
- **路径**: `/checkout`
- **功能**:
  - 生成唯一订单号（UUID）
  - 创建订单并保存到数据库
  - 将购物车中的商品转换为订单项（保存商品快照）
  - 更新图书库存和销量
  - 清空购物车
  - 跳转到订单确认页面（checkout.html）
  - 显示订单号

#### 获取所有订单 (GetOrders)
- **路径**: `/getOrders`
- **功能**: 管理员查看所有订单（订单管理后台）

#### 我的订单 (GetMyOrders)
- **路径**: `/getMyOrder`
- **功能**:
  - 用户查看自己的所有订单
  - 显示订单号、创建时间、商品数量、总金额
  - 显示订单状态：未发货、已发货、已完成
  - 支持查看订单详情

#### 订单详情 (GetOrderInfo)
- **路径**: `/getOrderInfo?orderId=xxx`
- **功能**: 查看订单的详细商品信息

#### 发货 (SendOrder)
- **路径**: `/sendOrder?orderId=xxx`
- **功能**: 管理员将订单状态从"未发货"更新为"已发货"

#### 确认收货 (TakeOrder)
- **路径**: `/takeOrder?orderId=xxx`
- **功能**: 用户确认收货，将订单状态更新为"已完成"

## 业务逻辑设计

### Session会话管理
1. **登录时**: 生成UUID作为Session ID，保存用户信息到数据库sessions表
2. **Cookie设置**: 将Session ID存入Cookie，HttpOnly属性增强安全性
3. **请求验证**: 每次请求通过Cookie获取Session ID，查询数据库验证登录状态
4. **注销时**: 删除数据库Session记录，使Cookie失效

### 购物车业务逻辑
1. **创建购物车**: 用户首次添加商品时，自动创建购物车（UUID作为ID）
2. **商品去重**: 检查购物车中是否已有该商品，有则更新数量，无则创建新项
3. **自动计算**: 通过模型方法GetTotalCount()和GetTotalAmount()计算总数量和总金额
4. **库存管理**: 结账时自动扣减库存，增加销量

### 订单状态流转
```
订单创建 (state=0) → 管理员发货 (state=1) → 用户确认收货 (state=2)
    ↓                   ↓                      ↓
  未发货              已发货                  交易完成
```

### 分页功能
- **每页显示**: 4条记录
- **总页数计算**: `(总记录数 - 1) / 每页记录数 + 1`
- **支持跳转**: 可直接跳转到指定页码
- **价格筛选**: 支持按价格范围筛选并保持分页

## 技术特点

### 1. MVC架构优势
- **清晰的职责分离**: Model定义数据结构，View负责展示，Controller处理业务逻辑
- **易于维护**: 代码结构清晰，便于功能扩展
- **可测试性**: 各层独立，便于单元测试
- **松耦合**: DAO层独立，数据访问与业务逻辑分离

### 2. DAO数据访问层
- **数据操作封装**: 所有SQL操作集中在DAO层，便于数据库切换
- **代码复用**: 相同的数据操作可在多处复用
- **维护便利**: SQL语句集中管理，易于优化和修改

### 3. 模板渲染
- **动态内容**: 使用Go Template语法{{ }}实现动态内容渲染
- **条件渲染**: {{if .IsLogin}}、{{range .Books}}等实现动态显示
- **数据绑定**: 将后端数据绑定到前端模板

### 4. Ajax异步交互
- **实时反馈**: 添加购物车、修改数量无需刷新页面
- **用户体验**: 实时显示操作结果，提升交互体验
- **性能优化**: 减少页面刷新，节省带宽
- **Ajax响应**: 使用model.Data结构返回JSON格式数据（金额、总金额、总数量）

### 5. 安全性考虑
- **Cookie设置**: HttpOnly属性防止XSS攻击
- **Session管理**: 服务器端Session验证，增强安全性
- **用户验证**: 登录状态验证，保护用户操作

## 项目亮点

### 1. 完整的电商流程
从用户注册、商品浏览、购物车管理到订单生成，实现了完整的电商业务流程。

### 2. 前后端分离思想
虽然采用模板渲染，但Ajax请求体现了前后端分离的现代Web开发思想。

### 3. 数据库设计规范
- 外键约束保证数据完整性
- 合理的表结构设计，支持级联操作
- 订单快照机制（order_items保存商品快照，避免历史订单信息丢失）
- 索引优化（主键、唯一键）
- 数据类型合理：DOUBLE(11,2)精确处理金额，DATETIME记录时间

### 4. 用户体验优化
- 实时价格筛选
- 库存状态提示
- 购物车数量实时更新
- 订单状态可视化

### 5. 管理功能完善
- 管理员图书管理（增删改查）
- 订单管理（发货、查看详情）
- 分页管理提高操作效率
- 独立的后台管理系统入口

## 部署说明

### 环境要求
- Go 1.25.2+
- MySQL 5.7+
- 操作系统：Windows/Linux/Mac

### 数据库配置
修改 `utils/db.go` 中的数据库连接信息：
```go
Db, err = sql.Open("mysql", "root:123456@tcp(localhost:3306)/bookstore")
```

### 初始化数据库
执行 `sql.sql` 文件创建数据库和表：
```bash
mysql -u root -p < sql.sql
```

### 运行项目
```bash
go mod tidy          # 下载依赖
go run main.go       # 启动项目
```

### 访问地址
- 首页: http://localhost:8080/main
- 登录: http://localhost:8080/pages/user/login.html
- 注册: http://localhost:8080/pages/user/regist.html
- 后台管理: http://localhost:8080/pages/manager/manager.html
- 图书管理: http://localhost:8080/pages/manager/book_manager.html 或访问 http://localhost:8080/getPageBooks
- 订单管理: http://localhost:8080/getOrders

## 前端页面说明

### 用户相关页面
- **登录页面** (`login.html`): 用户登录入口，包含用户名和密码输入
- **注册页面** (`regist.html`): 用户注册入口，包含用户名、密码、确认密码、邮箱输入，实时验证用户名
- **登录成功页面** (`login_success.html`): 登录成功提示，显示欢迎信息
- **注册成功页面** (`regist_success.html`): 注册成功提示，提供跳转到首页链接

### 购物车相关页面
- **购物车页面** (`cart.html`): 显示购物车商品列表，支持修改数量、删除商品、清空购物车、结账
- **结算页面** (`checkout.html`): 订单结算完成页面，显示订单号

### 订单相关页面
- **我的订单** (`order.html`): 用户查看自己的所有订单，支持查看详情和确认收货
- **订单管理** (`order_manager.html`): 管理员查看所有订单，支持发货操作
- **订单详情** (`order_info.html`): 查看订单的详细商品信息（书名、作者、价格、数量、金额、封面）

### 管理相关页面
- **后台管理** (`manager.html`): 管理员后台入口，提供图书管理和订单管理入口
- **图书管理** (`book_manager.html`): 图书列表展示，支持分页、添加、修改、删除图书
- **图书编辑** (`book_edit.html`): 图书编辑/添加页面，统一处理新增和编辑操作

### 静态资源说明
- **CSS**: `style.css` - 统一页面样式设计
- **JavaScript**: `jquery-1.7.2.js` - 提供Ajax和DOM操作功能
- **图片资源**:
  - `logo.gif` - 网站Logo
  - `Go.jpg` - Go编程图书封面
  - `python.jpg` - Python图书封面
  - `mysql.jpg` - MySQL图书封面
  - `js.jpg` - JavaScript图书封面
  - `default.jpg` - 默认图书封面

## 测试说明

项目包含完整的单元测试，位于 `Test/` 目录：
- `test_utils.go`: 测试工具函数
- `userdao_test.go`: 用户数据访问测试
- `bookdao_test.go`: 图书数据访问测试
- `cartdao_test.go`: 购物车数据访问测试
- `cartitemdao_test.go`: 购物项数据访问测试
- `orderdao_test.go`: 订单数据访问测试
- `user_controller_test.go`: 用户控制器测试
- `book_controller_test.go`: 图书控制器测试
- `order_controller_test.go`: 订单控制器测试
- `cart_controller_test.go`: 购物车控制器测试

**注意**: dao目录下也存在 `userdao_test.go` 文件，这是早期版本的测试文件。

运行测试：
```bash
go test ./Test/...
```

## 项目总结

404书城项目是一个功能完整、结构清晰、技术规范的Web电商项目。项目采用Go语言的简洁高效特性，结合MySQL数据库，实现了完整的书店管理系统。项目架构清晰，代码规范，易于维护和扩展，是学习Go Web开发的优秀案例。

### 项目特点
- 完整的电商业务流程：用户注册登录、商品浏览、购物车管理、订单处理
- 规范的代码结构：MVC+DAO分层清晰，职责明确
- 良好的用户体验：Ajax异步交互、实时反馈、分页浏览
- 完善的测试覆盖：单元测试覆盖所有核心功能模块
- 安全可靠：Session管理、Cookie安全设置、用户身份验证

### 适用场景
- Go Web开发学习
- 小型电商系统开发
- MVC架构实践
- 数据库设计和操作练习
- 前后端交互实现

### 技术收获
通过本项目可以学习到：
1. Go语言Web开发基础知识
2. MVC架构设计和实现
3. MySQL数据库操作
4. Session和Cookie管理
5. Ajax异步交互
6. 模板引擎使用
7. 电商业务流程设计