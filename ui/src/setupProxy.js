const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = function (app) {
  app.use(
    "/api/mlp",
    createProxyMiddleware({
      target: process.env.REACT_APP_MLP_API,
      pathRewrite: { "^/api/mlp": "" },
      changeOrigin: true,
    })
  );
  app.use(
    "/api/xp/v1",
    createProxyMiddleware({
      target: process.env.REACT_APP_XP_API,
      pathRewrite: { "^/api/xp/v1": "" },
      changeOrigin: true,
    })
  );
};
