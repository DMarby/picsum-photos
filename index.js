var cluster = require('cluster');
if (cluster.isMaster) {

  var cpuCount = require('os').cpus().length - 1;
  if (cpuCount < 2) {
    cpuCount = 2;
  }

  for (var i = 0, il=cpuCount; i < il; i++) {
    var worker = cluster.fork();
    console.log('Worker ' + worker.id + ' started');
  }

  cluster.on('exit', function (worker) {
    console.log('Worker ' + worker.id + ' died');
    var worker = cluster.fork();
    console.log('Worker ' + worker.id + ' started');
  });

} else {
  require('./server')(function (callback) {
    callback.listen(process.env.PORT || 5000, 'localhost');
  });
}