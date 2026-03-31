package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/errcode"
	"github.com/zzhtl/go-mountain/internal/repository"
)

// MenuService 菜单管理服务
type MenuService struct {
	repo *repository.BaseRepo[model.Menu]
	db   *gorm.DB
}

// NewMenuService 创建菜单服务
func NewMenuService(db *gorm.DB) *MenuService {
	return &MenuService{
		repo: repository.NewBaseRepo[model.Menu](db),
		db:   db,
	}
}

// List 获取所有菜单
func (s *MenuService) List(ctx context.Context) ([]model.Menu, error) {
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Order("sort, id")
	}
	return s.repo.FindAll(ctx, scope)
}

// Tree 获取菜单树形结构
func (s *MenuService) Tree(ctx context.Context) ([]*model.Menu, error) {
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Where("status = 1").Order("sort, id")
	}
	menus, err := s.repo.FindAll(ctx, scope)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus, 0), nil
}

// Get 获取单个菜单
func (s *MenuService) Get(ctx context.Context, id int64) (*model.Menu, error) {
	return s.repo.GetByID(ctx, id)
}

// Create 创建菜单
func (s *MenuService) Create(ctx context.Context, menu *model.Menu) error {
	if menu.Type == 0 {
		menu.Type = 1
	}
	menu.Status = 1
	return s.repo.Create(ctx, menu)
}

// Update 更新菜单
func (s *MenuService) Update(ctx context.Context, id int64, updates map[string]any) error {
	return s.repo.Update(ctx, id, updates)
}

// UpdateStatus 更新菜单状态
func (s *MenuService) UpdateStatus(ctx context.Context, id int64, status int) error {
	return s.repo.Update(ctx, id, map[string]any{"status": status})
}

// Delete 删除菜单
func (s *MenuService) Delete(ctx context.Context, id int64) error {
	// 检查是否有子菜单
	exists, _ := s.repo.Exists(ctx, "parent_id = ?", id)
	if exists {
		return errcode.ErrMenuHasChildren
	}
	// 删除角色菜单关联
	s.db.WithContext(ctx).Where("menu_id = ?", id).Delete(&model.RoleMenu{})
	return s.repo.Delete(ctx, id)
}

// InitDefaultMenus 初始化默认菜单
func (s *MenuService) InitDefaultMenus(ctx context.Context) error {
	var count int64
	s.db.WithContext(ctx).Model(&model.Menu{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 内容管理目录
	contentDir := model.Menu{ParentID: 0, Name: "content", Title: "内容管理", Path: "", Component: "", Icon: "Document", Sort: 1, Type: 1, Status: 1}
	s.db.WithContext(ctx).Create(&contentDir)

	contentMenus := []model.Menu{
		{ParentID: contentDir.ID, Name: "articles", Title: "文章管理", Path: "/admin/articles", Component: "content/ArticleList", Icon: "Document", Sort: 1, Type: 2, Status: 1},
		{ParentID: contentDir.ID, Name: "columns", Title: "栏目管理", Path: "/admin/columns", Component: "content/ColumnList", Icon: "Menu", Sort: 2, Type: 2, Status: 1},
	}
	for i := range contentMenus {
		s.db.WithContext(ctx).Create(&contentMenus[i])
	}

	// 业务管理目录
	bizDir := model.Menu{ParentID: 0, Name: "business", Title: "业务管理", Path: "", Component: "", Icon: "ShoppingCart", Sort: 2, Type: 1, Status: 1}
	s.db.WithContext(ctx).Create(&bizDir)

	bizMenus := []model.Menu{
		{ParentID: bizDir.ID, Name: "activities", Title: "活动管理", Path: "/admin/activities", Component: "business/ActivityList", Icon: "Calendar", Sort: 1, Type: 2, Status: 1},
		{ParentID: bizDir.ID, Name: "registrations", Title: "报名管理", Path: "/admin/registrations", Component: "business/RegistrationList", Icon: "List", Sort: 2, Type: 2, Status: 1},
		{ParentID: bizDir.ID, Name: "payments", Title: "支付管理", Path: "/admin/payments", Component: "business/PaymentList", Icon: "Money", Sort: 3, Type: 2, Status: 1},
	}
	for i := range bizMenus {
		s.db.WithContext(ctx).Create(&bizMenus[i])
	}

	// 系统管理目录
	sysDir := model.Menu{ParentID: 0, Name: "system", Title: "系统管理", Path: "", Component: "", Icon: "Setting", Sort: 3, Type: 1, Status: 1}
	s.db.WithContext(ctx).Create(&sysDir)

	sysMenus := []model.Menu{
		{ParentID: sysDir.ID, Name: "mp-users", Title: "小程序用户", Path: "/admin/users", Component: "system/UserList", Icon: "User", Sort: 1, Type: 2, Status: 1},
		{ParentID: sysDir.ID, Name: "backend-users", Title: "用户管理", Path: "/admin/backend-users", Component: "system/BackendUserList", Icon: "UserFilled", Sort: 2, Type: 2, Status: 1},
		{ParentID: sysDir.ID, Name: "roles", Title: "角色管理", Path: "/admin/roles", Component: "system/RoleList", Icon: "Key", Sort: 3, Type: 2, Status: 1},
		{ParentID: sysDir.ID, Name: "menus", Title: "菜单管理", Path: "/admin/menus", Component: "system/MenuList", Icon: "Grid", Sort: 4, Type: 2, Status: 1},
		{ParentID: sysDir.ID, Name: "operation-logs", Title: "操作日志", Path: "/admin/operation-logs", Component: "system/OperationLogList", Icon: "Notebook", Sort: 5, Type: 2, Status: 1},
		{ParentID: sysDir.ID, Name: "system-configs", Title: "系统配置", Path: "/admin/system-configs", Component: "system/SystemConfigList", Icon: "Tools", Sort: 6, Type: 2, Status: 1},
	}
	for i := range sysMenus {
		s.db.WithContext(ctx).Create(&sysMenus[i])
	}

	// 代码生成器
	codegenDir := model.Menu{ParentID: 0, Name: "codegen", Title: "代码生成", Path: "", Component: "", Icon: "Cpu", Sort: 4, Type: 1, Status: 1}
	s.db.WithContext(ctx).Create(&codegenDir)

	codegenMenu := model.Menu{ParentID: codegenDir.ID, Name: "codegen-config", Title: "生成配置", Path: "/admin/codegen", Component: "codegen/CodegenList", Icon: "Cpu", Sort: 1, Type: 2, Status: 1}
	s.db.WithContext(ctx).Create(&codegenMenu)

	// 为 admin 角色分配所有菜单
	var allMenus []model.Menu
	s.db.WithContext(ctx).Find(&allMenus)

	var adminRole model.Role
	if err := s.db.WithContext(ctx).Where("name = ?", "admin").First(&adminRole).Error; err == nil {
		for _, m := range allMenus {
			s.db.WithContext(ctx).Create(&model.RoleMenu{RoleID: adminRole.ID, MenuID: m.ID})
		}
	}

	return nil
}
