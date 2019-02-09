// vue.config.js
module.exports = {
  publicPath: "/",
  outputDir: "./dist",

  // options...
  assetsDir: "static",

  pluginOptions: {
    i18n: {
      locale: 'en',
      fallbackLocale: 'en',
      localeDir: 'locales',
      enableInSFC: true
    }
  },

  pages: {
    'index': {
      // entry for the page
      entry: 'src/pages/index/main.js',
      // the source template
      template: 'public/index.html',
      // output as dist/index.html
      filename: 'index.html',
      // when using title option,
      // template title tag needs to be <title><%= htmlWebpackPlugin.options.title %></title>
      title: 'Go-Bradcaster'
      // chunks to include on this page, by default includes
      // extracted common chunks and vendor chunks.
      //chunks: ['chunk-vendors', 'chunk-common', 'index']
    },
    // when using the entry-only string format,
    // template is inferred to be `public/subpage.html`
    // and falls back to `public/index.html` if not found.
    // Output filename is inferred to be `subpage.html`.
    'setup': {
      entry: 'src/pages/setup/main.js',
      template: 'public/index.html',
      filename: 'setup.html',
      title: 'Go-Bradcaster Setup'
      //chunks: ['chunk-vendors', 'chunk-common', 'index']
    }
  },

  devServer: {
    historyApiFallback: {
      rewrites: [
        { from: /\/index/, to: '/index.html' },
        { from: /\/setup/, to: '/setup.html' }
      ]
    }
  }

}
