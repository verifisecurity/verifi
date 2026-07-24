// Minimal app source: imports lodash, but never imports minimist. The usage
// signal should mark lodash as imported and minimist as not imported.
const _ = require('lodash');

module.exports = function padName(name) {
  return _.padStart(name, 10);
};
