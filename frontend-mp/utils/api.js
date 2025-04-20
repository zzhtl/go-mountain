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

module.exports = { login, register }; 