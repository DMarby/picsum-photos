$(document).ready(function() {
  var setupRandomImage = function (element) {
    var $element = $(element);
    var width = $element.width();
    var height = $element.outerHeight();
    var src = 'https://unsplash.it/' + width + '/' + height + '?random&element=' + $element.prop('tagName');

    $element.css('background-image', 'url('+src+')');
  }

  setupRandomImage('.intro-header');
  setupRandomImage('footer');

  var countToNumber = function (element, number, suffix, duration) {
    $({count: parseInt(element.text().split("+")[0].replace(/\,/g, ''))}).animate({count: number}, {
      duration: duration ? duration : 1000,
      easing: 'swing',
      step: function (now) {
        element.text((Math.floor(now) + suffix).replace(/(\d)(?=(\d\d\d)+(?!\d))/g, "$1,"));
      },
      complete: function () {
        countingFromZero = false;
      }
    });
  }

  var first = true;
  var countingFromZero = true;

  var socket = io.connect('http://45.55.96.43:3000');
  socket.on('stats', function (data) {
    if (!first && countingFromZero) {
      return;
    }
    countToNumber($('.imageCounter'), data.count, '', first ? 2000 : 800);
    countToNumber($('.bandwidthCounter'), Math.round((data.bandWidth / 1024 / 1024)), '+ GB', first ? 1000 : 800);
    countToNumber($('.images'), data.images, '', first ? 1000 : 800);
    if (first) {
      first = false;
    }
  });
});