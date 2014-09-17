var sharp = require('sharp');
var path = require('path');
var config = require('./config')();
var fs = require('fs');

var imageProcessor = require('./imageProcessor')(sharp, path, config, fs);
var images = require(config.image_store_path);

for (var i in images) {
  var width = 458;
  var height = 354;
  var filePath = images[i].filename;
  var blur = false;
  imageProcessor.getProcessedImage(width, height, null, false, false, filePath, true, function (err, imagePath) {
    if (err) {
      console.log('filePath: ' + filePath);
      console.log('imagePath: ' + imagePath);
      console.log('error: ' + err);
    }
  })
}