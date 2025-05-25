// API 调用封装
const API_BASE_URL = "http://localhost:8080/api/mp";

/**
 * 小程序登录，调用后端获取 openid
 */
function login(code) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE_URL}/login`,
      method: 'POST',
      data: { code },
      success: (res) => {
        if (res.statusCode === 200) resolve(res.data);
        else reject(res.data.error || '登录失败');
      },
      fail: (err) => reject(err)
    });
  });
}

/**
 * 绑定手机号，调用后端注册接口
 */
function register(phone, openid, name) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE_URL}/register`,
      method: 'POST',
      header: { 'content-type': 'application/json' },
      data: { phone, openid, name },
      success: (res) => {
        if (res.statusCode === 201 || res.statusCode === 200) resolve(res.data);
        else reject(res.data.error || '注册失败');
      },
      fail: (err) => reject(err)
    });
  });
}

/**
 * 获取栏目列表
 */
function getColumns() {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE_URL}/columns/`,
      method: 'GET',
      success: (res) => {
        if (res.statusCode === 200) resolve(res.data);
        else reject(res.data.error || '获取栏目失败');
      },
      fail: (err) => reject(err)
    });
  });
}

/**
 * 根据栏目获取文章列表
 */
function getArticlesByColumn(columnId, page = 1, pageSize = 10) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE_URL}/articles/column/${columnId}`,
      method: 'GET',
      data: { page, page_size: pageSize },
      success: (res) => {
        if (res.statusCode === 200) resolve(res.data);
        else reject(res.data.error || '获取文章列表失败');
      },
      fail: (err) => reject(err)
    });
  });
}

/**
 * 获取文章详情
 */
function getArticleDetail(id) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE_URL}/articles/${id}`,
      method: 'GET',
      success: (res) => {
        if (res.statusCode === 200) {
          // 处理日期格式
          if (res.data.created_at) {
            res.data.created_at = formatDate(res.data.created_at);
          }
          resolve(res.data);
        } else {
          reject(res.data.error || '获取文章详情失败');
        }
      },
      fail: (err) => reject(err)
    });
  });
}

/**
 * 格式化日期
 */
function formatDate(dateStr) {
  const date = new Date(dateStr);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

module.exports = { 
  login, 
  register,
  getColumns,
  getArticlesByColumn,
  getArticleDetail
}; 