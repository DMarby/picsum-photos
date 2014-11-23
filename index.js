var cluster = require('cluster')

if (cluster.isMaster) {
  var async = require('async')
  var config = require('./config')()
  var fs = require('fs')
  var path = require('path')
  var sharp = require('sharp')
  var io = require('socket.io')(config.stats_port)
  var vnstat = require('vnstat-dumpdb')
  var metadata = require(config.metadata_path)
  console.log('Config:')
  console.log(config)

  try {
    var stats = require(config.stats_path)
  } catch (error) {
    var stats = { count: 0 }
  }

  try {
    var images = require(config.image_store_path)
  } catch (e) {
    var images = []
  }

  fs.mkdir(config.cache_folder_path, function (error) {})

  var publicStats = {}
  var bandWidth = 0

  io.on('connection', function (socket) {
    socket.emit('stats', publicStats)
  })

  var fetchStats = function () {
    publicStats = {
      count: stats.count,
      bandWidth: bandWidth,
      images: images.length
    }

    io.emit('stats', publicStats)
  }

  var fetchBandwidth = function () {
    vnstat.dumpdb(function (error, data) {
      if (error) {
        console.log('Couldn\'t fetch bandwidth: ' + error)
      } else {
        bandWidth = data.traffic.total.tx
      }
    })
  }

  setInterval(fetchBandwidth, 1000 * 30)
  setInterval(fetchStats, 1000)
  fetchBandwidth()
  fetchStats()

  var cleanupAndExit = function () {
    cleanup()
    process.exit()
  }

  var cleanup = function () {
    saveStatsToFile()
  }

  var saveStatsToFile = function () {
    fs.writeFileSync(config.stats_path, JSON.stringify(stats), 'utf8')
  }

  process.on('exit', cleanup)
  process.on('SIGINT', cleanupAndExit)
  process.on('SIGTERM', cleanupAndExit)
  process.on('uncaughtException', function (error) {
    console.log('Uncaught exception: ')
    console.trace(error)
    cleanupAndExit()
  })

  var loadImages = function () {
    var newimages = []

    images.forEach(function (image) {
      var the_metadata = findMetadata(path.basename(image.filename))
      image.post_url = the_metadata.post_url
      image.author = the_metadata.author
      image.author_url = the_metadata.author_url
      newimages.push(image) 
    })

    scanDirectory(config.folder_path, function (error, results) {
      if (error) throw error
      
      var filteredResults = results.filter(function (filename) {
        return images.filter(function (image) { return image.filename == filename; }) == 0
      })

      if (filteredResults.length <= 0) {
        console.log('Done scanning, no new images')
        writeImagesToFile(newimages)
        return
      }

      async.each(filteredResults, function (filename, next) {
        sharp(filename).metadata(function (error, result) {  
          if (error) {
            console.trace('imageScan error: %s filename: %s', error, filename)
            return next(error)
          }

          result.filename = filename
          var the_metadata = findMetadata(path.basename(filename))
          result.id = parseInt(the_metadata.id)
          result.post_url = the_metadata.post_url
          result.author = the_metadata.author
          result.author_url = the_metadata.author_url
          newimages.push(result)

          next()
        })
      }, function (error) {
        console.log('Done scanning')
        writeImagesToFile(newimages)
      })
    })
  }

  var findMetadata = function (filename) {
    for (var i in metadata) {
      if (metadata[i].filename == filename) {
        return metadata[i]
      }
    }
  }

  var scanDirectory = function (dir, done) {
    var results = []
    fs.readdir(dir, function (error, list) {
      if (error) {
        return done(error)
      }
      
      if (!list.length) {
        return done (null, results)
      }

      async.each(list, function (file, next) {
        file = path.resolve(dir, file)
        fs.stat(file, function (error, stat) {
          if (stat && stat.isFile() && !endsWith(file, '.DS_Store') && !endsWith(file, 'metadata.json')) {
            results.push(file)
          }
          next(error)          
        })
      }, function (error) {
        done(error, results)
      })
    })
  }

  var endsWith = function (str, suffix) {
    return str.indexOf(suffix, str.length - suffix.length) !== -1
  }

  var writeImagesToFile = function (newimages) {
    newimages.sort(function (a, b) { 
      return a.id - b.id
    })

    images = newimages

    fs.writeFile(config.image_store_path, JSON.stringify(newimages), 'utf8', function (error) {
      startWebServers()
    })
  }

  var startWebServers = function () {
    var cpuCount = require('os').cpus().length - 1

    if (cpuCount < 2) {
      cpuCount = 2
    }

    for (var i = 0, il=cpuCount; i < il; i++) {
      startWorker()
    }

    cluster.on('exit', function (worker) {
      console.log('Worker ' + worker.id + ' died')
      startWorker()
    })

    setInterval(function () {
      saveStatsToFile()
    }, 5000)
  }

  var startWorker = function () {
    var worker = cluster.fork()
    worker.on('message', handleWorkerMessage)
    console.log('Worker ' + worker.id + ' started')
  }

  var handleWorkerMessage = function (msg) {
    stats.count++
  }

  loadImages()
} else {
  require('./server')(function (callback) {
    callback.listen(process.env.PORT || 5000, 'localhost')
  })
}