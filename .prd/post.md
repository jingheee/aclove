# 匿名论坛发帖功能 PRD

## 1. 执行摘要

### 问题陈述

当前 aclove 匿名论坛仅有用户系统和分类管理，缺少核心的内容生产功能——发帖。用户无法创建和分享内容，导致平台缺乏活力。

### 解决方案

构建一套完整的匿名发帖系统，支持 Markdown 编辑器、媒体附件、分类选择，并集成现有的匿名用户会话体系，确保用户身份匿名但行为可追溯。

### 成功标准

- 发帖接口响应时间 < 200ms（P95）
- 支持单帖最多 9 张图片/视频附件
- 帖子标题 5-100 字符，内容 10-10000 字符
- 同一用户 30 秒内不能重复发帖（防刷机制）
- 前端编辑器首屏加载时间 < 1s

---

## 2. 用户体验与功能

### 用户画像

| 角色       | 描述                           | 需求                         |
| ---------- | ------------------------------ | ---------------------------- |
| 匿名发帖者 | 无需注册，通过 Cookie 识别身份 | 快速发帖、隐私保护、简单编辑 |
| 浏览者     | 阅读帖子内容的访客             | 清晰的帖子展示、快速加载     |
| 管理员     | 管理帖子内容                   | 审核、删除、查看发帖者信息   |

### 用户故事

#### US-001: 创建帖子

**作为** 匿名用户，**我想** 在指定分类下发帖，**以便** 分享我的想法和内容。

**验收标准:**

- 必须选择已有分类才能发帖
- 标题必填，5-100 字符
- 内容必填，10-10000 字符，支持 Markdown
- 可选上传最多 9 张媒体文件
- 发帖成功后跳转到帖子详情页
- 发帖失败时显示具体错误信息

#### US-002: 编辑帖子

**作为** 帖子作者，**我想** 编辑我发布的帖子，**以便** 修正错误或更新内容。

**验收标准:**

- 仅允许作者本人编辑帖子
- 记录编辑次数和最后编辑时间
- 显示"已编辑"标识
- 编辑后保留原有媒体附件

#### US-003: 删除帖子

**作为** 帖子作者或管理员，**我想** 删除帖子，**以便** 移除不当内容。

**验收标准:**

- 作者可删除自己的帖子
- 管理员可删除任何帖子并填写删除原因
- 删除为软删除，数据保留

#### US-004: 浏览帖子列表

**作为** 浏览者，**我想** 按分类查看帖子列表，**以便** 发现感兴趣的内容。

**验收标准:**

- 支持按分类筛选
- 支持分页加载，每页 20 条
- 支持按时间、热度排序
- 列表项显示标题、摘要、作者（匿名标识）、时间、回复数、浏览数

#### US-005: 查看帖子详情

**作为** 浏览者，**我想** 查看帖子完整内容和评论，**以便** 深入了解讨论。

**验收标准:**

- 显示完整 Markdown 渲染内容
- 显示媒体附件预览
- 显示帖子元信息（时间、浏览数、投票数）
- 加载时间 < 300ms

### 非目标

- 支持富文本编辑器
- 不支持帖子置顶/加精（V2 版本）
- 不支持帖子搜索（V2 版本）
- 不支持帖子分享链接自定义

---

## 3. 技术规范

### 3.1 架构概览

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Vue 3 +       │────▶│   Go + Gin      │────▶│   PostgreSQL    │
│   Naive UI      │     │   REST API      │     │   (Posts Table) │
│                 │◀────│                 │◀────│                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                       │
        ▼                       ▼
