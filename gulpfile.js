const del = require('del')
const gulp = require('gulp')
const purgecss = require('gulp-purgecss')

var watch = false

gulp.task('clean', () => {
  return del(['static'])
})

gulp.task('assets', () => {
  return gulp
          .src(['src/**/*', '!src/**/*.min.css'])
          .pipe(gulp.dest('static'))
})

// Extractor from https://tailwindcss.com/docs/controlling-file-size/#removing-unused-css-with-purgecss
class TailwindExtractor {
  static extract(content) {
    return content.match(/[A-z0-9-:\/]+/g)
  }
}

gulp.task('purgecss', function() {
  var pipe = gulp.src('src/**/*.min.css')

  if (!watch) {
    pipe = pipe.pipe(
      purgecss({
        content: ['src/*.html'],
        extractors: [
          {
            extractor: TailwindExtractor,
            extensions: ['html']
          }
        ]
      })
    )
  }

  return pipe.pipe(gulp.dest('static/'))
})

gulp.task('watch', () => {
  watch = true
  gulp.watch(['src/**/*'], gulp.parallel('assets', 'purgecss'))
})

gulp.task('default', gulp.series('clean', gulp.parallel('assets', 'purgecss')))
