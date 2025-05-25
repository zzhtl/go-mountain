const api = require('../../utils/api.js');
const app = getApp();

Page({
  data: {
    columns: [],
    loading: true
  },

  onLoad() {
    this.loadColumns();
  },

  // 加载栏目列表
  loadColumns() {
    api.getColumns()
      .then((columns) => {
        this.setData({ 
          columns: columns,
          loading: false 
        });
      })
      .catch((err) => {
        this.setData({ loading: false });
        wx.showToast({ 
          title: '加载栏目失败', 
          icon: 'none' 
        });
      });
  },

  // 点击栏目，跳转到文章列表
  onColumnTap(e) {
    const columnId = e.currentTarget.dataset.id;
    const columnName = e.currentTarget.dataset.name;
    wx.navigateTo({
      url: `/pages/articles/list?columnId=${columnId}&columnName=${encodeURIComponent(columnName)}`
    });
  }
}); 