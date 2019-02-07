// vue.config.js
module.exports = {
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
  }
}
