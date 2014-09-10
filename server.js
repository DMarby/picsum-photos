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

  var checkParameters = function (params, queryparams, square, callback) {
    if ((square && !params.size) || (square && isNaN(parseInt(params.size))) || (!square && !params.width) || (!square && !params.height) || (!square && isNaN(parseInt(params.width))) || (!square && isNaN(parseInt(params.height))) || (queryparams.gravity && sharp.gravity[queryparams.gravity] == null)) {
      return callback(true, 400, 'Invalid arguments');
    }
    if (params.size > config.max_width || params.size > config.max_height || params.width > config.max_height || params.height > config.max_width) {
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

  var getGravity = function(gravity) {
    gravity = gravity ? gravity : 'center';
    gravity = gravity == 'centre' ? 'center' : gravity;
    return gravity;
  }

  var getDestination = function (width, height, gravity, blur, filePath, prefix) {
    return config.cache_folder_path + '/' + prefix + path.basename(filePath, path.extname(filePath)) + '-' + width + 'x' + height + '-' + gravity + (blur ? '-blur' : '') + '.jpg';
  }

  var getShortDestination = function (width, height, gravity, blur, filePath, prefix) {
    return config.cache_folder_path + '/' + prefix + width + '^' + height + '-' + gravity + (blur ? '-blurred' : '') + '.jpg';
  }

  var getAndCheckDestination = function (width, height, gravity, blur, filePath, prefix, shortName, callback) {
    var destination = shortName ? getShortDestination(width, height, gravity, blur, filePath, prefix) : getDestination(width, height, gravity, blur, filePath, prefix);
    fs.exists(destination, function (exists) {
      callback(exists, destination);
    })
  }

  var imageResize = function (width, height, gravity, filePath, destination, callback) {
    try {
      sharp(filePath).rotate().resize(width, height).crop(sharp.gravity[gravity]).progressive().toFile(destination, function (err) {
        callback(err, destination);
      });
    } catch (e) {
      callback(e, null);
    }
  }

  var getProcessedImage = function (width, height, gravity, gray, blur, filePath, shortName, callback) {
    gravity = getGravity(gravity);
    getAndCheckDestination(width, height, gravity, blur, filePath, gray ? 'gray-' : '', shortName, function (exists, destination) {
      if (exists) {
        return callback(null, destination);
      }
      imageResize(width, height, gravity, filePath, destination, function (err, destination) {
        if (err) {
          return callback(err);
        }
        if (gray) {
          var modifyImage = gm(destination).colorspace('GRAY');
          if (blur) {
            modifyImage.blur(0, 5);
          }
          modifyImage.write(destination, function (err) {
            if (err) {  
              return callback(err);
            }
            callback(null, destination);
          })
        } else {
          if (blur) {
            gm(destination).blur(0, 5).write(destination, function (err) {
              if (err) {
                return callback(err);
              }
              callback(null, destination);
            })
          } else {
            callback(null, destination);
          }
        }
      })
    })
  }

  var getWidthAndHeight = function (params, square, callback) {
    var width = square ? params.size : params.width;
    var height = square ? params.size : params.height;
    callback(width, height);
  }


  var serveImage = function(req, res, square, gray) {
    checkParameters(req.params, req.query, square, function (err, code, message) {
      getWidthAndHeight(req.params, square, function (width, height) {
        if (err) {
          return displayError(res, code, message);
        }

        var filePath;
        var blur = false;
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
        getProcessedImage(width, height, req.query.gravity, gray, !(!req.query.blur && req.query.blur != ''), filePath, (!req.query.image && !req.query.random && req.query.random != ''), function (err, imagePath) {
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
  }

  app.use(express.static(path.join(__dirname, 'public')));

  app.all('*', function (req, res, next) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    next();
  });

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
        id: item.id,
        author: item.author,
        author_url: item.author_url,
        post_url: item.post_url
      }
      newImages.push(image);
    }
    res.jsonp(newImages);
  })

  app.get('/:size', function (req, res, next) {
    serveImage(req, res, true, false);
  })

  app.get('/g/:size', function (req, res, next) {
    serveImage(req, res, true, true);
  })

  app.get('/:width/:height', function (req, res, next) {
    serveImage(req, res, false, false);
  })

  app.get('/g/:width/:height', function (req, res, next) {
    serveImage(req, res, false, true);
  })

  app.get('*', function (req, res, next) {
    res.status(404);
    res.send({ error: 'Resource not found' });
  })

  callback(app);
}
