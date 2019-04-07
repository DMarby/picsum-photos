window.addEventListener('DOMContentLoaded', function () {
  setupRandomImage('header')
  setupRandomImage('footer')
})

function setupRandomImage (element) {
  var element = document.querySelector(element)
  var src = '/' + element.offsetWidth + '/' + element.offsetHeight + '?random&element=' + element.tagName
  element.style['background-image'] = 'url(' + src + ')'
}
