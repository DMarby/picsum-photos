module.exports = function (callback) {
  var gm = require('gm');
  var fs = require('fs');
  var path = require('path'); 
  var sharp = require('sharp');
  var express = require('express')
  var config = require('./config')();
  var packageinfo = require('./package.json');

  var app = express();

  try {
    var images = require(config.image_store_path);
  } catch (e) {
    var images = [];
  }

  fs.mkdir(config.cache_folder_path, function(e) {});

  process.addListener('uncaughtException', function (err) {
    console.log('Uncaught exception: ');
    console.trace(err);
  })

  var countImage = function () {
    process.send("count");
  }

  var checkParameters = function (params, callback) {
    if (!params.width || !params.height || isNaN(parseInt(params.width)) || isNaN(parseInt(params.height))) {
      return callback(true, 400, 'Invalid arguments');
    }
    if (params.width > config.max_height || params.height > config.max_width) {
      return callback(true, 413, 'Specified dimensions too large');
    }
    callback(false);
  }

  var displayError = function (res, code, message) {
    res.status(code);
    res.send({ error: message });
  }

  var findMatchingImage = function (id, callback) {
    var matchingImages = images.filter(function (image) { return image.id == id; });
    if (matchingImages.length == 0) {
      return false;
    }
    return matchingImages[0].filename;
  }

  var getDestination = function (width, height, filePath, prefix) {
    return config.cache_folder_path + '/' + prefix + path.basename(filePath, path.extname(filePath)) + '-' + width + 'x' + height + '.jpg';
  }

  var getShortDestination = function (width, height, filePath, prefix) {
    return config.cache_folder_path + '/' + prefix + width + '^' + height + '.jpg';
  }

  var getAndCheckDestination = function (width, height, filePath, prefix, shortName, callback) {
    var destination = shortName ? getShortDestination(width, height, filePath, prefix) : getDestination(width, height, filePath, prefix);
    fs.exists(destination, function (exists) {
      callback(exists, destination);
    })
  }

  var imageResize = function (width, height, filePath, destination, callback) {
    try {
      sharp(filePath).rotate().resize(width, height).crop().progressive().toFile(destination, function (err) {
        callback(err, destination);
      });
    } catch (e) {
      callback(e, null);
    }
  }

  var getProcessedImage = function (width, height, filePath, shortName, callback) {
    getAndCheckDestination(width, height, filePath, '', shortName, function (exists, destination) {
      if (exists) {
        return callback(null, destination);
      }
      imageResize(width, height, filePath, destination, function (err, destination) {
        if (err) {
          return callback(err);
        }
        callback(null, destination);
      })
    })
  }

  var getProcessedGrayImage = function (width, height, filePath, shortName, callback) {
    getAndCheckDestination(width, height, filePath, 'gray-', shortName, function (exists, destination) {
      if (exists) {
        return callback(null, destination);
      }
      imageResize(width, height, filePath, destination, function (err, destination) {
        if (err) {
          return callback(err);
        }
        gm(destination).colorspace('GRAY').write(destination, function (err) {
          if (err) {  
            return callback(err);
          }
          callback(null, destination);
        })
      })
    })
  }

  app.use(express.static(path.join(__dirname, 'public')));

  app.get('/info', function (req, res, next) {
    res.jsonp({ name: packageinfo.name, version: packageinfo.version, author: packageinfo.author });
  })

  app.get('/list', function (req, res, next) {
    var newImages = [];
    for (var i in images) {
      var item = images[i];
      var image = {
        format: item.format,
        width: item.width,
        height: item.height,
        filename: path.basename(item.filename),
        id: item.id
      }
      newImages.push(image);
    }
    res.jsonp(newImages);
  })

  app.get('/:width/:height', function (req, res, next) {
    checkParameters(req.params, function (err, code, message) {
      if (err) {
        return displayError(res, code, message);
      }

      var filePath;
      if (req.query.image) {
        var matchingImage = findMatchingImage(req.query.image);
        if (matchingImage) {
          filePath = matchingImage;
        } else {
          return displayError(res, 400, 'Invalid image id');
        }
      } else {
        filePath = images[Math.floor(Math.random() * images.length)].filename;
      }

      getProcessedImage(req.params.width, req.params.height, filePath, (!req.query.image && !req.query.random && req.query.random != ''), function (err, imagePath) {
        if (err) {
          console.log('filePath: ' + filePath);
          console.log('imagePath: ' + imagePath);
          console.log('error: ' + err);
          return displayError(res, 500, 'Something went wrong');
        }
        res.sendFile(imagePath);
        countImage();
      })
    })
  })

  app.get('/g/:width/:height', function (req, res, next) {
    checkParameters(req.params, function (err, code, message) {
      if (err) {
        return displayError(res, code, message);
      }

      var filePath;
      if (req.query.image) {
        var matchingImage = findMatchingImage(req.query.image);
        if (matchingImage) {
          filePath = matchingImage;
        } else {
          return displayError(res, 400, 'Invalid image id');
        }
      } else {
        filePath = images[Math.floor(Math.random() * images.length)].filename;
      }

      getProcessedGrayImage(req.params.width, req.params.height, filePath, (!req.query.image && !req.query.random && req.query.random != ''), function (err, imagePath) {
        if (err) {
          console.log('filePath: ' + filePath);
          console.log('imagePath: ' + imagePath);
          console.log('error: ' + err);
          return displayError(res, 500, 'Something went wrong');
        }
        res.sendFile(imagePath);
        countImage();
      })
    })
  })

  app.get('*', function (req, res, next) {
    res.status(404);
    res.send({ error: 'Resource not found' });
  })

  callback(app);
}
