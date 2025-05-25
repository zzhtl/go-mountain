const api = require('../../utils/api.js');

Page({
  data: {
    columnId: null,
    columnName: '',
    articles: [],
    page: 1,
    pageSize: 10,
    total: 0,
    loading: false,
    hasMore: true
  },

  onLoad(options) {
    const { columnId, columnName } = options;
    this.setData({
      columnId: columnId,
      columnName: decodeURIComponent(columnName || '')
    });
    
    wx.setNavigationBarTitle({
      title: this.data.columnName || '文章列表'
    });
    
    this.loadArticles();
  },

  // 加载文章列表
  loadArticles() {
    if (this.data.loading || !this.data.hasMore) return;
    
    this.setData({ loading: true });
    
    api.getArticlesByColumn(this.data.columnId, this.data.page, this.data.pageSize)
      .then((res) => {
        const newArticles = res.list || [];
        this.setData({
          articles: this.data.page === 1 ? newArticles : this.data.articles.concat(newArticles),
          total: res.total,
          loading: false,
          hasMore: this.data.articles.length + newArticles.length < res.total
        });
      })
      .catch((err) => {
        this.setData({ loading: false });
        wx.showToast({ 
          title: '加载失败', 
          icon: 'none' 
        });
      });
  },

  // 点击文章，跳转到详情页
  onArticleTap(e) {
    const articleId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/pages/articles/detail?id=${articleId}`
    });
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true });
    this.loadArticles();
    wx.stopPullDownRefresh();
  },

  // 上拉加载更多
  onReachBottom() {
    if (this.data.hasMore) {
      this.setData({ page: this.data.page + 1 });
      this.loadArticles();
    }
  }
}); 