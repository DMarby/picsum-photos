const del = require('del')
const gulp = require('gulp')
const purgecss = require('gulp-purgecss')

var watch = false

gulp.task('clean', () => {
  return del(['dist'])
})

gulp.task('assets', () => {
  return gulp
          .src(['web/**/*', '!web/**/*.min.css'])
          .pipe(gulp.dest('dist'))
})

// Extractor from https://tailwindcss.com/docs/controlling-file-size/#removing-unused-css-with-purgecss
class TailwindExtractor {
  static extract(content) {
    return content.match(/[A-z0-9-:\/]+/g)
  }
}

gulp.task('purgecss', function() {
  var pipe = gulp.src('web/**/*.min.css')

  if (!watch) {
    pipe = pipe.pipe(
      purgecss({
        content: ['web/*.html'],
        extractors: [
          {
            extractor: TailwindExtractor,
            extensions: ['html']
          }
        ]
      })
    )
  }

  return pipe.pipe(gulp.dest('dist/'))
})

gulp.task('watch', () => {
  watch = true
  gulp.watch(['web/**/*'], gulp.parallel('assets', 'purgecss'))
})

gulp.task('default', gulp.series('clean', gulp.parallel('assets', 'purgecss')))
