const api = require('../../utils/api.js');
const app = getApp();
Page({
  data: {
    phone: '',
    name: ''
  },
  onInputPhone(e) {
    this.setData({ phone: e.detail.value });
  },
  onInputName(e) {
    this.setData({ name: e.detail.value });
  },
  onSubmit() {
    const { phone, name } = this.data;
    const openid = app.globalData.user.openid;
    api.register(phone, openid, name)
      .then((user) => {
        // 更新全局用户信息
        app.globalData.user = { ...app.globalData.user, ...user };
        wx.redirectTo({ url: '/pages/home/home' });
      })
      .catch((err) => {
        wx.showToast({ title: err.toString(), icon: 'none' });
      });
  }
}); 