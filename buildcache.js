var sharp = require('sharp');
var path = require('path');
var config = require('./config')();
var fs = require('fs');

var imageProcessor = require('./imageProcessor')(sharp, path, config, fs);
var images = require(config.image_store_path);

var index = process.argv[2] || 0;
console.log('Start: %s', index);

var nextImage = function (the_index, callback) { 
  var width = 458;
  var height = 354;
  var filePath = images[the_index].filename;
  var blur = false;
  imageProcessor.getProcessedImage(width, height, null, false, false, filePath, false, function (err, imagePath) {
    if (err) {
      console.log('filePath: ' + filePath);
      console.log('imagePath: ' + imagePath);
      console.log('error: ' + err);
    }
    console.log('%s done', the_index);
    callback();
  })
}

var imageLinksCallback = function () {
  index++;
  if (index >= images.length) {
    console.log('Done!');
    return;
  }
  nextImage(index, imageLinksCallback);
}

nextImage(index, imageLinksCallback);