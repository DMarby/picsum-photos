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
  var moment = require('moment')
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
        bandWidth = data.traffic ? data.traffic.total.tx : data.eth0.traffic.total.tx
      }
    })
  }

  setInterval(fetchBandwidth, 1000 * 30)
  setInterval(fetchStats, 1000)
  fetchBandwidth()
  fetchStats()

  var exited = false

  var saveToFileAndExit = function () {
    if (exited) {
      return
    } else {
      exited = true
    }

    console.log('Current stats:', stats)

    fs.writeFileSync(config.stats_path, JSON.stringify(stats), 'utf8')
    process.exit(0)
  }

  var saveToFile = function (callback) {
    fs.writeFile(config.stats_path, JSON.stringify(stats), 'utf8', function (error) {
      callback()
    })
  }

  process.on('exit', saveToFileAndExit)
  process.on('SIGINT', saveToFileAndExit)
  process.on('SIGTERM', saveToFileAndExit)
  process.on('uncaughtException', function (error) {
    console.log('Uncaught exception: ')
    console.trace(error)
    saveToFileAndExit()
  })

  var loadImages = function () {
    var newImages = []

    async.each(metadata, function (image, next) {
      if (image.deleted) {
        return setImmediate(next)
      }

      var existingImage = imageExists(image)

      if (existingImage) {
        existingImage.post_url = image.post_url
        existingImage.author = image.author
        existingImage.author_url = image.author_url
        newImages.push(existingImage)

        return setImmediate(next)
      }

      var filename = path.resolve(config.folder_path, image.filename)

      console.log('Getting info for new image %s', filename)

      sharp(filename).metadata(function (error, result) {
        if (error) {
          console.trace('imageScan error: %s filename: %s', error, filename)
          return setImmediate(next)
        }

        result.filename = filename
        result.id = image.id
        result.post_url = image.post_url
        result.author = image.author
        result.author_url = image.author_url
        newImages.push(result)

        next()
      })
    }, function (error) {
      writeImagesToFile(newImages)
    })
  }

  var imageExists = function (image) {
    for (var i in images) {
      if (images[i].id === image.id) {
        return images[i]
      }
    }

    return false
  }

  var writeImagesToFile = function (newImages) {
    newImages.sort(function (a, b) {
      return a.id - b.id
    })

    images = newImages

    fs.writeFile(config.image_store_path, JSON.stringify(newImages), 'utf8', function (error) {
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

    var triggerSaveToFile = function () {
      saveToFile(function () {
        setImmediate(setTimeout, triggerSaveToFile, 1000 * 5)
      })
    }

    setTimeout(triggerSaveToFile, 1000 * 5)
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
  var config = require('./config')()
  require('./server')(function (callback) {
    callback.listen(process.env.PORT || config.port, '0.0.0.0')
  })
}
