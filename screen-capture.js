var system = require('system');
var url = system.args[1];
var fileName = system.args[2];
var format = system.args[3];

var page = require('webpage').create();
page.viewportSize = { width: 1920, height: 1080 };
page.resourceTimeout = 20000;
page.open(url, function() {
  page.render(fileName,{'format':format,quality:'20'});
  phantom.exit();
});
