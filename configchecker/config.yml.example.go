package configchecker

var simpleConfigYmlExample = []byte(`general:
  development: true
  logging: true
  pprof: true
email:
  host: #EMAIL_HOST
  port: #EMAIL_PORT
  username: #EMAIL_USERNAME
  password: #EMAIL_PASSWORD
  noReplyAddress: #EMAIL_NO_REPLY_ADDRESS
urls:
  admin: #URLS_ADMIN
  pprof: #PPROF_ADMIN
paths:
  assets:
    publicFiles: ` + defaultAssetsFilesPublic + `
    privateFiles: ` + defaultAssetsFilesPrivate + `
`)

// nolint:gochecknoglobals
var configYmlExample = []byte(`general:
  development: true
  logging: true
  pprof: true
  languages: [en]
  defaultLanguage: en
server:
  host: localhost
  port: 8443
  httpRedirectPort: 8080
database:
  host: #DATABASE_HOST
  port: #DATABASE_PORT
  name: #DATABASE_NAME
  users:
    selecter: selecter
    creator: creator
    inserter: inserter
    updater: updater
    deletor: deletor
    migrator: migrator
email:
  host: #EMAIL_HOST
  port: #EMAIL_PORT
  username: #EMAIL_USERNAME
  password: #EMAIL_PASSWORD
  noReplyAddress: #EMAIL_NO_REPLY_ADDRESS
security:
  globalAuthentication: false
  bcryptRounds: 12
  formTokenLifespan: 8m
  formTokenCleanupInterval: 10s
session:
  cookieName: s
  expiration: 45m
  rememberMeExpiration: 720h
urls:
  admin: #URLS_ADMIN
  pprof: #PPROF_ADMIN
assets:
  gzip: true
  brotli: true
  gzipFiles: true
  brotliFiles: true
  cacheMaxAge: 60
paths:
  server:
    sslCertificateFile: ./app/server/localhost.crt
    sslKeyFile: ./app/server/localhost.key
  database:
    sslRootCertificateFile: ./app/database/cert/ca.crt
    selecter:
      sslCertificateFile: ./app/database/certs/client.selecter.crt
      sslKeyFile: ./app/database/certs/client.selecter.key
    creator:
      sslCertificateFile: ./app/database/certs/client.creator.crt
      sslKeyFile: ./app/database/certs/client.creator.key
    inserter:
      sslCertificateFile: ./app/database/certs/client.inserter.crt
      sslKeyFile: ./app/database/certs/client.inserter.key
    updater:
      sslCertificateFile: ./app/database/certs/client.updater.crt
      sslKeyFile: ./app/database/certs/client.updater.key
    deletor:
      sslCertificateFile: ./app/database/certs/client.deletor.crt
      sslKeyFile: ./app/database/certs/client.deletor.key
    migrator:
      sslCertificateFile: ./app/database/certs/client.migrator.crt
      sslKeyFile: ./app/database/certs/client.migrator.key
  assets:
    stylesheets: ./app/assets/css
    javascript: ./app/assets/js
    images: ./app/assets/images
    publicRootFiles: ./app/assets/files/root
    publicFiles: ` + defaultAssetsFilesPublic + `
    privateFiles: ` + defaultAssetsFilesPrivate + `
  translations: ./app/translations
`)
