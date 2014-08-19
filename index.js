var cluster = require('cluster');
if (cluster.isMaster) {

  var config = require('./config');
  var fs = require('fs');
  var filequeue = require('filequeue');
  var path = require('path'); 
  var imagesize = require('imagesize');
  
  var fq = new filequeue(200);
  var highestImageId = 0;

  try {
    var stats = require(config.stats_path);
  } catch (e) {
    var stats = { count: 0 };
  }

  try {
    var images = require(config.image_store_path);
  } catch (e) {
    var images = [];
  }

  if (images.length != 0) {
    highestImageId = images.length;
  }

  var cleanupAndExit = function () {
    cleanup();
    process.exit();
  }

  var cleanup = function () {
    fs.writeFileSync(config.stats_path, JSON.stringify(stats), 'utf8');
  }

  process.on('exit', cleanup);
  process.on('SIGINT', cleanupAndExit);
  process.on('uncaughtException', cleanupAndExit);

  var handleWorkerMessage = function (msg) {
    stats.count++;
  }

  var endsWith = function (str, suffix) {
    return str.indexOf(suffix, str.length - suffix.length) !== -1;
  }

  var startWebServers = function () {
    var cpuCount = require('os').cpus().length - 1;
    if (cpuCount < 2) {
      cpuCount = 2;
    }

    for (var i = 0, il=cpuCount; i < il; i++) {
      var worker = cluster.fork();
      worker.on('message', handleWorkerMessage);
      console.log('Worker ' + worker.id + ' started');
    }

    cluster.on('exit', function (worker) {
      console.log('Worker ' + worker.id + ' died');
      var worker = cluster.fork();
      worker.on('message', handleWorkerMessage);
      console.log('Worker ' + worker.id + ' started');
    });
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

  var loadImages = function () {
    scanDirectory(config.folder_path, function (err, results) {
      if (err) throw err;
      var filteredResults = results.filter(function (filename) {
        return images.filter(function (image) { return image.filename == filename; }) == 0;
      });
      var left = filteredResults.length;
      if (left <= 0) {
        return startWebServers();
      }
      filteredResults.forEach(function (filename) {
        var rs = fq.createReadStream(filename);  
        imagesize(rs, function (err, result) {  
          if (err) {
            return console.log('imageScan error: ' + err);
          }

          result.filename = filename;
          result.id = highestImageId++;
          images.push(result);

          if (!--left) {
            fs.writeFile(config.image_store_path, JSON.stringify(images), 'utf8', function (err) {});
            startWebServers();
          }
        });
      });
    });
  }

  loadImages();
} else {
  require('./server')(function (callback) {
    callback.listen(process.env.PORT || 5000, 'localhost');
  });
}