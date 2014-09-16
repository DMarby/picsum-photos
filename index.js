var cluster = require('cluster');
if (cluster.isMaster) {

  var config = require('./config')();
  var fs = require('fs');
  var path = require('path'); 
  var sharp = require('sharp');
  var io = require('socket.io')(config.stats_port);
  var vnstat = require('vnstat-dumpdb');
  var metadata = require(config.metadata_path);
  console.log('Config:');
  console.log(config);

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

  var publicStats = {};
  var bandWidth = 0;

  io.on('connection', function (socket) {
    socket.emit('stats', publicStats);
  });

  var fetchStats = function () {
    publicStats = {
      count: stats.count,
      bandWidth: bandWidth,
      images: images.length
    }
    io.emit('stats', publicStats);
  }

  var fetchBandwidth = function () {
    vnstat.dumpdb(function (err, data) {
      if (err) {
        console.log('Couldn\'t fetch bandwidth: ' + err);
      } else {
        bandWidth = data.traffic.total.tx;
      }
    })
  }

  setInterval(fetchBandwidth, 1000 * 30);
  setInterval(fetchStats, 1000);
  fetchBandwidth();
  fetchStats();

  var cleanupAndExit = function () {
    cleanup();
    process.exit();
  }

  var cleanup = function () {
    fs.writeFileSync(config.stats_path, JSON.stringify(stats), 'utf8');
  }

  process.on('exit', cleanup);
  process.on('SIGINT', cleanupAndExit);
  process.on('SIGTERM', cleanupAndExit);
  process.on('uncaughtException', function(err) {
    console.log('Uncaught exception: ');
    console.trace(err);
    cleanupAndExit();
  });

  var handleWorkerMessage = function (msg) {
    stats.count++;
  }

  var endsWith = function (str, suffix) {
    return str.indexOf(suffix, str.length - suffix.length) !== -1;
  }

  var findMetadata = function (filename) {
    for (var i in metadata) {
      if (metadata[i].filename == filename) {
        return metadata[i];
      }
    }
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
          if (stat && stat.isFile() && !endsWith(file, '.DS_Store') && !endsWith(file, 'metadata.json')) {
            results.push(file);
          }
          if (!--pending) done(null, results);
        });
      });
    });
  };

  var loadImages = function () {
    var newimages = [];
    for (var i in images) {
      var item = images[i];
      var the_metadata = findMetadata(path.basename(item.filename));
      item.post_url = the_metadata.post_url;
      item.author = the_metadata.author;
      item.author_url = the_metadata.author_url;
      newimages.push(item);
    }
    scanDirectory(config.folder_path, function (err, results) {
      if (err) throw err;
      var filteredResults = results.filter(function (filename) {
        return images.filter(function (image) { return image.filename == filename; }) == 0;
      });
      var left = filteredResults.length;
      if (left <= 0) {
        console.log('Done scanning, no new images');
        newimages.sort(function (a,b) { 
          return a.id > b.id; 
        });
        images = newimages;
        fs.writeFile(config.image_store_path, JSON.stringify(newimages), 'utf8', function (err) {
          startWebServers();
        });
        return;
      }
      filteredResults.forEach(function (filename) {
        sharp(filename).metadata(function (err, result) {  
          if (err) {
            return console.trace('imageScan error: %s filename: %s', err, filename);
          }

          result.filename = filename;
          var the_metadata = findMetadata(path.basename(filename));
          result.id = parseInt(the_metadata.id);
          result.post_url = the_metadata.post_url;
          result.author = the_metadata.author;
          result.author_url = the_metadata.author_url;
          newimages.push(result);

          if (!--left) {
            console.log('Done scanning');
            newimages.sort(function (a,b) { 
              return a.id > b.id; 
            });
            images = newimages;
            fs.writeFile(config.image_store_path, JSON.stringify(newimages), 'utf8', function (err) {
              startWebServers();
            });
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