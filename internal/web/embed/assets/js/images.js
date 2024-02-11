window.addEventListener('DOMContentLoaded', function () {
  document.getElementById('prev').addEventListener('click', handleNavigationButton)
  document.getElementById('next').addEventListener('click', handleNavigationButton)

  loadPageFromhHash()
})

window.addEventListener('hashchange', loadPageFromhHash)

function loadPageFromhHash() {
  var page = 1

  if (window.location.hash) {
    page = window.location.hash.substring(1)
  }

  loadPage(page)
}

function handleNavigationButton (event) {
  event.preventDefault()

  var page = event.target.getAttribute('data-page')
  if (!page) {
    return
  }

  // Set the hash, this will change the page by triggering the 'hashchange' event
  window.location.hash = page
  window.scrollTo({ top: 0 })
}

function loadPage (page) {
  var xhr = new XMLHttpRequest()
  xhr.open('GET', '/v2/list?page=' + page, true)
  xhr.onreadystatechange = function () {
    if (xhr.readyState === 4 && xhr.status === 200) {
      var images = JSON.parse(xhr.responseText)

      var container = document.querySelector('.image-list')
      container.innerHTML = ''

      for (var image of images) {
        var template = document.querySelector('#image-template')
        var clone = document.importNode(template.content, true)

        // Image
        clone.querySelector('img').src = '/id/' + image.id + '/367/267'
        clone.querySelector('.download-url').href = image.download_url

        // Author
        clone.querySelector('.author').innerHTML = image.author
        clone.querySelector('.author-url').href = image.url

        // Image id indicator
        clone.querySelector('.image-id').innerHTML = '#' + image.id
        clone.querySelector('.image-id').href = image.download_url

        container.appendChild(clone)
      }

      var linkHeaders = parseLinkHeader(xhr.getResponseHeader('Link'))

      updateButton('prev', linkHeaders.prev)
      updateButton('next', linkHeaders.next)
    }
  }

  xhr.send()
}

function updateButton (id, page_url) {
  var button = document.getElementById(id)

  if (page_url) {
    var url = new URL(page_url)
    var urlParams = new URLSearchParams(url.search)
    button.setAttribute('data-page', urlParams.get('page'))
    button.classList.add('hover:text-white', 'hover:bg-gray-500')
    button.classList.remove('cursor-not-allowed', 'opacity-50')
  } else {
    button.removeAttribute('data-page')
    button.classList.add('cursor-not-allowed', 'opacity-50')
    button.classList.remove('hover:text-white', 'hover:bg-gray-500')
  }
}

// From https://gist.github.com/deiu/9335803
function parseLinkHeader (header) {
  var linkexp = /<[^>]*>\s*(\s*;\s*[^\(\)<>@,;:"\/\[\]\?={} \t]+=(([^\(\)<>@,;:"\/\[\]\?={} \t]+)|("[^"]*")))*(,|$)/g
  var paramexp = /[^\(\)<>@,;:"\/\[\]\?={} \t]+=(([^\(\)<>@,;:"\/\[\]\?={} \t]+)|("[^"]*"))/g

  var matches = header.match(linkexp)
  var rels = {}

  for (var i = 0; i < matches.length; i++) {
      var split = matches[i].split('>')
      var href = split[0].substring(1)
      var ps = split[1]

      var s = ps.match(paramexp)
      for (var j = 0; j < s.length; j++) {
          var p = s[j]
          var paramsplit = p.split('=')
          var name = paramsplit[0]
          var rel = paramsplit[1].replace(/["']/g, '')
          rels[rel] = href
      }
  }

  return rels
}