┌─────────────────┐     ┌─────────────────┐
│  Markdown Editor│     │   Redis Cache   │
│  (bytemd)       │     │   (Hot Posts)   │
└─────────────────┘     └─────────────────┘
```

### 3.2 数据模型

基于现有 `posts` 表结构：

```go
// PostDO 已存在于 backend/models/query/post_gen.go
type PostDO struct {
    ID               int64          // 雪花 ID
    UserID           int64          // 匿名用户 ID
    CategoryID       int64          // 分类 ID
    Title            string         // 标题
    Content          string         // Markdown 原文
    ContentRendered  *string        // 渲染后的 HTML
    MediaAttachments datatypes.JSON // 媒体附件 [{url, type, size}]
    ViewCount        *int           // 浏览数
    ReplyCount       *int           // 回复数
    UpvoteCount      *int           // 点赞数
    Score            *float64       // 热度分数
    IP               *string        // 发布者 IP
    UserAgent        *string        // 发布者 UA
    EditorType       *string        // 编辑器类型 (markdown)
    EditCount        *int           // 编辑次数
    LastEditedAt     *time.Time     // 最后编辑时间
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        gorm.DeletedAt // 软删除
}
```

### 3.3 API 接口设计

#### POST /api/posts

创建帖子

**请求体:**

```json
{
  "category_id": 123456789,
  "title": "帖子标题",
  "content": "## 正文内容\n支持 Markdown",
  "media_attachments": [
    {
      "url": "https://cdn.example.com/img1.jpg",
      "type": "image",
      "size": 1024000
    }
  ]
}
```

**响应:**

```json
{
  "id": 987654321,
  "title": "帖子标题",
  "content": "## 正文内容...",
  "category_id": 123456789,
  "created_at": "2026-01-31T12:00:00Z"
}
```

**错误码:**

- 400: 参数错误（标题/内容长度不符、分类不存在）
- 429: 发帖过于频繁
- 403: 用户被封禁或冷却中

#### GET /api/posts

获取帖子列表

**查询参数:**

- `category_id`: 分类 ID（可选）
- `page`: 页码，默认 1
- `page_size`: 每页数量，默认 20，最大 50
- `sort`: 排序方式，`newest` | `hot` | `top`，默认 `newest`

**响应:**

```json
{
  "items": [
    {
      "id": 987654321,
      "title": "帖子标题",
      "summary": "摘要内容...",
      "category_id": 123456789,
      "author": { "id": 111222333, "anonymous": true },
      "view_count": 100,
      "reply_count": 20,
      "upvote_count": 50,
      "created_at": "2026-01-31T12:00:00Z"
    }
  ],
  "total": 1000,
  "page": 1,
  "page_size": 20
}
```

#### GET /api/posts/:id

获取帖子详情

**响应:**

```json
{
  "id": 987654321,
  "title": "帖子标题",
  "content": "## Markdown 原文",
  "content_rendered": "<h2>正文内容</h2>",
  "media_attachments": [...],
  "category": {"id": 123, "name": "技术讨论"},
  "author": {"id": 111, "anonymous": true},
  "view_count": 100,
  "reply_count": 20,
  "upvote_count": 50,
  "edit_count": 0,
  "created_at": "2026-01-31T12:00:00Z"
}
```

#### PUT /api/posts/:id

编辑帖子

**请求体:**

```json
{
  "title": "新标题",
  "content": "新内容",
  "media_attachments": [...]
}
```

#### DELETE /api/posts/:id

删除帖子

**请求体（管理员）:**

```json
{
  "reason": "违规内容"
}
```

### 3.4 前端组件设计

#### 组件清单

| 组件名          | 路径                                  | 职责              |
| --------------- | ------------------------------------- | ----------------- |
| PostEditor      | `components/post/PostEditor.vue`      | 帖子创建/编辑表单 |
| PostList        | `components/post/PostList.vue`        | 帖子列表展示      |
| PostCard        | `components/post/PostCard.vue`        | 单条帖子卡片      |
| PostDetail      | `components/post/PostDetail.vue`      | 帖子详情页        |
| MarkdownPreview | `components/post/MarkdownPreview.vue` | Markdown 预览     |
| MediaUploader   | `components/post/MediaUploader.vue`   | 媒体上传组件      |

#### PostEditor 组件接口

```vue
<script setup>
const props = defineProps({
  categoryId: Number, // 默认选中分类
  editMode: Boolean, // 是否为编辑模式
  postId: Number, // 编辑模式下的帖子 ID
});

const emit = defineEmits([
  "submit", // 提交成功
  "cancel", // 取消编辑
  "error", // 错误回调
]);
</script>
```

### 3.5 状态管理

```javascript
// stores/post.js
export const usePostStore = defineStore('post', {
  state: () => ({
    currentPost: null,
    postList: [],
    loading: false,
    error: null,
  }),

  actions: {
    async createPost(data) { ... },
    async fetchPosts(params) { ... },
    async fetchPostDetail(id) { ... },
    async updatePost(id, data) { ... },
    async deletePost(id, reason) { ... },
  }
});
```

### 3.6 安全与隐私

1. **匿名身份保护**
   - API 返回的作者信息仅包含匿名 ID，不包含 IP、UA 等敏感信息
   - 仅管理员可查看发帖者真实身份

2. **内容安全**
   - Markdown 渲染使用 DOMPurify 过滤 XSS
   - 图片上传限制类型：jpg, png, gif, webp
   - 单文件大小限制：5MB

3. **防刷机制**
   - 同一用户 30 秒冷却期
   - 单日发帖上限：50 帖
   - 异常行为自动触发验证码

4. **审计日志**
   - 记录所有发帖、编辑、删除操作
   - 保留 IP、UA、时间戳

---

## 4. 实现计划

### Phase 1: 核心 API (Week 1)

- [ ] PostService 服务层实现
- [ ] PostHandler 处理器实现
- [ ] 路由注册与中间件集成
- [ ] 单元测试覆盖

### Phase 2: 前端编辑器 (Week 1-2)

- [ ] PostEditor 组件开发
- [ ] Markdown 编辑器集成 (bytemd)
- [ ] 媒体上传组件
- [ ] 表单验证

### Phase 3: 列表与详情 (Week 2)

- [ ] PostList 组件
- [ ] PostCard 组件
- [ ] PostDetail 组件
- [ ] 分页与排序

### Phase 4: 优化与测试 (Week 3)

- [ ] 性能优化（缓存、懒加载）
- [ ] 集成测试
- [ ] 安全审计
- [ ] 文档完善

---

## 5. 风险分析

| 风险         | 影响 | 缓解措施                         |
| ------------ | ---- | -------------------------------- |
| 用户恶意刷帖 | 高   | 冷却期、发帖上限、内容审核队列   |
| XSS 攻击     | 高   | DOMPurify、CSP、输入过滤         |
| 存储成本激增 | 中   | 图片压缩、CDN、定期清理          |
| 性能瓶颈     | 中   | Redis 缓存、分页优化、数据库索引 |

---

## 6. 附录

### 6.1 依赖清单

**后端:**

- `github.com/yuin/goldmark` - Markdown 渲染
- `github.com/microcosm-cc/bluemonday` - HTML 过滤

**前端:**

- `@bytemd/vue` - Markdown 编辑器
- `bytemd/plugin-gfm` - GitHub Flavored Markdown

### 6.2 数据库索引

```sql
-- 已存在，如需优化可添加：
CREATE INDEX idx_posts_category_created ON posts(category_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_user_created ON posts(user_id, created_at DESC);
CREATE INDEX idx_posts_score ON posts(score DESC) WHERE deleted_at IS NULL;
```
