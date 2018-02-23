module.exports = function (sharp, path, config) {
  var ImageProcessor = {
    getProcessedImage: function (width, height, gravity, gray, blur, filePath, shortName, callback) {
      gravity = ImageProcessor.getGravity(gravity)
      ImageProcessor.imageResize(width, height, gravity, filePath, gray, blur, function (error, image) {
        if (error) {
          return callback(error)
        }

        callback(null, image)
      })
    },

    getGravity: function(gravity) {
      gravity = gravity ? gravity : 'center'
      gravity = gravity == 'centre' ? 'center' : gravity
      return gravity
    },

    imageResize: function (width, height, gravity, filePath, gray, blur, callback) {
      try {
        var image = sharp(filePath).rotate().resize(width, height).crop(sharp.gravity[gravity])

        if (gray) {
          image.grayscale()
        }

        if (blur) {
          image.blur(10)
        }

        image.toFormat('jpeg', { progressive: true }).toBuffer(function (error, data) {
          callback(error, data)
        })
      } catch (error) {
        callback(error, null)
      }
    },

    getWidthAndHeight: function (params, square, callback) {
      var width = square ? params.size : params.width
      var height = square ? params.size : params.height
      callback(width, height)
    }
  }

  return ImageProcessor
}
