const api = require('./utils/api.js');
App({
  onLaunch() {
    wx.login({
      success: (res) => {
        api.login(res.code)
          .then((user) => {
            this.globalData.user = user;
            if (!user.phone) {
              wx.redirectTo({ url: '/pages/register/register' });
            } else {
              wx.redirectTo({ url: '/pages/home/home' });
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