var gm = require('gm')

module.exports = exports = function (sharp, path, config, fs) {
  var ImageProcessor = {
    'getProcessedImage': function (width, height, gravity, gray, blur, filePath, shortName, callback) {
      gravity = ImageProcessor.getGravity(gravity)
      ImageProcessor.getAndCheckDestination(width, height, gravity, blur, filePath, gray ? 'gray-' : '', shortName, function (exists, destination) {
        if (exists) {
          return callback(null, destination)
        }

        ImageProcessor.imageResize(width, height, gravity, filePath, destination, gray, function (error, destination) {
          if (error) {
            ImageProcessor.deleteFile(destination)
            return callback(error)
          }

          if (blur) {
            gm(destination).blur(0, 5).write(destination, function (error) {
              if (error) {
                ImageProcessor.deleteFile(destination)
                return callback(error)
              }
              callback(null, destination)
            })
          } else {
            callback(null, destination)
          }
        })
      })
    },

    'getGravity': function(gravity) {
      gravity = gravity ? gravity : 'center'
      gravity = gravity == 'centre' ? 'center' : gravity
      return gravity
    },

    'getAndCheckDestination': function (width, height, gravity, blur, filePath, prefix, shortName, callback) {
      var destination = shortName ? ImageProcessor.getShortDestination(width, height, gravity, blur, filePath, prefix) : ImageProcessor.getDestination(width, height, gravity, blur, filePath, prefix)
      fs.exists(destination, function (exists) {
        callback(exists, destination)
      })
    },

    'getDestination': function (width, height, gravity, blur, filePath, prefix) {
      return config.cache_folder_path + '/' + prefix + path.basename(filePath, path.extname(filePath)) + '-' + width + 'x' + height + '-' + gravity + (blur ? '-blur' : '') + '.jpeg'
    },

    'getShortDestination': function (width, height, gravity, blur, filePath, prefix) {
      return config.cache_folder_path + '/' + prefix + width + '^' + height + '-' + gravity + (blur ? '-blurred' : '') + '.jpeg'
    },

    'imageResize': function (width, height, gravity, filePath, destination, gray, callback) {
      try {
        var image = sharp(filePath).rotate().resize(width, height).crop(sharp.gravity[gravity]);
        
        if (gray) {
          image.grayscale()
        }

        image.jpeg().progressive().toFile(destination, function (error) {
          callback(error, destination)
        })
      } catch (error) {
        callback(error, null)
      }
    },

    'deleteFile': function (destination) {
      fs.unlink(destination, function (error) {
        console.log('Error, deleted file')
      })
    },

    'getWidthAndHeight': function (params, square, callback) {
      var width = square ? params.size : params.width
      var height = square ? params.size : params.height
      callback(width, height)
    }
  }

  return ImageProcessor
}