module.exports = function (callback) {
  var fs = require('fs')
  var path = require('path')
  var express = require('express')
  var cors = require('cors')
  var compress = require('compression')
  var sharp = require('sharp')
  var config = require('./config')()
  var packageinfo = require('./package.json')
  var imageProcessor = require('./imageProcessor')(sharp, path, config, fs)
  var util = require('util')

  var app = express()

  sharp.cache(0)

  try {
    var images = require(config.image_store_path)
  } catch (e) {
    var images = []
  }

  process.addListener('uncaughtException', function (err) {
    console.log('Uncaught exception: ')
    console.trace(err)
  })

  app.use(compress());

  app.use(cors());

  app.use(express.static(path.join(__dirname, 'public'), {
    extensions: ['html'],
    maxAge: '1h'
  }));

  app.get('/info', function (req, res, next) {
    res.jsonp({ name: packageinfo.name, version: packageinfo.version, author: packageinfo.author })
  })

  app.get('/list', function (req, res, next) {
    var newImages = []

    for (var i in images) {
      var item = images[i]
      var image = {
        format: item.format,
        width: item.width,
        height: item.height,
        filename: path.basename(item.filename),
        id: item.id,
        author: item.author,
        author_url: item.author_url,
        post_url: item.post_url
      }

      newImages.push(image)
    }

    res.jsonp(newImages)
  })

  app.get('/:size', function (req, res, next) {
    serveImage(req, res, true, false)
  })

  app.get('/g/:size', function (req, res, next) {
    serveImage(req, res, true, true)
  })

  app.get('/:width/:height', function (req, res, next) {
    serveImage(req, res, false, false)
  })

  app.get('/g/:width/:height', function (req, res, next) {
    serveImage(req, res, false, true)
  })

  app.all('*', function (req, res, next) {
    res.status(404)
    res.send({ error: 'Resource not found' })
  })

  var serveImage = function(req, res, square, gray) {
    checkParameters(req.params, req.query, square, function (error, code, message) {
      if (error) {
        return displayError(res, code, message)
      }

      imageProcessor.getWidthAndHeight(req.params, square, function (width, height) {
        var filePath
        var blur = !(!req.query.blur && req.query.blur !== '')
        var isRandom = (req.query.random || req.query.random === '')

        if (isRandom) {
          res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate')

          var params = 'image=' + images[Math.floor(Math.random() * images.length)].id

          if (req.query.gravity) {
            params += '&gravity=' + imageProcessor.getGravity(req.query.gravity)
          }

          if (blur) {
            params += '&blur=true'
          }

          if (gray) {
            res.redirect(util.format('/g/%s/%s/?%s', width, height, params))
          } else {
            res.redirect(util.format('/%s/%s/?%s', width, height, params))
          }

          return
        }

        if (req.query.image) {
          var matchingImage = findMatchingImage(req.query.image)

          if (matchingImage) {
            filePath = matchingImage.filename

            if (parseInt(width) == 0) {
              width = matchingImage.width
            }

            if (parseInt(height) == 0) {
              height = matchingImage.height
            }
          } else {
            return displayError(res, 404, 'Invalid image id')
          }
        } else {
          filePath = images[Math.floor(Math.random() * images.length)].filename
        }

        imageProcessor.getProcessedImage(filePath, parseInt(width), parseInt(height), req.query.gravity, gray, blur, function (error, image) {
          if (error) {
            console.log('filePath: ' + filePath)
            console.log('error: ' + error)
            return displayError(res, 500, 'Something went wrong')
          }

          res.setHeader('Cache-Control', 'public, max-age=604800') // Cache for 1 week
          res.set('Content-Type', 'image/jpeg')
          res.send(image)
        })
      })
    })
  }

  var checkParameters = function (params, queryparams, square, callback) {
    if ((square && !params.size) || (square && isNaN(parseInt(params.size))) || (!square && !params.width) || (!square && !params.height) || (!square && isNaN(parseInt(params.width))) || (!square && isNaN(parseInt(params.height))) || (queryparams.gravity && sharp.gravity[queryparams.gravity] != 0 && !sharp.gravity[queryparams.gravity])) {
      return callback(true, 400, 'Invalid arguments')
    }

    if (params.size > config.max_width || params.size > config.max_height || params.height > config.max_height || params.width > config.max_width) {
      if (queryparams.image) {
        var matchingImage = findMatchingImage(queryparams.image)

        if (matchingImage && params.height == matchingImage.height && params.width == matchingImage.width) {
          return callback(false)
        }
      }

      return callback(true, 413, 'Specified dimensions too large')
    }

    callback(false)
  }

  var findMatchingImage = function (id) {
    var matchingImages = images.filter(function (image) {
      return image.id == id
    })

    if (!matchingImages.length) {
      return false
    }

    return matchingImages[0]
  }

  var displayError = function (res, code, message) {
    res.status(code)
    res.send({ error: message })
  }

  callback(app)
}
