const api = require('./utils/api.js');
App({
  onLaunch() {
    wx.login({
      success: (res) => {
        api.login(res.code)
          .then((user) => {
            this.globalData.user = user;
            // 如果已经绑定手机号，保留在当前页面，否则跳转到注册页
            if (!user.phone) {
              // 获取当前页面路径
              const pages = getCurrentPages();
              const currentPage = pages[pages.length - 1];
              
              // 如果当前不在注册页，才跳转
              if (currentPage && currentPage.route !== 'pages/register/register') {
                wx.redirectTo({ url: '/pages/register/register' });
              }
            }
          })
          .catch((err) => {
            wx.showToast({ title: err.toString(), icon: 'none' });
          });
      }
    });
  },
  globalData: {
    user: null
  }
}); 