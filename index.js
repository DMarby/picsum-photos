var gm = require('gm');
var fs = require('fs');
var path = require('path'); 
var sharp = require('sharp');
var filequeue = require('filequeue');
var imagesize = require('imagesize');
var express = require('express')
var config = require('./config');
var pjson = require('./package.json');

var fq = new filequeue(200);
var app = express();
var highestImageId = 0;
try {
  var images = require('./images.json');
} catch (e) {
  var images = [];
}
if (images.length != 0) {
  highestImageId = images.length;
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
    return callback(true);
  }
  callback(null, matchingImages[0].filename);
}

var getDestination = function (width, height, filePath, prefix) {
  return 'cache/' + prefix + path.basename(filePath, path.extname(filePath)) + '-' + width + 'x' + height + path.extname(filePath);
}

var getShortDestination = function (width, height, filePath, prefix) {
  return 'cache/' + prefix + width + '^' + height + path.extname(filePath);
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

var endsWith = function (str, suffix) {
  return str.indexOf(suffix, str.length - suffix.length) !== -1;
}

var scanDirectory = function (dir, done) {
  var results = [];
  fs.readdir(dir, function (err, list) {
    if (err) {
      return done(err);
    }
    var pending = list.length;
    if (!pending) {
      return done (null, results);
    }
    list.forEach(function (file) {
      file = path.resolve(dir, file);
      fs.stat(file, function (err, stat) {
        if (stat && stat.isFile() && !endsWith(file, '.DS_Store')) {
          results.push(file);
        }
        if (!--pending) done(null, results);
      });
    });
  });
};

var imageScan = function () {
  scanDirectory(config.folder_path, function (err, results) {
    if (err) throw err;
    var filteredResults = results.filter(function (filename) {
      return images.filter(function (image) { return image.filename == filename; }) == 0;
    });
    var left = filteredResults.length;
    filteredResults.forEach(function (filename) {
      var rs = fq.createReadStream(filename);  
      imagesize(rs, function (err, result) {  
        if (err) {
          return console.log(err);
        }

        console.log(filename);

        result.filename = filename;
        result.id = highestImageId++;
        images.push(result);

        if (!--left) {
          fs.writeFile('images.json', JSON.stringify(images), 'utf8', function (err) {});
        }
      });
    });
  });
}

app.get('/', function (req, res, next) {
  res.sendFile('public/index.html', { root: '.'} );
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
      findMatchingImage(req.query.image, function (err, matchingImage) {
        if (err) {
          return displayError(res, 400, 'Invalid image id');
        }
        filePath = matchingImage;
      })
    } else {
      filePath = images[Math.floor(Math.random() * images.length)].filename;
    }

    getProcessedImage(req.params.width, req.params.height, filePath, (!req.query.image && !req.query.random && req.query.random != ''), function (err, imagePath) {
      if (err) {
        return displayError(res, 500, 'Something went wrong');
      }
      res.sendFile(imagePath, { root: '.' });
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
      findMatchingImage(req.query.image, function (err, matchingImage) {
        if (err) {
          return displayError(res, 400, 'Invalid image id');
        }
        filePath = matchingImage;
      })
    } else {
      filePath = images[Math.floor(Math.random() * images.length)].filename;
    }

    getProcessedGrayImage(req.params.width, req.params.height, filePath, (!req.query.image && !req.query.random && req.query.random != ''), function (err, imagePath) {
      if (err) {
        return displayError(res, 500, 'Something went wrong');
      }
      res.sendFile(imagePath, { root: '.' });
    })
  })
})

app.get('*', function (req, res, next) {
  res.status(404);
  res.send({ error: 'Resource not found' });
})

imageScan();
app.listen(process.env.PORT || 5000);