var http = require('http');

var options = {
  host: 'localhost',
  path: '/append',
  port: '8080',
  method: 'POST'
};

callback = function(response) {
  var str = ''
  response.on('data', function (chunk) {
    str += chunk;
  });

  response.on('end', function () {
    console.log(str);
  });
}

var req = http.request(options, callback);
req.write({"somedata": "crap"});
req.end();