package errcode

import "errors"

// 通用错误
var (
	ErrInvalidParam   = errors.New("参数错误")
	ErrNotFound       = errors.New("资源不存在")
	ErrAlreadyExists  = errors.New("资源已存在")
	ErrInternalServer = errors.New("服务器内部错误")
)

// 认证错误
var (
	ErrUnauthorized    = errors.New("未授权")
	ErrInvalidToken    = errors.New("无效的令牌")
	ErrTokenExpired    = errors.New("令牌已过期")
	ErrInvalidPassword = errors.New("用户名或密码错误")
	ErrAccountDisabled = errors.New("账户已被禁用")
	ErrForbidden       = errors.New("无权限访问")
)

// 业务错误
var (
	ErrActivityNotOpen     = errors.New("活动未开放报名")
	ErrActivityFull        = errors.New("活动报名已满")
	ErrAlreadyRegistered   = errors.New("已经报名过该活动")
	ErrPaymentFailed       = errors.New("支付失败")
	ErrColumnHasArticles   = errors.New("该栏目下存在文章，无法删除")
	ErrMenuHasChildren     = errors.New("存在子菜单，无法删除")
	ErrRoleInUse              = errors.New("该角色正在被使用，无法删除")
	ErrActivityHasRegistrations = errors.New("活动存在有效报名，无法删除")
	ErrRegistrationNotFound   = errors.New("报名记录不存在")
	ErrRegistrationCancelled  = errors.New("报名已取消")
	ErrPaymentNotFound        = errors.New("支付记录不存在")
	ErrPaymentAlreadyPaid     = errors.New("订单已支付")
	ErrRefundFailed           = errors.New("退款失败")
)
